package main

import (
	"crypto/sha256"
)

//merkle tree

type HashTree struct {
	Root *HashNode
}

type HashNode struct {
	Left  *HashNode
	Right *HashNode
	Data  []byte
}

func NewHashNode(left, right *HashNode, data []byte) *HashNode {
	node := HashNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prev := append(left.Data, right.Data...)
		hash := sha256.Sum256(prev)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

func NewHashTree(data [][]byte) *HashTree {
	var nodes []HashNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, dd := range data {
		node := NewHashNode(nil, nil, dd)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var new_level []HashNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewHashNode(&nodes[j], &nodes[j+1], nil)
			new_level = append(new_level, *node)
		}
		nodes = new_level
	}

	return &HashTree{&nodes[0]}
}
