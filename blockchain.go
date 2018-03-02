package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Blockchain ...
type Blockchain struct {
	Chain               []Block       `json:"chain"`
	CurrentTransactions []Transaction `json:"current_transactions"`
	Nodes               []string      `json:"nodes"`
}

// Block ...
type Block struct {
	Index        uint64        `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        uint64        `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

// Transaction ...
type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    uint64 `json:"amount"`
}

// NewBlockchain ..
func NewBlockchain() *Blockchain {
	bc := new(Blockchain)
	bc.CurrentTransactions = nil
	bc.Chain = nil
	bc.Nodes = nil
	bc.newBlock(uint64(100), "1")
	return bc
}

func (bc *Blockchain) registerNode(address string) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		log.Println(err)
		return
	}
	host := parsedURL.Host
	if host == "" {
		return
	}
	for _, node := range bc.Nodes {
		if node == host {
			return
		}
	}
	bc.Nodes = append(bc.Nodes, host)
}

func (bc *Blockchain) validChain(chain []Block) bool {
	lastBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		block := chain[currentIndex]
		if block.PreviousHash != hash(lastBlock) {
			return false
		}
		if !validProof(lastBlock.Proof, block.Proof, lastBlock.PreviousHash) {
			return false
		}
		lastBlock = block
		currentIndex++
	}
	return true
}

type chainResponse struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

func (bc *Blockchain) resolveConflicts() bool {
	var newChain []Block
	neighbours := bc.Nodes
	maxLength := len(bc.Chain)
	for _, node := range neighbours {
		url := fmt.Sprintf("http://%s.chain", node)
		res, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return false
		}
		defer res.Body.Close()
		byteArr, _ := ioutil.ReadAll(res.Body)
		var response chainResponse
		if err = json.Unmarshal(byteArr, &response); err != nil {
			log.Println(err)
			return false
		}
		length := response.Length
		chain := response.Chain
		if length > maxLength && bc.validChain(chain) {
			maxLength = length
			newChain = chain
		}
	}
	if len(newChain) > 0 {
		bc.Chain = newChain
		return true
	}
	return false

}

func (bc *Blockchain) newBlock(proof uint64, previousHash string) Block {
	block := Block{
		Index:        uint64(len(bc.Chain) + 1),
		Timestamp:    time.Now(),
		Transactions: bc.CurrentTransactions,
		Proof:        proof,
		PreviousHash: previousHash,
	}
	bc.CurrentTransactions = nil
	bc.Chain = append(bc.Chain, block)
	return block
}

func (bc *Blockchain) newTransaction(sender, recipient string, amount uint64) uint64 {
	transaction := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	bc.CurrentTransactions = append(bc.CurrentTransactions, transaction)
	return bc.lastBlock().Index + 1
}

func (bc *Blockchain) proofOfWork(lastBlock Block) uint64 {
	lastProof := lastBlock.Proof
	lastHash := hash(lastBlock)
	proof := uint64(0)
	for validProof(lastProof, proof, lastHash) == false {
		proof++
	}
	return proof
}

func (bc *Blockchain) lastBlock() Block {
	return bc.Chain[len(bc.Chain)-1]
}

func hash(block Block) string {
	jsonBytes, _ := json.Marshal(block)
	hashBytes := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hashBytes[:])
}

func validProof(lastProof, proof uint64, lastHash string) bool {
	guess := fmt.Sprintf("%x%x%x", lastProof, proof, lastHash)
	guessBytes := sha256.Sum256([]byte(guess))
	guessHash := hex.EncodeToString(guessBytes[:])
	return guessHash[:4] == "0000"
}

// Mine ...
type Mine struct {
	Message      string        `json:"message"`
	Index        uint64        `json:"index"`
	Transactions []Transaction `json:"transactions"`
	Proof        uint64        `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

func mineHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
	lastBlock := bc.lastBlock()
	proof := bc.proofOfWork(lastBlock)

	bc.newTransaction("0", nodeIdentifier, 1)
	previousHash := hash(lastBlock)
	block := bc.newBlock(proof, previousHash)

	mine := Mine{
		Message:      "New Block Forged",
		Index:        block.Index,
		Transactions: block.Transactions,
		Proof:        block.Proof,
		PreviousHash: block.PreviousHash,
	}
	res, err := json.Marshal(mine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func newTransactionHandler(w http.ResponseWriter, r *http.Request) {
	type jsonBody struct {
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
		Amount    uint64 `json:"amount"`
	}
	decoder := json.NewDecoder(r.Body)
	var b jsonBody
	if err := decoder.Decode(&b); err != nil {
		log.Fatal(err)
	}

	index := bc.newTransaction(b.Sender, b.Recipient, b.Amount)
	type response struct {
		Message string `json:"message"`
	}
	resNewTransaction := response{
		Message: "Transaction will be added to Block" + strconv.FormatUint(index, 10),
	}
	res, err := json.Marshal(resNewTransaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func chainHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chain  []Block `json:"chain"`
		Length int     `json:"length"`
	}
	resChain := response{
		Chain:  bc.Chain,
		Length: len(bc.Chain),
	}
	res, err := json.Marshal(resChain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func nodesRegisterHandler(w http.ResponseWriter, r *http.Request) {
	type jsonBody struct {
		Nodes []string `json:"nodes"`
	}
	decoder := json.NewDecoder(r.Body)
	var b jsonBody
	if err := decoder.Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, node := range b.Nodes {
		bc.registerNode(node)
	}
	type response struct {
		Message    string   `json:"message"`
		TotalNodes []string `json:"total_nodes"`
	}
	var resNodesRegister response
	resNodesRegister = response{
		Message:    "New nodes have been added",
		TotalNodes: bc.Nodes,
	}
	res, err := json.Marshal(resNodesRegister)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func nodesResolveHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Message string  `json:"message"`
		Chain   []Block `json:"chain"`
	}
	var resNodesResolve response
	replaced := bc.resolveConflicts()
	if replaced {
		resNodesResolve = response{
			Message: "Our chain was replaced",
			Chain:   bc.Chain,
		}
	} else {
		resNodesResolve = response{
			Message: "Our chain is authoritative",
			Chain:   bc.Chain,
		}
	}
	res, err := json.Marshal(resNodesResolve)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

var bc *Blockchain
var nodeIdentifier string

func main() {
	args := os.Args
	port := "5000"
	if len(args) > 2 && args[1] == "-p" {
		port = args[2]
	}
	bc = NewBlockchain()
	out, _ := exec.Command("uuidgen").Output()
	nodeIdentifier = strings.Replace(string(out), "-", "", -1)
	http.HandleFunc("/mine", mineHandler)
	http.HandleFunc("/transactions/new", newTransactionHandler)
	http.HandleFunc("/chain", chainHandler)
	http.HandleFunc("/nodes/register", nodesRegisterHandler)
	http.HandleFunc("/nodes/resolve", nodesResolveHandler)
	log.Printf("Server listening on localhost:%s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
