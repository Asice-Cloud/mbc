package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"log"
	"os"
)

const (
	version     = byte(0x00)
	sum_len     = 4
	wallet_file = "wallets_%s.data"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	wallet := Wallet{}
	return &wallet
}

// wallet
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, public
}

func (w Wallet) GetAddress() []byte {
	public_hash := Hash_Pub_Key(w.PublicKey)

	payload := append([]byte{version}, public_hash...)
	check_sum := check_sum(payload)

	full := append(payload, check_sum...)
	address := B58Encode(full)
	return address
}

func Hash_Pub_Key(public []byte) []byte {
	public_256 := sha256.Sum256(public)

	repimd160 := ripemd160.New()
	_, err := repimd160.Write(public_256[:])
	if err != nil {
		log.Panic(err)
	}
	public_160 := repimd160.Sum(nil)

	return public_160
}

func check_sum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	return second[:sum_len]
}

func ValidAddress(address string) bool {
	pub_key_hash := B58Decode([]byte(address))
	actual_check_sum := pub_key_hash[len(pub_key_hash)-sum_len:]
	version := pub_key_hash[0]
	pub_key_hash = pub_key_hash[1 : len(pub_key_hash)-sum_len]

	target_check_sum := check_sum(append([]byte{version}, pub_key_hash...))

	return bytes.Compare(actual_check_sum, target_check_sum) == 0
}

// Wallets
func NewWallets(node_id string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile(node_id)

	return &wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFromFile(node_id string) error {
	wallet_file := fmt.Sprintf(wallet_file, node_id)
	if _, err := os.Stat(wallet_file); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(wallet_file)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws *Wallets) SaveToFile(node_id string) {
	var content bytes.Buffer

	wallet_file := fmt.Sprintf(wallet_file, node_id)
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(wallet_file, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
