package main

import "github.com/boltdb/bolt"

const utxo_bucket = "chainstate"

type UTXO_set struct {
	BlockChain *BlockChain
}

func (u *UTXO_set) FindSpendableOutputs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspent_outputs := make(map[string][]int)
	accumulated := 0
	db := u.BlockChain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxo_bucket))
		cur := bucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			tx_id := string(k)
			outs := deserialize_outputs(v)

			for out_idx, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspent_outputs[tx_id] = append(unspent_outputs[tx_id], out_idx)
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return accumulated, unspent_outputs
}

func (u *UTXO_set) FindUTXO(publicKeyHash []byte) []TX_output {
	var UTXOs []TX_output
	db := u.BlockChain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxo_bucket))
		cur := bucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			outs := deserialize_outputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publicKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return UTXOs
}

func (u *UTXO_set) CountTransactions() int {
	count := 0
	db := u.BlockChain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxo_bucket))
		cur := bucket.Cursor()

		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			count++
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return count
}

func (u *UTXO_set) Reindex() {
	db := u.BlockChain.db
	bucket_name := []byte(utxo_bucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucket_name)
		if err != nil && err != bolt.ErrBucketNotFound {
			panic(err)
		}

		_, err = tx.CreateBucket(bucket_name)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	UTXO := u.BlockChain.Find_UTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_name)

		for tx_id, outs := range UTXO {
			key := []byte(tx_id)
			err := bucket.Put(key, outs.serialize_outputs())
			if err != nil {
				panic(err)
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func (u UTXO_set) Update(block *Block) {
	db := u.BlockChain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxo_bucket))
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					update_outs := TX_outputs{}
					outs_bytes := bucket.Get(vin.TXID)
					outs := deserialize_outputs(outs_bytes)

					for out_idx, out := range outs.Outputs {
						if out_idx != vin.Vout {
							update_outs.Outputs = append(update_outs.Outputs, out)
						}
					}

					if len(update_outs.Outputs) == 0 {
						err := bucket.Delete(vin.TXID)
						if err != nil {
							panic(err)
						}
					} else {
						err := bucket.Put(vin.TXID, update_outs.serialize_outputs())
						if err != nil {
							panic(err)
						}
					}
				}
			}

			new_outputs := TX_outputs{}
			for _, out := range tx.Vout {
				new_outputs.Outputs = append(new_outputs.Outputs, out)
			}

			err := bucket.Put(tx.ID, new_outputs.serialize_outputs())
			if err != nil {
				panic(err)
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
}
