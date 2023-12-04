package main

import (
	"fileChunk/fileRetrieve"
	"fileChunk/fileSplit"
	"fmt"
)

func main() {
	// Set the input file name
	inputFile := "data.jpg"
	// Set the desired chunk size in bytes
	chunkSize := int64(102400)

	// Split the input file into chunks
	chunkNames, err := fileSplit.SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting file:", err)
		return
	}

	// Set the output file name for the reconstructed file
	outputFileName := "new.jpg"

	// Retrieve and concatenate the chunks to reconstruct the original file
	err = fileRetrieve.RetrieveChunks(chunkNames, outputFileName)
	if err != nil {
		fmt.Println("Error retrieving chunks:", err)
		return
	}

	// Print a success message if the process completes without errors
	fmt.Println("File split and retrieved successfully!")
}
