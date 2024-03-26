package Wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const (
	version        = byte(0x00)
	ChecksumLength = 4 //in bytes.
)

type Wallet struct {
	PrivateKey PrivateKey
	PublicKey  []byte
}

type PrivateKey struct {
	D *big.Int
	X *big.Int
	Y *big.Int
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

func ValidateAddress(address string) bool {
	var pubKeyHash []byte = Base58Decode([]byte(address))
	var actualChecksum []byte = pubKeyHash[(len(pubKeyHash) - ChecksumLength):] //taking the last ChecksumLength bytes
	var version byte = pubKeyHash[0]
	pubKeyHash = pubKeyHash[1:(len(pubKeyHash) - ChecksumLength)] //taking the bytes between version and checksum
	var targetChecksum []byte = Checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// This function creates our private and public key
func NewKeyPair() (PrivateKey, []byte) {
	var curve elliptic.Curve = elliptic.P256() //(Outputs on this elliptic curve can range from 0 to 2 ^ 256{256 bytes})
	var Priv *ecdsa.PrivateKey
	var err error
	Priv, err = ecdsa.GenerateKey(curve, rand.Reader)
	Handle(err)
	var PublicKey []byte = append(Priv.PublicKey.X.Bytes(), Priv.PublicKey.Y.Bytes()...)
	return PrivateKey{D: Priv.D, X: Priv.PublicKey.X, Y: Priv.PublicKey.Y}, PublicKey
}
func MakeWallet() *Wallet {
	var PrivateKey PrivateKey
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
	return secondHash[:ChecksumLength] //first 4 bytes of the second hash

}

func (priv PrivateKey) ToECDSA() ecdsa.PrivateKey {
	return ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: priv.X, Y: priv.Y}, D: priv.D}
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
