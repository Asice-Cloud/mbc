package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

// transaction
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TXID) == 0 && tx.Vin[0].Vout == -1
}

func (tx Transaction) Serialize() []byte {
	var encoder bytes.Buffer

	enc := gob.NewEncoder(&encoder)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoder.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TX_input
	var outputs []TX_output

	for _, vin := range tx.Vin {
		inputs = append(inputs, TX_input{vin.TXID, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TX_output{vout.Value, vout.PublicKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Sign(private ecdsa.PrivateKey, prev_txs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prev_txs[hex.EncodeToString(vin.TXID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	for in_id, vin := range txCopy.Vin {
		prev_tx := prev_txs[hex.EncodeToString(vin.TXID)]
		txCopy.Vin[in_id].Signature = nil
		txCopy.Vin[in_id].PublicKey = prev_tx.Vout[vin.Vout].PublicKeyHash

		data_to_sign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &private, []byte(data_to_sign))
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[in_id].Signature = signature
		txCopy.Vin[in_id].PublicKey = nil
	}
}

// returns human readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PublicKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PublicKeyHash))
	}

	return strings.Join(lines, "\n")
}

// verify signature
func (tx *Transaction) Verify(prev_txs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	for _, vin := range tx.Vin {
		if prev_txs[hex.EncodeToString(vin.TXID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for in_id, vin := range tx.Vin {
		prev_tx := prev_txs[hex.EncodeToString(vin.TXID)]
		txCopy.Vin[in_id].Signature = nil
		txCopy.Vin[in_id].PublicKey = prev_tx.Vout[vin.Vout].PublicKeyHash

		r := big.Int{}
		s := big.Int{}

		sig_len := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sig_len / 2)])
		s.SetBytes(vin.Signature[(sig_len / 2):])

		x := big.Int{}
		y := big.Int{}
		key_len := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(key_len / 2)])
		y.SetBytes(vin.PublicKey[(key_len / 2):])

		raw_pub_key := ecdsa.PublicKey{curve, &x, &y}
		data_to_verify := fmt.Sprintf("%x\n", txCopy)

		if ecdsa.Verify(&raw_pub_key, []byte(data_to_verify), &r, &s) == false {
			return false
		}
		txCopy.Vin[in_id].PublicKey = nil
	}
	return true
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TX_input{[]byte{}, -1, nil, []byte(data)}
	txout := TX_output{subsidy, nil}
	tx := Transaction{nil, []TX_input{txin}, []TX_output{txout}}
	tx.ID = tx.Hash()
	return &tx
}

func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXO_set *UTXO_set) *Transaction {
	var inputs []TX_input
	var outputs []TX_output

	public := Hash_Pub_Key(wallet.PublicKey)
	acc, Valid_outputs := UTXO_set.FindSpendableOutputs(public, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for tx_id, outs := range Valid_outputs {
		tx_id, _ := hex.DecodeString(tx_id)
		for _, out := range outs {
			input := TX_input{tx_id, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTX_output(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTX_output(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO_set.BlockChain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

func Deserialize_transaction(data []byte) Transaction {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		log.Panic(err)
	}
	return tx
}

// input
func (in *TX_input) UseKey(public []byte) bool {
	lock := Hash_Pub_Key(in.PublicKey)

	return bytes.Compare(lock, public) == 0
}

// output
func NewTX_output(value int, address string) *TX_output {
	tx_o := &TX_output{value, nil}
	tx_o.Lock([]byte(address))

	return tx_o
}

func (out TX_output) Lock(address []byte) {
	public := B58Decode(address)
	public = public[1 : len(public)-4]
	out.PublicKeyHash = public
}

func (out TX_output) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, pubKeyHash) == 0
}

// outputs
func (outs TX_outputs) serialize_outputs() []byte {
	var buff bytes.Buffer

	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func deserialize_outputs(data []byte) TX_outputs {
	var outputs TX_outputs

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}
