package Cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/pred695/golang-blockchain/Blockchain"
	"github.com/pred695/golang-blockchain/Wallet"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount of coins")
	fmt.Println(" createwallet - Creates a new wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) GetBalance(address string) {

	if(!Wallet.ValidateAddress(address)){
		log.Panic("Address is not valid")
	}

	var chain *Blockchain.Blockchain = Blockchain.ContinueBlockchain(address)
	var UTXOSet Blockchain.UTXOSet = Blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	var balance int = 0
	var pubKeyHash []byte = Wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	var UTXOs []Blockchain.TxOutput = UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) CreateBlockChain(address string) {

	if(!Wallet.ValidateAddress(address)){
		log.Panic("Address is not valid")
	}

	chain := Blockchain.InitBlockChain(address)
	defer chain.Database.Close()
	var UTXOSet Blockchain.UTXOSet = Blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()
	fmt.Println("Finished!")
}

func (cli *CommandLine) PrintChain() {
	chain := Blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := Blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		for _, tx := range block.Transactions{
			fmt.Println(tx.To_String())
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) Send(from, to string, amount int) {

	if(!Wallet.ValidateAddress(from)){
		log.Panic("Sender's Address is not valid")
	}
	if(!Wallet.ValidateAddress(to)){
		log.Panic("Receiver's Address is not valid")
	}

	chain := Blockchain.ContinueBlockchain(from)
	var UTXO Blockchain.UTXOSet = Blockchain.UTXOSet{Blockchain: chain}
	//Since the user who is sending the coins is the one who is creating the transaction(block as of now), we need to consider this transaction as a coinbase transaction
	var cbTx Blockchain.Transaction = *Blockchain.CoinbaseTx(from, "")
	defer chain.Database.Close()

	tx := UTXO.NewTransaction(from, to, amount)
	var block *Blockchain.Block = chain.AddBlock([]*Blockchain.Transaction{&cbTx, tx})
	UTXO.Update(block)
	fmt.Println("Success!")
}

func (cli *CommandLine) CreateWallet() {
	wallets, _ := Wallet.CreateWallets() //loads the wallets from the file
	address := wallets.AddWallet() 	 //adds a new wallet
	wallets.SaveFile()

	fmt.Printf("New address: %s\n", address)

}

func (cli *CommandLine) ListAddresses() {
	wallets, err := Wallet.CreateWallets() //loads the wallets from the file
	Blockchain.Handle(err)
	addresses := wallets.GetAllAddresses()
	fmt.Println(len(addresses))
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) ReindexUTXO() {
	var chain *Blockchain.Blockchain = Blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	var UTXOSet Blockchain.UTXOSet = Blockchain.UTXOSet{Blockchain: chain}

	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		Handle(err)
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		Handle(err)
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		Handle(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		Handle(err)
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		Handle(err)
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		Handle(err)
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		Blockchain.Handle(err)	
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.Send(*sendFrom, *sendTo, *sendAmount)
	}
	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.ListAddresses()
	}
	if reindexUTXOCmd.Parsed() {
		cli.ReindexUTXO()
	}
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}