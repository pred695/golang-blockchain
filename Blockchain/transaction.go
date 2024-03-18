package Blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	ID        []byte //references to the previous output that led to the input
	OutputIdx int    //index of the referenced output which is spent in the transaction
	Sign      string //signature or the sender's public key
}

// Outpoint is the index of the output in the transaction + the transaction id
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer //more efficient than using strings/[]byte
	//encode the transaction
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded) //the encoder writes to the buffer and stores in the "encoded variable"
	err := encoder.Encode(tx)           //encodes the transaction into a byte slice
	Handle(err)
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// Coinbase Transaction --> A transaction that creates a new coin, it is the first transaction in a block(rewarding transaction).
// it has no inputs(no reference to previous outputs and no outpoint) and only one output.
func CoinbaseTx(rec_address string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", rec_address)
	}
	txin := TxInput{ID: []byte{}, OutputIdx: -1, Sign: data}
	txout := TxOutput{Value: 100, PubKey: rec_address}
	tx := Transaction{ID: nil, Inputs: []TxInput{txin}, Outputs: []TxOutput{txout}}
	tx.SetID() //creates the hash id for the transaction
	return &tx
}

func (chain *Blockchain) NewTransaction(senders_add string, rec_address string, amount int) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	var accumulated int
	var spendableOutputs map[string][]int
	accumulated, spendableOutputs = chain.FindSpendableOutputs(senders_add, amount)
	if accumulated < amount {
		log.Panic("Error: Not enough funds")
	}
	for txid, outputs := range spendableOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)
		for _, output := range outputs {
			var input TxInput = TxInput{ID: txID, OutputIdx: output, Sign: senders_add}
			inputs = append(inputs, input)	//creating inputs for the unspent outputs(since we have the receiver's address and the amount)
		}
	}
	outputs = append(outputs, TxOutput{Value: amount, PubKey: rec_address}) //creating the output for the receiver

	if accumulated > amount{
		outputs = append(outputs, TxOutput{Value: accumulated - amount, PubKey: senders_add}) //creating the output for the sender(change amount in case if the sender sends more than the amount)
	}

	var tx Transaction = Transaction{ID: nil, Inputs: inputs, Outputs: outputs}
	tx.SetID()
	return &tx
}

func (tx *Transaction) Is_Coinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].OutputIdx == -1
}

func (inputTx *TxInput) CanUnlock(data string) bool {
	return inputTx.Sign == data
}

func (outputTx *TxOutput) CorrectRecAddress(data string) bool {
	return outputTx.PubKey == data
}
