package fileSplit

import (
	"fmt"
	"io"
	"os"
)

// SplitFile splits a given file into chunks of a specified size and returns
// the names of the generated chunks.
func SplitFile(inputFile *os.File, chunkSize int64) ([]string, error) {
	// Ensure the input file is closed when the function completes
	defer inputFile.Close()

	// Retrieve information about the input file
	fileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, err
	}

	// Get the total size of the input file
	fileSize := fileInfo.Size()
	var chunkNames []string

	// Loop through the file, creating chunks of the specified size
	for i := int64(0); i < fileSize; i += chunkSize {
		// Create a new chunk file with a name based on the index
		chunkFile, err := os.Create(fmt.Sprintf("%v_chunk%d", inputFile.Name(), i/chunkSize+1))
		if err != nil {
			return nil, err
		}

		// Copy the next chunkSize bytes from the input file to the chunk file
		_, err = io.CopyN(chunkFile, inputFile, chunkSize)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Close the chunk file
		chunkFile.Close()

		// Append the chunk file name to the slice
		chunkNames = append(chunkNames, chunkFile.Name())
	}

	// Return the names of the generated chunks
	return chunkNames, nil
}
