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

func SplitFile(inputFile *os.File, chunkSize int64) ([]string, []string, error) {
	defer inputFile.Close()

	fileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, nil, err
	}

	fileSize := fileInfo.Size()
	var chunkNames []string
	var hashValues []string

	hasher := sha256.New()

	for i := int64(0); i < fileSize; i += chunkSize {
		chunkFile, err := os.Create(fmt.Sprintf("%v_chunk%d", inputFile.Name(), i/chunkSize+1))
		if err != nil {
			return nil, nil, err
		}

		multiWriter := io.MultiWriter(chunkFile, hasher)

		_, err = io.CopyN(multiWriter, inputFile, chunkSize)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		chunkFile.Close()

		// Calculate hash value
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		hashedFileName := fmt.Sprintf("%s", hashValue)

		// Rename the chunk file with its hash value
		err = os.Rename(chunkFile.Name(), hashedFileName)
		if err != nil {
			return nil, nil, err
		}

		chunkNames = append(chunkNames, hashedFileName)
		hashValues = append(hashValues, hashValue)

		hasher.Reset()
	}

	return chunkNames, hashValues, nil
}

func RetrieveChunksAndVerify(chunkNames []string, hashValues []string, outputFileName string) error {
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	hasher := sha256.New()

	for i, chunkName := range chunkNames {
		chunkFile, err := os.Open(chunkName)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// Create a multi-reader to both read from the file and calculate the hash
		multiReader := io.TeeReader(chunkFile, hasher)

		_, err = io.Copy(outputFile, multiReader)
		if err != nil {
			return err
		}

		// Verify the hash of the chunk
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		if hashValue != hashValues[i] {
			return fmt.Errorf("hash verification failed for chunk %s", chunkName)
		}

		// Reset the hash for the next iteration
		hasher.Reset()
	}

	return nil
}

func main() {
	chunkSize := int64(500000)
	inputFile, err := os.Open("test.png")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer inputFile.Close()

	chunkNames, hashValues, err := SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting and hashing file:", err)
		return
	}

	ctx := context.Background()
	store := memory.NewMemoryStorage()
	mt, _ := merkletree.NewMerkleTree(ctx, store, 32)

	for index, value := range chunkNames {
		mt.Add(ctx, big.NewInt(int64(index)), big.NewInt(0)) // Need to adjust the second parameter based on our use case need
		fmt.Println(ctx, index, value)

		fmt.Println(mt.Root())

		// Proof of membership for each chunk
		proofExist, _, _ := mt.GenerateProof(ctx, big.NewInt(int64(index)), mt.Root())
		fmt.Printf("Proof of membership for chunk %d: %v\n", index, proofExist.Existence)

		err := newFunction(proofExist, chunkNames, hashValues, "restored_data.jpg")
		if err != nil {
			fmt.Println("Error retrieving and verifying chunks:", err)
			return
		}
	}

	// Proof of non-membership for a non-existing chunk (e.g., index 100)
	nonExistingIndex := big.NewInt(100)
	proofNotExist, _, _ := mt.GenerateProof(ctx, nonExistingIndex, mt.Root())
	fmt.Printf("Proof of non-membership for chunk %d: %v\n", nonExistingIndex.Int64(), proofNotExist.Existence)

	claimToMarshal, _ := json.Marshal(mt.Root())
	fmt.Println(string(claimToMarshal))
}

func newFunction(proofExist *merkletree.Proof, chunkNames []string, hashValues []string, outputFileName string) error {
	if proofExist.Existence {
		return RetrieveChunksAndVerify(chunkNames, hashValues, outputFileName)
	}
	return fmt.Errorf("proof of non-membership received")
}
