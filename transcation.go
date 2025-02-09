package main

const (
	subsidy = 10
)

type Transaction struct {
	ID   []byte
	Vin  []TX_input
	Vout []TX_output
}

type TX_input struct {
	TXID      []byte
	Vout      int
	Signature []byte
	PublicKey []byte
}

type TX_output struct {
	Value         int
	PublicKeyHash []byte
}

type TX_outputs struct {
	Outputs []TX_output
}
