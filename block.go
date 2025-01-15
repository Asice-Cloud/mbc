package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Height int64

	PrevBlockHash []byte

	Data []byte

	Timestamp int64

	Hash []byte

	Nonce int64
}

func (b *Block) SetupHash() {
	// second param means n-decimal(2-36)
	time_string := strconv.FormatInt(b.Timestamp, 2)
	time_bytes := []byte(time_string)

	height_bytes := IntToBytes(b.Height)

	block_bytes := bytes.Join([][]byte{height_bytes, b.PrevBlockHash, b.Data, time_bytes, b.Hash}, []byte{})

	hash := sha256.Sum256(block_bytes)

	b.Hash = hash[:]

}

// factory method
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Data:          []byte(data),
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nonce:         0,
	}

	// through proof of work to judge and get valid hash, nonce
	pw := NewProofOfWork(block)
	hash, nonce := pw.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func GenesisBlock(data string) *Block {
	return NewBlock(data, 1,
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
