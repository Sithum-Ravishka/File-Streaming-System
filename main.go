package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"

	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
)

func SplitFile(inputFile string, chunkSize int64) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()

	chunkNames := make([]string, 0)

	for i := int64(0); i < fileSize; i += chunkSize {
		remaining := fileSize - i
		toCopy := chunkSize
		if remaining < chunkSize {
			toCopy = remaining
		}

		chunkName := fmt.Sprintf("%s_chunk%d", inputFile, i/chunkSize+1)
		chunkFile, err := os.Create(chunkName)
		if err != nil {
			return nil, err
		}

		// Copy the specified toCopy bytes from the original file to the chunk file
		_, err = io.CopyN(chunkFile, file, toCopy)
		if err != nil && err != io.EOF {
			chunkFile.Close()
			return nil, err
		}

		chunkFile.Close()
		chunkNames = append(chunkNames, chunkName)
	}

	return chunkNames, nil
}

// Sparse MT
func main() {

	const (
		chunkSize = int64(256000)
		inputFile = "test.png"
	)
	chunkNames, err := SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting and hashing file:", err)
		return
	}

	ctx := context.Background()

	// Tree storage
	store := memory.NewMemoryStorage()

	// Generate a new MerkleTree with 32 levels
	mt, _ := merkletree.NewMerkleTree(ctx, store, 32)

	for i, chunkName := range chunkNames {
		index := big.NewInt(int64(i)) // Use the correct index based on your application
		value, err := calculateHash(chunkName)
		if err != nil {
			fmt.Println("Error calculating hash for chunk:", err)
			return
		}
		mt.Add(ctx, index, value)

		fmt.Println(mt)
		fmt.Println("Root Key after adding chunk", i+1, ":", mt.Root().String())

	}

	// Proof of membership of a leaf with index 1
	proofExist, value, _ := mt.GenerateProof(ctx, big.NewInt(1), mt.Root())

	fmt.Println("Proof of membership:", proofExist.Existence)
	fmt.Println("Value corresponding to the queried index:", value)

	// Proof of non-membership of a leaf with index 4
	proofNotExist, _, _ := mt.GenerateProof(ctx, big.NewInt(10), mt.Root())

	fmt.Println("Proof of membership:", proofNotExist.Existence)

	// transform root from bytes array to json
	claimToMarshal, _ := json.Marshal(mt.Root())

	fmt.Println(string(claimToMarshal))
}

func calculateHash(chunkName string) (*big.Int, error) {
	file, err := os.Open(chunkName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Calculate hash (you might need to choose an appropriate hashing algorithm)
	// Here, we are using a simple checksum as an example
	hash := calculateChecksum(file)
	return hash, nil
}

func calculateChecksum(r io.Reader) *big.Int {
	// Replace this with your preferred hash calculation logic
	// For example, you can use a cryptographic hash library like SHA-256
	// Here, we're using a simple checksum for illustration purposes
	checksum := 0
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		for _, b := range buf[:n] {
			checksum += int(b)
		}
	}
	return big.NewInt(int64(checksum))
}
