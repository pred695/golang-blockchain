package Blockchain

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

type Blockchain struct {
	Blocks []*Block //Blocks with a capital B since it needs to be public
}

// Method for creating the block(Retuns a block pointer)
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0} //Create a block with the data and the previous hash
	proof := NewProof(block)                              //Derive the hash of the block
	nonce, hash := proof.Run()                           //Run the proof of work algorithm
	block.Hash = hash[:]                                 //Set the hash of the block to the hash derived from the proof of work algorithm
	block.Nonce = nonce
	return block //Return the block
}

func (chain *Blockchain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, new)
}

// Creates the genesis block -- the first block in the blockchain
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func InitBlockChain() *Blockchain {
	return &Blockchain{[]*Block{Genesis()}}
}
