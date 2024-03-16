package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"rsc.io/quote"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

type Blockchain struct {
	blocks []*Block
}

// Method for deriving the block hash(The argument is pointer to the Block variable and returns nothing)
func (b *Block) DeriveHash() {
	//Join the data and the previous hash with a new empty byte slice  into a double dimensional byte slice with the separator being an empty byte slice
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info) //Hash the info
	b.Hash = hash[:]            //Assign the hash to the block in the form of a slice
}

// Method for creating the block(Retuns a block pointer)
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash} //Create a block with the data and the previous hash
	block.DeriveHash()                                //Derive the hash of the block
	return block                                      //Return the block
}

func (chain *Blockchain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}

// Creates the genesis block -- the first block in the blockchain
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func InitBlockChain() *Blockchain {
	return &Blockchain{[]*Block{Genesis()}}
}

func main() {
	fmt.Println(quote.Go())
	chain := InitBlockChain()
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
	}
}
