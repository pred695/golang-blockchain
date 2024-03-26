package Blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/pred695/golang-blockchain/Wallet"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
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

func (outputs TxOutputs) SerializeOutputs() []byte {
	var buffer bytes.Buffer
	var encoder *gob.Encoder = gob.NewEncoder(&buffer)
	var err error = encoder.Encode(outputs)
	Handle(err)
	return buffer.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	var decoder *gob.Decoder = gob.NewDecoder(bytes.NewReader(data))
	var err error = decoder.Decode(&outputs)
	Handle(err)
	return outputs
}

func (inputTx *TxInput) UsesKey(pubKeyHash []byte) bool {
	var lockingHash []byte = Wallet.CreatePubKeyHash(inputTx.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0 //the input contains the public key that can unlock the output
}

func (outputTx *TxOutput) Lock(address []byte) {
	var pubKeyHash []byte = Wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-Wallet.ChecksumLength] //taking the bytes between version and checksum
	outputTx.PubKeyHash = pubKeyHash                                   //the output is locked with the public key hash, to be unlocked by the input's  public key
}

func (output *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubKeyHash) == 0
}
