package Blockchain

import (
	"encoding/hex"

	"github.com/dgraph-io/badger"
)

const (
	utxoPrefix   = "utxo-"
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDeletion [][]byte) error {
		// if err := u.Blockchain.Database.Update(
		// 	func(txn *badger.Txn) error{
		// 		for _, key := range keysForDeletion{
		// 			err := txn.Delete(key)
		// 			Handle(err)
		// 		}
		// 		return nil
		// 	}); err != nil{
		// 	return err
		// }
		err := u.Blockchain.Database.Update(
			func(txn *badger.Txn) error {
				for _, key := range keysForDeletion {
					err := txn.Delete(key)
					Handle(err)
				}
				return nil
			})
		Handle(err)
		return nil //leave
	}

	collectSize := 1000
	u.Blockchain.Database.View(
		func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = collectSize
			opts.PrefetchValues = false
			var it *badger.Iterator = txn.NewIterator(opts)
			defer it.Close()

			keysForDeletion := make([][]byte, 0, collectSize)
			keysCollected := 0
			for it.Seek(prefix); /*all of the keys containing the prefix*/ it.ValidForPrefix(prefix); it.Next() {
				var key []byte = it.Item().KeyCopy(nil)
				keysForDeletion = append(keysForDeletion, key)
				keysCollected++
				if keysCollected == collectSize {
					err := deleteKeys(keysForDeletion)
					Handle(err) //yaha change
					keysForDeletion = make([][]byte, 0, collectSize)
					keysCollected = 0
				}
			}

			if keysCollected > 0 {
				err := deleteKeys(keysForDeletion)
				Handle(err)
			}
			return nil
		})

}

func (u UTXOSet) Reindex() {
	var db *badger.DB = u.Blockchain.Database
	u.DeleteByPrefix([]byte(utxoPrefix))

	var UTXOs map[string]TxOutputs = u.Blockchain.FindUTXO()
	var err error = db.Update(func(txn *badger.Txn) error {
		for txID, outputs := range UTXOs {
			key, err := hex.DecodeString(txID) //converting the hex string to byte array
			Handle(err)
			key = append([]byte(utxoPrefix), key...)
			err = txn.Set(key, outputs.SerializeOutputs())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

func (u UTXOSet) Update(block *Block) {
	var db *badger.DB = u.Blockchain.Database
	var err error = db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.Is_Coinbase() == false {
				for _, input := range tx.Inputs {
					var updatedOutputs TxOutputs
					var inID []byte = append([]byte(utxoPrefix), input.ID...)
					item, err := txn.Get(inID)
					Handle(err)
					value, err := item.ValueCopy([]byte{})
					Handle(err)

					var outputs TxOutputs = DeserializeOutputs(value)

					for outIdx, output := range outputs.Outputs {
						if outIdx != input.OutputIdx { //if the output is not the one being spent
							updatedOutputs.Outputs = append(updatedOutputs.Outputs, output)
						}
					}
					if len(updatedOutputs.Outputs) == 0 {
						err := txn.Delete(inID)
						Handle(err)
					} else {
						err := txn.Set(inID, updatedOutputs.SerializeOutputs())
						Handle(err)
					}
				}
			}
			//change here if problem arises
			var newOutputs TxOutputs
			for _, output := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, output)
			}
			var txID []byte = append([]byte(utxoPrefix), tx.ID...)
			err := txn.Set(txID, newOutputs.SerializeOutputs())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

func (u UTXOSet) CountTransactions() int {
	var db *badger.DB = u.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		var opts badger.IteratorOptions = badger.DefaultIteratorOptions
		var it *badger.Iterator = txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(utxoPrefix)); it.ValidForPrefix([]byte(utxoPrefix)); it.Next() {
			counter++
		}
		return nil
	})
	Handle(err)
	return counter
}

func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	var db *badger.DB = u.Blockchain.Database
	err := db.View(func(txn *badger.Txn) error {

		var opts badger.IteratorOptions = badger.DefaultIteratorOptions

		var it *badger.Iterator = txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(utxoPrefix)); it.ValidForPrefix([]byte(utxoPrefix)); it.Next(){
			var item *badger.Item = it.Item()

			value, err := item.ValueCopy([]byte{})
			Handle(err)

			var outputs TxOutputs = DeserializeOutputs(value)
			for _, output := range outputs.Outputs{
				if output.IsLockedWithKey(pubKeyHash){
					UTXOs = append(UTXOs, output)
				}
			}

		}
		return nil
	})
	Handle(err)
	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	var unspentOutputs map[string][]int = make(map[string][]int)
	var accumulated int = 0
	var db *badger.DB = u.Blockchain.Database
	err := db.View(func(txn *badger.Txn) error {
		var opts badger.IteratorOptions = badger.DefaultIteratorOptions
		var it *badger.Iterator = txn.NewIterator(opts)	
		defer it.Close()
		for it.Seek([]byte(utxoPrefix)); it.ValidForPrefix([]byte(utxoPrefix)); it.Next(){
			var item *badger.Item = it.Item()
			var key []byte = item.Key()
			value, err := item.ValueCopy([]byte{})
			Handle(err)

			key = key[prefixLength:] //removing the prefix

			var txID string = hex.EncodeToString(key)
			var outputs TxOutputs = DeserializeOutputs(value)

			for outIdx, out := range outputs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)
	return accumulated, unspentOutputs
}
