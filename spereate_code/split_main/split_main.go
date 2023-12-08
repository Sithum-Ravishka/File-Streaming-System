package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
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
		chunkNames = append(chunkNames, chunkFile.Name())

		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		hashValues = append(hashValues, hashValue)

		hasher.Reset()
	}

	return chunkNames, hashValues, nil
}

func main() {
	chunkSize := int64(500000)
	inputFile, err := os.Open("data.jpg")
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

	// Print chunk names and hash values
	fmt.Println("Chunk Names:", chunkNames)
	fmt.Println("Hash Values:", hashValues)
}
