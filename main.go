package main

import (
	"log"
)

func main() {
	genesis_bc := BlockChainWithGenesis()

	genesis_bc.AddBlock("illusion")
	genesis_bc.AddBlock("Legion")
	genesis_bc.AddBlock("void")

	for _, block := range genesis_bc.Blocks {
		log.Println(block)
	}
}
