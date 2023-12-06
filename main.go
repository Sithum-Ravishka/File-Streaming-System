package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"

	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
)

func SplitFile(inputFile *os.File, chunkSize int64) ([]int64, error) {
	defer inputFile.Close()

	fileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()

	chunkSizes := make([]int64, 0)
	hasher := sha256.New()

	for i := int64(0); i < fileSize; i += chunkSize {
		chunkFile, err := os.Create(fmt.Sprintf("%v_chunk%d", inputFile, i/chunkSize+1))
		if err != nil {
			return nil, err
		}

		multiWriter := io.MultiWriter(chunkFile, hasher)

		_, err = io.CopyN(multiWriter, inputFile, chunkSize)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunkFile.Close()
		chunkSizes = append(chunkSizes, chunkSize)

		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		newChunkName := hashValue
		err = os.Rename(chunkFile.Name(), newChunkName)
		if err != nil {
			return nil, err
		}
		hasher.Reset()
	}

	return chunkSizes, nil
}

func main() {
	chunkSize := int64(500000)
	inputFile, err := os.Open("data.jpg")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	chunkSizes, err := SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting and hashing file:", err)
		return
	}

	ctx := context.Background()
	store := memory.NewMemoryStorage()
	mt, _ := merkletree.NewMerkleTree(ctx, store, 32)

	for index, value := range chunkSizes {
		mt.Add(ctx, big.NewInt(int64(index)), big.NewInt(value))
		fmt.Println(ctx, index, value)

		// Proof of membership for each chunk
		proofExist, _, _ := mt.GenerateProof(ctx, big.NewInt(int64(index)), mt.Root())
		fmt.Printf("Proof of membership for chunk %d: %v\n", index, proofExist.Existence)
	}

	// Proof of non-membership for a non-existing chunk (e.g., index 100)
	nonExistingIndex := big.NewInt(100)
	proofNotExist, _, _ := mt.GenerateProof(ctx, nonExistingIndex, mt.Root())
	fmt.Printf("Proof of non-membership for chunk %d: %v\n", nonExistingIndex.Int64(), proofNotExist.Existence)

	claimToMarshal, _ := json.Marshal(mt.Root())
	fmt.Println(string(claimToMarshal))
}
