package Blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./temp/blocks"
	dbFile      = "./temp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type Blockchain struct {
	LastHash []byte     //The hash of the previous block
	Database *badger.DB //pointer to the database.
}

// to iterate over the blockchain
type BlockchainIterator struct {
	CurrentHash []byte //similar to the last hash field
	Database    *badger.DB
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func InitBlockChain(address string) *Blockchain {
	var lastHash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		//Create a coinbase transaction, the first transaction in the blockchain
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")

		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)

		err = txn.Set([]byte("lh"), genesis.Hash) //setting the value of genesis.Hash = []byte("lh"), the last hash.
		lastHash = genesis.Hash
		return err
	})
	Handle(err)

	blockchain := Blockchain{LastHash: lastHash, Database: db}
	return &blockchain
}

func ContinueBlockchain(address string) *Blockchain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		//find the lastHash
		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, err = item.ValueCopy([]byte{}) //get the value associated with the item and store it in lastHash.
		return err
	})
	Handle(err)

	blockchain := Blockchain{LastHash: lastHash, Database: db}
	return &blockchain
}

func (chain *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte
	//View function allows to read transactions from the database.
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh")) //get the current last hash
		Handle(err)
		lastHash, err = item.ValueCopy([]byte{})
		return err
	})

	Handle(err)
	newBlock := CreateBlock(transactions, lastHash) //create a new block with the data and the last hash
	//new block created, perform read and write operations --> use Update function.
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize()) //set the hash of the new block to the serialized version of the new block.
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash) //set the last hash to the hash of the new block.
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

// Iterating from the newest to the genesis block(reverse iteration)
func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{CurrentHash: chain.LastHash, Database: chain.Database}
	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.ValueCopy([]byte{}) //getting the block in the encoded form of bytes, need to deserialize into block.
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}

func (chain *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTransactions []Transaction
	spentTXOs := make(map[string][]int) //map of string(key) and slice of int(value)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, output := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs //if the output is already spent, continue to the next output
						}
					}
				}
				if output.IsLockedWithKey(pubKeyHash) { //output's public key is the same as the address
					unspentTransactions = append(unspentTransactions, *tx)
				}
			}
			if tx.Is_Coinbase() == false { //skipping coinbase transactions as it has no inputs.
				for _, input := range tx.Inputs { //finding other outputs which are referenced using inputs.
					if input.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(input.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], input.OutputIdx)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break //reached the genesis block
		}
	}
	return unspentTransactions
}

func (chain *Blockchain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	var unspentTx []Transaction = chain.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTx {
		for _, output := range tx.Outputs {
			if output.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, output)
			}
		}
	}
	return UTXOs
}

func (chain *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	var unspentOutputs map[string][]int
	unspentOutputs = make(map[string][]int)
	var unspentTxs = chain.FindUnspentTransactions(pubKeyHash)
	var accumulated int = 0
	//spendable outputs are always the outputs of the unspent transactions
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, output := range tx.Outputs {
			if output.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += output.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func (chain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	var iter *BlockchainIterator = chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
			if len(block.PrevHash) == 0 {
				break
			}
		}
		return Transaction{}, errors.New("Transaction does not exist")
	}
}

func (chain *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	var prevTxs map[string]Transaction = make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTX, err := chain.FindTransaction(input.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTxs)
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.Is_Coinbase() {
		return true
	}

	var prevTxs map[string]Transaction = make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTX, err := chain.FindTransaction(input.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTxs)
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
