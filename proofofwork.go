package main

import (
	"bytes"
	"crypto/sha256"
	"log"
	"math"
	"math/big"
	"time"
)

// def difficulty of mining
const (
	difficulty = 5
	max_nonce  = math.MaxInt64
	//max_nonce = 36
)

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	return &ProofOfWork{Block: block, Target: target}
}

func (pow *ProofOfWork) prepare(nonce int64) []byte {
	return bytes.Join([][]byte{
		pow.Block.PrevBlockHash,
		pow.Block.Data,
		IntToBytes(pow.Block.Timestamp),
		IntToBytes(int64(difficulty)),
		IntToBytes(int64(nonce)),
		IntToBytes(int64(pow.Block.Height)),
	}, []byte{})
}

func (pow *ProofOfWork) Run() ([]byte, int64) {
	var hash_int big.Int
	var hash [32]byte
	nonce := int64(0)

	t1 := time.Now()
	for nonce < max_nonce {
		data := pow.prepare(nonce)
		hash = sha256.Sum256(data)

		hash_int.SetBytes(hash[:])

		if hash_int.Cmp(pow.Target) == -1 {
			break
		}
		nonce++
	}
	log.Println(time.Since(t1))
	return hash[:], nonce
}
