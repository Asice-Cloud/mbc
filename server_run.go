package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handle_connection(conn net.Conn, bc *BlockChain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytes_to_command(extract_command(request))
	log.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handle_addr(request)
	case "block":
		handle_block(request, bc)
	case "inv":
		handle_inv(request, bc)
	case "getblocks":
		handle_get_blocks(request, bc)
	case "getdata":
		handle_get_data(request, bc)
	case "tx":
		handle_tx(request, bc)
	case "version":
		handle_version(request, bc)
	default:
		log.Println("Unknown command!")

	}
	conn.Close()
}

func start_server(node_id, miner_address string) {
	node_address = fmt.Sprintf("localhost:%s", node_id)
	mining_address = miner_address
	ln, err := net.Listen(protocol, node_address)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	bc := NewBlockChain(node_id)
	if node_address != known_nodes[0] {
		send_version(known_nodes[0], bc)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handle_connection(conn, bc)
	}
}
