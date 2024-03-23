package Wallet

import (
	"github.com/mr-tron/base58"
	"github.com/pred695/golang-blockchain/Blockchain"
)

// Base58 uses 6 less characters than base64, and is more human readable. the characters removed are 0, O, I, l, +, /
// Encoding the byte slice to base58(decoding into string and then converting to byte slice)
func Base58Encode(input []byte) []byte {
	var encoded string = base58.Encode(input)
	return []byte(encoded) //converting string to byte slice
}

func Base58Decode(input []byte) []byte {
	decoded, err := base58.Decode(string(input[:])) //input[:] references the original byte slice
	Blockchain.Handle(err)
	return decoded
}
