package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Height        int
	PrevBlockHash []byte
	Transactions  []*Transaction
	Timestamp     int64
	Hash          []byte
	Nonce         int
}

func NewBlock(transactions []*Transaction, prev_hash []byte, height int) *Block {
	block := &Block{height, prev_hash, transactions, time.Now().Unix(), []byte{}, 0}
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) GetTranscation() []byte {
	var transcations [][]byte
	for _, tx := range b.Transactions {
		transcations = append(transcations, tx.Serialize())
	}
	ht := NewHashTree(transcations)

	return ht.Root.Data
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(bb []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(bb))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
