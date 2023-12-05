package main

import (
	"fileChunk/fileRetrieve"
	"fileChunk/fileSplit"
	"fmt"
)

const (
	chunkSize = int64(502400)
	inputFile = "data.jpg"
)

func main() {
	// Split the input file into chunks and calculate hash values
	chunkNames, hashValues, err := fileSplit.SplitAndHashFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting and hashing file:", err)
		return
	}

	// Set the output file name for the reconstructed file
	outputFileName := "new.jpg"

	// Retrieve and concatenate the chunks to reconstruct the original file,
	// verifying the integrity of each chunk using the calculated hash values
	err = fileRetrieve.RetrieveChunksAndVerify(chunkNames, hashValues, outputFileName)
	if err != nil {
		fmt.Println("Error retrieving and verifying chunks:", err)
		return
	}

	// Print hash values for each chunk
	fmt.Println("Hash values for each chunk:")
	for i, hashValue := range hashValues {
		fmt.Printf("Chunk %d: %s\n", i+1, hashValue)
	}

	// Print a success message if the process completes without errors
	fmt.Println("File split, hashed, retrieved, and verified successfully!")
}
