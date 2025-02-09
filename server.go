package main

const (
	protocol       = "tcp"
	node_version   = 1
	command_length = 12
)

var (
	node_address      string
	mining_address    string
	known_nodes       = []string{"localhost:3000"}
	blocks_in_transit = [][]byte{}
	mempool           = make(map[string]Transaction)
)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type get_blocks struct {
	AddrFrom string
}

type get_data struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type _version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}
