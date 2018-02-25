
# Blockchain

Golang port of dvf/blockchain

### Run the server
```
$ git clone git@github.com:dongri/blockchain.git
$ cd blockchain
$ go run blockchain.go
Server listening on localhost:5000.

```

### Chain
```
$ curl localhost:5000/chain
```

### Mining
```
$ curl localhost:5000/mine
```

### New Transaction
```
$ curl -X POST -H "Content-Type: application/json" -d '{
"sender": "d4ee26eee15148ee92c6cd394edd974e",
"recipient": "someone-other-address",
"amount": 5
}' localhost:5000/transactions/new
```

### Add node
```
$ curl -X POST -H "Content-Type: application/json" -d '{
"nodes": ["http://localhost:5001","http://localhost:5002"]
}' localhost:5000/nodes/register
```

### Resolve node
```
$ curl localhost:5000/nodes/resolve 
```
