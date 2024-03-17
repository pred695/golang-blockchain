package main

import (
	"fmt"

	"github.com/pred695/golang-blockchain/Blockchain"
	"github.com/pred695/golang-blockchain/cli"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Go())
	chain := Blockchain.InitBlockChain()
	defer chain.Database.Close()
	Cli := cli.CommandLine{C_blockchain: chain}
	Cli.Run()
}
