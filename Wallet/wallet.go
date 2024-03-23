package Wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	version        = byte(0x00)
	checksumLength = 4 //in bytes.
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

/*
	Private Key --> ecdsa --> Public Key --> sha256 --> ripemd 160 --> public key hash
	Parallely,
	public key hash --> sha256 --> sha256 -->  Take first 4 bytes --> checksum + public key hash + version --> base58 --> address
*/

func (w Wallet) CreateAddress() []byte {
	var pubHash []byte = CreatePubKeyHash(w.PublicKey)

	var versionHash []byte = append([]byte{version}, pubHash...)
	var checksum []byte = Checksum(versionHash)

	var fullHash []byte = append(versionHash, checksum...)
	var address []byte = Base58Encode(fullHash)

	// fmt.Printf("Private Key: %x\n", w.PrivateKey)
	// fmt.Printf("Public Key: %x\n", w.PublicKey)
	// fmt.Printf("Public Key Hash: %x\n", pubHash)
	fmt.Printf("Address: %s\n", address)

	return address
}

// This function creates our private and public key
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	var curve elliptic.Curve = elliptic.P256() //(Outputs on this elliptic curve can range from 0 to 2 ^ 256{256 bytes})
	var PrivateKey *ecdsa.PrivateKey
	var err error
	PrivateKey, err = ecdsa.GenerateKey(curve, rand.Reader)
	Handle(err)
	var PublicKey []byte = append(PrivateKey.PublicKey.X.Bytes(), PrivateKey.PublicKey.Y.Bytes()...)
	return *PrivateKey, PublicKey
}

func MakeWallet() *Wallet {
	var PrivateKey ecdsa.PrivateKey
	var PublicKey []byte
	PrivateKey, PublicKey = NewKeyPair()
	var wallet Wallet = Wallet{PrivateKey, PublicKey}
	return &wallet
}

func CreatePubKeyHash(PubKey []byte) []byte {
	var PubKeyHash [32]byte = sha256.Sum256(PubKey) //sha256 returns a 32 byte array, taking a byte slice as input
	hasher := ripemd160.New()
	_, err := hasher.Write(PubKeyHash[:]) //writing the public key hash to the hasher
	Handle(err)
	var finalHash []byte = hasher.Sum(nil) //hasher.Sum returns a byte slice
	return finalHash
}

func Checksum(payload []byte) []byte {
	var firstHash [32]byte = sha256.Sum256(payload)
	var secondHash [32]byte = sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength] //first 4 bytes of the second hash

}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
