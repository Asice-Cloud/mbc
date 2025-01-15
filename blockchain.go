package main

type Blockchain struct {
	Blocks []*Block
}

func BlockChainWithGenesis() *Blockchain {
	genesis := GenesisBlock("Genesis")
	return &Blockchain{[]*Block{genesis}}
}

func (bc *Blockchain) AddBlock(data string) {
	pre_block := bc.Blocks[len(bc.Blocks)-1]
	new_block := NewBlock(data, pre_block.Height+1, pre_block.Hash)
	bc.Blocks = append(bc.Blocks, new_block)
}

func (bc *Blockchain) AddInto(data string, height int64, pre_hash []byte) {
	new_block := NewBlock(data, height, pre_hash)
	bc.Blocks = append(bc.Blocks, new_block)
}
