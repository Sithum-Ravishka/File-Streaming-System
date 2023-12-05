package main

import (
	"fileChunk/fileRetrieve"
	"fileChunk/fileSplit"
	"fmt"
)

const (
	chunkSize = int64(102400)
	inputFile = "data.jpg"
)

func main() {

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
