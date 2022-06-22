package main

import (
	"github.com/auturnn/peer-base-nodes/blockchain"
	"github.com/auturnn/peer-base-nodes/db"
	"github.com/auturnn/peer-base-nodes/rest"
)

const port int = 8080

func main() {
	defer db.Close()
	db.InitDB(port)
	blockchain.Mempool()
	rest.Start(port)
}
