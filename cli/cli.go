package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/pred695/golang-blockchain/Blockchain"
)

//Naming the variables and the methods/functions/closures with capital letters to make them public.
type CommandLine struct {
	C_blockchain *Blockchain.Blockchain
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("add -block <BLOCK_DATA> - add a block to the blockchain")
	fmt.Println("print - prints the blocks in the chain")
}
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage() //again print the usage interface, invalid number of arguments.
		runtime.Goexit()
	}
}
func (cli *CommandLine) AddBlock(data string) {
	cli.C_blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}
func (cli *CommandLine) PrintChain() {
	iter := cli.C_blockchain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		//Proof of work
		proof := Blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(proof.Validate()))
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		} //reached the genesis block
	}
}
func (cli *CommandLine) Run() {
	cli.ValidateArgs()
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")
	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		Blockchain.Handle(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		Blockchain.Handle(err)

	default:
		cli.PrintUsage()
		runtime.Goexit()
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.AddBlock(*addBlockData)
	} else if printChainCmd.Parsed() {
		cli.PrintChain()
	}
}
