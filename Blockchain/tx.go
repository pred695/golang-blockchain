package Blockchain

import (
	"bytes"

	"github.com/pred695/golang-blockchain/Wallet"
)

type TxOutput struct {
	Value  int
	PubKeyHash []byte
}

type TxInput struct {
	ID        []byte //references to the previous output that led to the input
	OutputIdx int    //index of the referenced output which is spent in the transaction
	Signature []byte //signature or the sender's public key
	PubKey    []byte //unhashed
}

func NewTxOutput(value int, address string) *TxOutput {
	var tx_output TxOutput = TxOutput{Value: value, PubKeyHash: nil}

	tx_output.Lock([]byte(address))

	return &tx_output
}

func (inputTx *TxInput) UsesKey(pubKeyHash []byte) bool {
	var lockingHash []byte = Wallet.CreatePubKeyHash(inputTx.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0	//the input contains the public key that can unlock the output
}

func (outputTx *TxOutput) Lock(address []byte) {
	var pubKeyHash []byte = Wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1:len(pubKeyHash) - Wallet.ChecksumLength]	//taking the bytes between version and checksum
	outputTx.PubKeyHash = pubKeyHash	//the output is locked with the public key hash, to be unlocked by the input's  public key
}

func (output *TxOutput) IsLockedWithKey (pubKeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubKeyHash) == 0
}