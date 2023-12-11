package main

import (
	"context"
	"encoding/json"
	"fileChunk/fileRetrieve"
	"fileChunk/fileSplit"
	"fmt"
	"math/big"
	"os"

	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
)

func main() {
	// Add chunk size
	chunkSize := int64(1900000)

	// File load in here
	inputFile, err := os.Open("data.jpg")

	// Check if there is an error
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	// Used to close inputFile after function run is complete.
	defer inputFile.Close()

	// Call the file split function
	chunkNames, hashValues, err := fileSplit.SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting and hashing file:", err)
		return
	}

	ctx := context.Background()
	store := memory.NewMemoryStorage()
	mt, _ := merkletree.NewMerkleTree(ctx, store, 32)

	// Get index and value of from the ChunkNames slice
	// Add to the merkle tree to chunk file
	for index, value := range chunkNames {
		mt.Add(ctx, big.NewInt(int64(index)), big.NewInt(0)) // Need to adjust the second parameter based on our use case need
		fmt.Println(ctx, index, value)

		// Print the merkle root
		fmt.Println(mt.Root())

		// Proof of membership for each chunk
		proofExist, _, _ := mt.GenerateProof(ctx, big.NewInt(int64(index)), mt.Root())
		fmt.Printf("Proof of membership for chunk %d: %v\n", index, proofExist.Existence)

		user := "B"
		if user == "A" {
			err := newFunction(proofExist, chunkNames, hashValues, "restored_data.jpg")
			if err != nil {
				fmt.Println("Error retrieving and verifying chunks:", err)
				return
			}
		} else {
			fmt.Println("Invalid User Authentication")
			return
		}
	}

	// Proof of non-membership for a non-existing chunk (e.g., index 100)
	nonExistingIndex := big.NewInt(100)
	proofNotExist, _, _ := mt.GenerateProof(ctx, nonExistingIndex, mt.Root())
	fmt.Printf("Proof of non-membership for chunk %d: %v\n", nonExistingIndex.Int64(), proofNotExist.Existence)

	// Get root-hash in the merkle tree
	claimToMarshal, _ := json.Marshal(mt.Root())
	fmt.Printf("%s", claimToMarshal)
}

func newFunction(proofExist *merkletree.Proof, chunkNames []string, hashValues []string, outputFileName string) error {
	// Check file proof with chunk file data for retrieve
	if proofExist.Existence {
		return fileRetrieve.RetrieveChunksAndVerify(chunkNames, hashValues, outputFileName)
	}
	return fmt.Errorf("proof of non-membership received")
}

// func userFunction(user, chunkNames []string, hashValues []string, outputFileName string) error {
// 	// Check file proof with chunk file data for retrieve
// 	user = "A"

// 	if user == "A" {
// 		return fileRetrieve.RetrieveChunksAndVerify(chunkNames, hashValues, outputFileName)
// 	}
// 	return fmt.Errorf("proof of non-membership received")
// }
