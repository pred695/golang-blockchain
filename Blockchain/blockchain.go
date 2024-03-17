package Blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./temp/blocks"
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

func InitBlockChain() *Blockchain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath            //stores the database in the temp folder
	opts.ValueDir = dbPath       //stores the value
	db, err := badger.Open(opts) //returns a tuple containing (pointer to the db, error).
	Handle(err)
	//Database opened, lets check if there exists a blockchain in the database or not.
	//Update function allows to read and write transactions our the database.
	err = db.Update(func(txn *badger.Txn) error {
		//"lh" is the key in the key value pair of lasthash part of the database.
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			//we don't have a database or a blockchain existing, create a genesis block
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved") //Creating the genesis block.
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash //setting the lastHash to the hash of the genesis block.
			return err
		} else {
			//we already have a blockchain, so we need to get the last hash.
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lastHash, err = item.ValueCopy([]byte{}) //get the value associated with the item and store it in lastHash.
			return err
		}
	})
	Handle(err) //handling the returned error from the method above.
	blockchain := Blockchain{lastHash, db}
	return &blockchain
}

func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	//View function allows to read transactions from the database.
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh")) //get the current last hash
		Handle(err)
		lastHash, err = item.ValueCopy([]byte{})
		return err
	})
	Handle(err)
	newBlock := CreateBlock(data, lastHash) //create a new block with the data and the last hash
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

//Iterating from the newest to the genesis block(reverse iteration)
func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error{
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.ValueCopy([]byte{})	//getting the block in the encoded form of bytes, need to deserialize into block.
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}


func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
