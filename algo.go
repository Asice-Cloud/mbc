package main

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58
func B58Encode(in []byte) []byte {
	var res []byte
	x := big.NewInt(0).SetBytes(in)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		res = append(res, b58Alphabet[mod.Int64()])
	}

	if in[0] == 0x00 {
		res = append(res, b58Alphabet[0])
	}

	Reverse(res)
	return res
}

func B58Decode(in []byte) []byte {
	res := big.NewInt(0)
	for _, b := range in {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		res.Mul(res, big.NewInt(58))
		res.Add(res, big.NewInt(int64(charIndex)))
	}

	resBytes := res.Bytes()
	if in[0] == b58Alphabet[0] {
		resBytes = append([]byte{0x00}, resBytes...)
	}

	return resBytes
}
