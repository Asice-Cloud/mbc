package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

func GobEncode(data any) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func node_is_known(addr string) bool {
	for _, node := range known_nodes {
		if node == addr {
			return true
		}
	}
	return false
}

func command_to_bytes(command string) []byte {
	var bytes [command_length]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytes_to_command(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func extract_command(request []byte) []byte {
	return request[:command_length]
}

func request_blocks() {
	for _, node := range known_nodes {
		send_get_blocks(node)
	}
}

func send_addr(address string) {
	nodes := addr{known_nodes}
	nodes.AddrList = append(nodes.AddrList, node_address)
	payload := GobEncode(nodes)
	request := append(command_to_bytes("addr"), payload...)

	send_data(address, request)
}

func send_block(address string, b *Block) {
	data := block{node_address, b.Serialize()}
	payload := GobEncode(data)
	request := append(command_to_bytes("block"), payload...)

	send_data(address, request)
}

func send_data(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var update_nodes []string
		for _, node := range known_nodes {
			if node != addr {
				update_nodes = append(update_nodes, node)
			}
		}
		known_nodes = update_nodes
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func send_inv(address string, kind string, items [][]byte) {
	inventory := inv{node_address, kind, items}
	payload := GobEncode(inventory)
	request := append(command_to_bytes("inv"), payload...)

	send_data(address, request)
}

func send_get_blocks(address string) {
	payload := GobEncode(get_blocks{node_address})
	request := append(command_to_bytes("getblocks"), payload...)

	send_data(address, request)
}

func send_get_data(address string, kind string, id []byte) {
	payload := GobEncode(get_data{node_address, kind, id})
	request := append(command_to_bytes("getdata"), payload...)

	send_data(address, request)
}

func send_tx(address string, tnx *Transaction) {
	data := tx{node_address, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(command_to_bytes("tx"), payload...)

	send_data(address, request)
}

func send_version(addr string, bc *BlockChain) {
	best_height := bc.GetBestHeight()
	payload := GobEncode(_version{node_version, best_height, node_address})
	request := append(command_to_bytes("version"), payload...)

	send_data(addr, request)
}

func handle_addr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	known_nodes = append(known_nodes, payload.AddrList...)
	log.Printf("There are %d known nodes now!\n", len(known_nodes))
	request_blocks()
}

func handle_block(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	block := DeserializeBlock(blockData)

	log.Println("Recevied a new block!")
	bc.AddBlock(block)
	log.Printf("Added block %x\n", block.Hash)

	if len(blocks_in_transit) > 0 {
		block_hash := blocks_in_transit[0]
		send_get_data(payload.AddrFrom, "block", block_hash)
		blocks_in_transit = blocks_in_transit[1:]
	} else {
		UTXO_set := UTXO_set{bc}
		UTXO_set.Reindex()
	}
}

func handle_inv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)
	if payload.Type == "block" {
		blocks_in_transit = payload.Items
		block_hash := payload.Items[0]
		send_get_data(payload.AddrFrom, "block", block_hash)

		new_in_transit := [][]byte{}
		for _, b := range blocks_in_transit {
			if bytes.Compare(b, block_hash) != 0 {
				new_in_transit = append(new_in_transit, b)
			}
		}
		blocks_in_transit = new_in_transit
	}

	if payload.Type == "tx" {
		tx_id := payload.Items[0]
		if mempool[hex.EncodeToString(tx_id)].ID == nil {
			send_get_data(payload.AddrFrom, "tx", tx_id)
		}
	}
}

func handle_get_blocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload get_blocks

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	send_inv(payload.AddrFrom, "block", blocks)
}

func handle_get_data(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload get_data

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		send_block(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		tx_id := hex.EncodeToString(payload.ID)
		tx := mempool[tx_id]
		send_tx(payload.AddrFrom, &tx)
		delete(mempool, tx_id)
	}
}

func handle_tx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := Deserialize_transaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if node_address == known_nodes[0] {
		for _, node := range known_nodes {
			if node != node_address && node != payload.AddFrom {
				send_inv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(mining_address) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbtx := NewCoinbaseTX(mining_address, "")
			txs = append(txs, cbtx)

			new_block := bc.MineBlock(txs)
			utso_set := UTXO_set{bc}
			utso_set.Reindex()

			log.Println("New block is mined!")

			for _, tx := range txs {
				tx_id := hex.EncodeToString(tx.ID)
				delete(mempool, tx_id)
			}

			for _, node := range known_nodes {
				if node != node_address {
					send_inv(node, "block", [][]byte{new_block.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handle_version(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload _version

	buff.Write(request[command_length:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	my_best_height := bc.GetBestHeight()
	foreigner_best_height := payload.BestHeight

	if my_best_height < foreigner_best_height {
		send_get_blocks(payload.AddrFrom)
	} else if my_best_height > foreigner_best_height {
		send_version(payload.AddrFrom, bc)
	}
	send_addr(payload.AddrFrom)
	if !node_is_known(payload.AddrFrom) {
		known_nodes = append(known_nodes, payload.AddrFrom)
	}
}
