package Blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/pred695/golang-blockchain/Wallet"
)

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
	txin := TxInput{ID: []byte{}, OutputIdx: -1, Signature: nil, PubKey: []byte(data)}
	txout := NewTxOutput(100, rec_address)
	tx := Transaction{ID: nil, Inputs: []TxInput{txin}, Outputs: []TxOutput{*txout}}
	tx.SetID() //creates the hash id for the transaction
	return &tx
}

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	var encoder *gob.Encoder = gob.NewEncoder(&encoded)
	var err error = encoder.Encode(tx)
	Handle(err)

	return encoded.Bytes()
}

func (tx *Transaction) HashTransaction() []byte {
	var hash [32]byte

	var txCopy Transaction = *tx
	txCopy.ID = []byte{} //setting this to nil since we don't want to hash the ID, creating the circular dependency, the output of this function is going to be used as the ID.

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) Is_Coinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].OutputIdx == -1
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, input := range tx.Inputs {
		inputs = append(inputs, TxInput{ID: input.ID, OutputIdx: input.OutputIdx, Signature: nil, PubKey: nil})
	}

	for _, output := range tx.Outputs {
		outputs = append(outputs, TxOutput{Value: output.Value, PubKeyHash: output.PubKeyHash})

	}
	var txCopy Transaction = Transaction{ID: tx.ID, Inputs: inputs, Outputs: outputs}
	return txCopy
}

func (chain *Blockchain) NewTransaction(send_address string, rec_address string, amount int) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := Wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(send_address)
	var pubKeyHash []byte = Wallet.CreatePubKeyHash(w.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, rec_address)) //creating the output for the receiver

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, send_address)) //sending the remaining amount back to the sender
	}

	var tx Transaction = Transaction{ID: nil, Inputs: inputs, Outputs: outputs}
	tx.ID = tx.HashTransaction()
	chain.SignTransaction(&tx, w.PrivateKey.ToECDSA())

	return &tx
}

func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if tx.Is_Coinbase() {
		return true
	}

	for _, input := range tx.Inputs {
		if prevTxs[hex.EncodeToString(input.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction does not exist")
		}
	}

	var txCopy Transaction = tx.TrimmedCopy() //copy of the transaction without the signature and public key
	var curve elliptic.Curve = elliptic.P256()

	for inID, input := range txCopy.Inputs {
		var prevTx Transaction = prevTxs[hex.EncodeToString(input.ID)] //getting the previous transactions referenced by the input
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[input.OutputIdx].PubKeyHash
		txCopy.ID = txCopy.HashTransaction()
		txCopy.Inputs[inID].PubKey = nil //clearing it again so it doesn't affect the next iteration and signing

		var r, s big.Int //(Signing Component, Nonce Component)
		var signatureLen int = len(input.Signature)
		r.SetBytes(input.Signature[:(signatureLen / 2)])
		s.SetBytes(input.Signature[(signatureLen / 2):])

		var x, y big.Int
		var keyLen int = len(input.PubKey)
		x.SetBytes(input.PubKey[:(keyLen / 2)])
		y.SetBytes(input.PubKey[(keyLen / 2):])

		var rawPubKey ecdsa.PublicKey = ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		return ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s)
	}
	return false
}

func (tx *Transaction) Sign(private_key ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	//map[string][Transaction] --> map[string(hash of the transaction)] = transaction
	if tx.Is_Coinbase() {
		return
	}
	for _, input := range tx.Inputs {
		if prevTXs[hex.EncodeToString((input.ID))].ID == nil {
			log.Panic("ERROR: Previous transaction does not exist")
		}
	}
	var txCopy = tx.TrimmedCopy() //copy of the transaction without the signature and public key
	for inID, input := range txCopy.Inputs {

		var prevTx Transaction = prevTXs[hex.EncodeToString(input.ID)] //getting the previous transactions referenced by the input
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[input.OutputIdx].PubKeyHash
		txCopy.ID = txCopy.HashTransaction()
		txCopy.Inputs[inID].PubKey = nil //clearing it again so it doesn't affect the next iteration and signing
		r, s, err := ecdsa.Sign(rand.Reader, &private_key, txCopy.ID)
		Handle(err)
		var signature []byte = append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inID].Signature = signature
	}

}

func (priv PrivateKey) ToECDSA() ecdsa.PrivateKey {
	return ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: priv.X, Y: priv.Y}, D: priv.D}
}

func (tx *Transaction) To_String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for idx, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("Input %d:", idx))
		lines = append(lines, fmt.Sprintf("	Transaction ID: %x", input.ID))
		lines = append(lines, fmt.Sprintf("	Output Index: %d", input.OutputIdx))
		lines = append(lines, fmt.Sprintf("	Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("	PubKey: %x", input.PubKey))
	}

	for idx, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf(" Output %d:", idx))
		lines = append(lines, fmt.Sprintf("	Value: %d", output.Value))
		lines = append(lines, fmt.Sprintf("	PubKeyHash: %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}
