package Blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// Method for creating the block(Retuns a block pointer)
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0} //Create a block with the data and the previous hash
	proof := NewProof(block)                             //Derive the hash of the block
	nonce, hash := proof.Run()                           //Run the proof of work algorithm
	block.Hash = hash[:]                                 //Set the hash of the block to the hash derived from the proof of work algorithm
	block.Nonce = nonce
	return block //Return the block
}

// Creates the genesis block -- the first block in the blockchain
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
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
