package Wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "temp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func CreateWallets() (*Wallets, error) {
	var ws Wallets = Wallets{make(map[string]*Wallet)}
	var err error = ws.LoadFile()
	return &ws, err
}

func (wallets *Wallets) GetWallet(address string) Wallet {
	return *wallets.Wallets[address]
}

func (wallets *Wallets) AddWallet() string {
	var wallet *Wallet = MakeWallet()
	var byte_address = wallet.CreateAddress()
	var address string = fmt.Sprintf("%s", byte_address)

	wallets.Wallets[address] = wallet
	return address
}

func (wallets *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range wallets.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws *Wallets) SaveFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
