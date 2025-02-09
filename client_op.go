package main

import (
	"log"
	"strconv"
)

func (c *Client) get_balance(address, node_id string) {
	if !ValidAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := NewBlockChain(node_id)
	UTXO_set := UTXO_set{bc}
	defer bc.db.Close()

	balance := 0
	public_hash := B58Decode([]byte(address))
	public_hash = public_hash[1 : len(public_hash)-sum_len]
	UTXOs := UTXO_set.FindUTXO(public_hash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("Balance of '%s': %d\n", address, balance)
}

func (c *Client) create_blockchain(address, node_id string) {
	if !ValidAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockChain(address, node_id)
	defer bc.db.Close()
	UTXO_set := UTXO_set{bc}
	UTXO_set.Reindex()

	log.Println("Successfullly created blockchain")
}

func (c *Client) create_wallet(node_id string) {
	wallets, _ := NewWallets(node_id)
	address := wallets.CreateWallet()
	wallets.SaveToFile(node_id)
	log.Printf("New address is: %s\n", address)
}

func (c *Client) list_addresses(node_id string) {
	wallets, _ := NewWallets(node_id)
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		log.Println(address)
	}
}

func (c *Client) print_chain(node_id string) {
	bc := NewBlockChain(node_id)
	defer bc.db.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()
		log.Printf("Height: %d\n", block.Height)
		log.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		log.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		log.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (c *Client) reindex_UTXO(node_id string) {
	bc := NewBlockChain(node_id)
	defer bc.db.Close()
	UTXO_set := UTXO_set{bc}
	UTXO_set.Reindex()
	log.Println("Reindexed UTXO set")
}

func (c *Client) send(from, to string, amount int, node_id string, mine_now bool) {
	if !ValidAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockChain(node_id)
	utxo_set := UTXO_set{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(node_id)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &utxo_set)

	if mine_now {
		cb_tx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cb_tx, tx}

		new_block := bc.MineBlock(txs)
		utxo_set.Update(new_block)
	} else {
		send_tx(known_nodes[0], tx)
	}

	log.Println("Success: Send")
}

func (c *Client) start_node(miner_address, node_id string) {
	log.Printf("Starting node %s\n", node_id)
	if len(miner_address) > 0 {
		if !ValidAddress(miner_address) {
			log.Panic("ERROR: Miner address is not valid")
		}
	}
	start_server(node_id, miner_address)
}
