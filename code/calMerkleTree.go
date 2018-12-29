package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// CalRoot : Calculating Merkle Tree Root for an array of []byte
func CalRoot(data [][]byte) (headRoot []byte) {
	var Root Node
	return Root.GenerateRoot(data, false).NodeHash
}

// CalTree : Calculating Merkle Tree, and print the whole tree
func CalTree(data [][]byte) {
	var Root Node
	Root.GenerateRoot(data, true)
}

// Node : Node of a Merkle Tree
type Node struct {
	NodeData []byte
	NodeHash []byte
}

// CalSHA256Hash : Calculate a sha256 hash
func (n Node) CalSHA256Hash(input []byte) []byte {
	h := sha256.New()
	h.Write(input)
	return h.Sum(nil)
}

// GenerateRoot : Create Merkle Tree
func (n Node) GenerateRoot(data [][]byte, printFlag bool) (RootNode *Node) {

	// Prepare leaf nodes
	var nodes []*Node

	for i := 0; i < len(data); i = i + 1 {
		nodes = append(nodes, &Node{
			NodeData: data[i],
			NodeHash: n.CalSHA256Hash(data[i]),
		})
	}

	// Print the leaf node if printFlag is ON
	if printFlag == true {
		fmt.Printf("\nLeaf Nodes: \n")
		for i := 0; i < len(nodes); i = i + 1 {
			fmt.Printf("%x, [%s]\n", nodes[i].NodeHash, nodes[i].NodeData)
		}
	}

	// Build tree from bottom layer
	for {
		if len(nodes) != 1 {

			// If the number of node in bottom layer is not even, replicate and push the last node
			if len(nodes)%2 == 1 {
				nodes = append(nodes, &Node{
					NodeData: nodes[len(nodes)-1].NodeData,
					NodeHash: nodes[len(nodes)-1].NodeHash,
				})
			}

			// Build the upper level nodes
			var upperNodes []*Node
			for i := 0; i < len(nodes); i = i + 2 {
				upperNodes = append(upperNodes, &Node{
					NodeData: bytes.Join([][]byte{nodes[i].NodeHash, nodes[i+1].NodeHash}, []byte{}),
					NodeHash: n.CalSHA256Hash(bytes.Join([][]byte{nodes[i].NodeHash, nodes[i+1].NodeHash}, []byte{})),
				})
			}

			// Replace leaf node with upper level nodes.
			// Continue tree building until Root node is reached, i.e. len(nodes) == 1
			nodes = upperNodes

			// Print the upper node if printFlag is ON
			if printFlag == true {
				fmt.Printf("\nUpper Nodes: \n")
				for i := 0; i < len(nodes); i = i + 1 {
					fmt.Printf(" Hash[%d]	%x\n", i, nodes[i].NodeHash)
					fmt.Printf(" Data[%d]	%x\n", i, nodes[i].NodeData)
				}
			}

		} else {
			if printFlag == true {
				fmt.Printf("\n")
			}
			break
		}
	}
	RootNode = nodes[0]
	return RootNode
}
