
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
"sender": "66ad49fc75114203-8db2c886c13339ea",
"recipient": "98c438146f674768a1bb0e225a75311f",
"amount": 50
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
