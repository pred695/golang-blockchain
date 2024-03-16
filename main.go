package main

import (
	"fmt"
	"strconv"

	"github.com/pred695/golang-blockchain/Blockchain"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Go())
	chain := Blockchain.InitBlockChain()
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.Blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		//Proof of work
		proof := Blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(proof.Validate()))
		fmt.Println()
	}
}
