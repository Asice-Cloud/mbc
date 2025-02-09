package main

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const (
	database              = "blockchain__%s.db"
	buckets               = "blocks"
	genesis_coinbase_data = "What is love, baby don't hurt me, don't hurt me, no more"
)

type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

func NewBlockChain(node_id string) *BlockChain {
	dbfd := fmt.Sprintf(database, node_id)
	if db_exists(dbfd) == false {
		log.Default().Println("No existing blockchain found. Create one!")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbfd, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := BlockChain{tip, db}
	return &bc
}

func CreateBlockChain(address, node_id string) *BlockChain {
	dbfd := fmt.Sprintf(database, node_id)
	if db_exists(dbfd) {
		log.Default().Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesis_coinbase_data)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbfd, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(buckets))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}
	return &bc
}

// iterator
type BI struct {
	current_hash []byte
	db           *bolt.DB
}

func (i *BI) Next() *Block {
	var block *Block
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		encoded_block := b.Get(i.current_hash)
		block = DeserializeBlock(encoded_block)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.current_hash = block.PrevBlockHash
	return block
}

func (bc *BlockChain) Iterator() *BI {
	bci := &BI{bc.tip, bc.db}
	return bci
}

// ops
func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		block_in_chain := b.Get(block.Hash)
		if block_in_chain != nil {
			return nil
		}
		block_data := block.Serialize()
		err := b.Put(block.Hash, block_data)
		if err != nil {
			log.Panic(err)
		}
		last_hash := b.Get([]byte("l"))
		last_block := DeserializeBlock(last_hash)
		if block.Height > last_block.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockChain) FindTransaction(id []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, id) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, fmt.Errorf("Transaction not found")
}

func (bc *BlockChain) Find_UTXO() map[string]TX_outputs {
	UTXO := make(map[string]TX_outputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := string(tx.ID)
		Outputs:
			for out_idx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spent_out := range spentTXOs[txID] {
						if spent_out == out_idx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					in_txID := string(in.TXID)
					spentTXOs[in_txID] = append(spentTXOs[in_txID], in.Vout)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *BlockChain) GetBestHeight() int {
	var last_block *Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		last_hash := b.Get([]byte("l"))
		last_block = DeserializeBlock(b.Get(last_hash))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return last_block.Height
}

func (bc *BlockChain) GenerateHeight() int {
	var last_block *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		last_hash := b.Get([]byte("l"))
		last_block = DeserializeBlock(b.Get(last_hash))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return last_block.Height
}

func (bc *BlockChain) GetBlock(block_hash []byte) (Block, error) {
	var block Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		encoded_block := b.Get(block_hash)
		if encoded_block == nil {
			return fmt.Errorf("Block not found")
		}
		block = *DeserializeBlock(encoded_block)
		return nil
	})
	if err != nil {
		return Block{}, err
	}
	return block, nil
}

func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()
	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return blocks
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var last_hash []byte
	var last_height int

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		last_hash = b.Get([]byte("l"))
		last_block := DeserializeBlock(b.Get(last_hash))
		last_height = last_block.Height
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	new_block := NewBlock(transactions, last_hash, last_height+1)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(buckets))
		err := b.Put(new_block.Hash, new_block.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), new_block.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = new_block.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return new_block
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TXID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[string(vin.TXID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *BlockChain) SignTransaction(tx *Transaction, private_key ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TXID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[string(vin.TXID)] = prevTX
	}

	tx.Sign(private_key, prevTXs)
}
