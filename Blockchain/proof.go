package Blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

// Difficulty is the number of leading zeros that must be present in the hash
/*
Step 1: Take Data from the block
Step 2: Take the previous hash from the block
Step 3: Take the nonce from the block
Step 4: Take the difficulty from the block
Step 5: Join the data, previous hash, nonce and difficulty into a single byte slice
Step 6: Hash the byte slice
Step 7: Return the hash
*/

const Difficulty = 18
const MaxNonce = math.MaxInt64

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	proof := &ProofOfWork{block, target}
	return proof
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}

func (proof *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			proof.Block.HashTransactions(),
			proof.Block.PrevHash,
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data
}

func (proof *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nonce := 0
	for nonce < MaxNonce {
		data := proof.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(proof.Target) == -1 {
			break //our hash is less than the target we are looking for, success.
		} else {
			nonce++
		}
		fmt.Println()
	}
	return nonce, hash[:]
}
//Computational part of the algorithm is relatively expensive, validation part is pretty simple.
//Changing a particular block requires the "expensive" part of the algorithm to be run again, hence the blockchain is secure(tamper-proof).
func (proof *ProofOfWork) Validate() bool {
	var intHash big.Int
	data := proof.InitData(proof.Block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	result := intHash.Cmp(proof.Target)
	return result == -1
}
