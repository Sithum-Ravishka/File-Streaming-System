package main

import (
	"crypto/sha256"
	"fileChunk/fileSplit"
	"fmt"
)

type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

func calculateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func buildMerkleTree(data []string) *MerkleNode {

	if len(data) == 1 {
		return &MerkleNode{Hash: calculateHash(data[0]), Left: nil, Right: nil}
	}

	mid := len(data) / 2
	left := buildMerkleTree(data[:mid])
	right := buildMerkleTree(data[mid:])
	return &MerkleNode{Hash: calculateHash(left.Hash + right.Hash), Left: left, Right: right}
}

func printTree(node *MerkleNode, indent string) {
	if node != nil {
		fmt.Println(indent+"Hash:", node.Hash)
		if node.Left != nil {
			printTree(node.Left, indent+"  ")
		}
		if node.Right != nil {
			printTree(node.Right, indent+"  ")
		}
	}
}

const (
	chunkSize = int64(502400)
	inputFile = "data.jpg"
)

func main() {
	// Split the input file into chunks
	chunkNames, err := fileSplit.SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting file:", err)
		return
	}

	// fmt.Printf("Size: %v\n", name)

	root := buildMerkleTree(chunkNames)
	printTree(root, "")
	fmt.Println("Root Hash:", root.Hash)
}
