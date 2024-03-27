package Blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// Method for creating the block(Retuns a block pointer), A block can contain multiple transactions(atleast one)
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	var block Block = Block{[]byte{}, txs, prevHash, 0} //Create a block with the data and the previous hash
	var proof *ProofOfWork = NewProof(&block)           //Derive the hash of the block
	var nonce int                                       //Run the proof of work algorithm
	var hash []byte
	nonce, hash = proof.Run() //Run the proof of work algorithm
	block.Hash = hash[:]      //Set the hash of the block to the hash derived from the proof of work algorithm
	block.Nonce = nonce
	return &block //Return the block
}

// Creates the genesis block -- the first block in the blockchain
func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range block.Transactions {
		// txHashes = append(txHashes, tx.HashTransaction()) //append the hashed version of the transaction to the slice of hashes
		txHashes = append(txHashes, tx.Serialize()) //append the serialized version of the transaction to the slice of hashes
	}
	var tree *MerkleTree = NewMerkleTree(txHashes) //create a new merkle tree with the hashes of the transactions

	return tree.RootNode.Data
}

// encodes the block into a byte slice
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// decodes the byte slice into a block pointer
func Deserialize(data []byte) *Block {

	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
