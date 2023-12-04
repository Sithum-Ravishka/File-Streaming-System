package main

import (
	"fmt"
	"io"
	"os"
)

// SplitFile splits a file into chunks of a specified size
func SplitFile(inputFile string, chunkSize int64) ([]string, error) {

	// Open the input file for reading
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close() // Ensure the file is closed when the function exits

	// Get information about the file, including its size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size() // Get the size of the file in bytes

	chunkNames := make([]string, 0) // Create a slice to store the names of the created chunks

	// Loop through the file in chunks and create new chunk files
	for i := int64(0); i < fileSize; i += chunkSize {
		// Create a unique name for each chunk file based on the input file name and chunk index
		chunkName := fmt.Sprintf("%s_chunk%d", inputFile, i/chunkSize+1)

		// Create a new file for the chunk
		chunkFile, err := os.Create(chunkName)
		if err != nil {
			return nil, err
		}

		// Copy the specified chunkSize bytes from the original file to the chunk file
		_, err = io.CopyN(chunkFile, file, chunkSize)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunkFile.Close() // Close the chunk file

		// Add the name of the created chunk file to the slice
		chunkNames = append(chunkNames, chunkName)
	}

	// Return the names of all created chunk files
	return chunkNames, nil
}

func main() {
	inputFile := "data.jpg"
	chunkSize := int64(102400) // Set your desired chunk size in bytes

	// Split the file into chunks using the SplitFile function
	_, err := SplitFile(inputFile, chunkSize)
	if err != nil {
		fmt.Println("Error splitting file:", err)
		return
	}

	fmt.Println("File split successfully!")
}
