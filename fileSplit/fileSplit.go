package fileSplit

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// SplitFile splits a given file into chunks of a specified size and returns
// the names of the generated chunks and their corresponding hash values.
func SplitFile(inputFile *os.File, chunkSize int64) ([]string, []string, error) {
	// Ensure the input file is closed when the function completes
	defer inputFile.Close()

	// Retrieve information about the input file
	fileInfo, err := inputFile.Stat()
	if err != nil {
		return nil, nil, err
	}

	// Get the total size of the input file
	fileSize := fileInfo.Size()
	var chunkNames []string
	var hashValues []string

	// Use SHA-256 as the hashing algorithm
	hasher := sha256.New()

	// Loop through the file, creating chunks of the specified size
	for i := int64(0); i < fileSize; i += chunkSize {
		// Create a new chunk file with a name based on the index
		chunkFile, err := os.Create(fmt.Sprintf("%v_chunk%d", inputFile.Name(), i/chunkSize+1))
		if err != nil {
			return nil, nil, err
		}

		// Use a MultiWriter to simultaneously write to the chunk file and update the hash
		multiWriter := io.MultiWriter(chunkFile, hasher)

		// Copy the next chunkSize bytes from the input file to the chunk file and update the hash
		_, err = io.CopyN(multiWriter, inputFile, chunkSize)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		// Close the chunk file
		chunkFile.Close()

		// Calculate the hash value of the chunk
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		// Use the hash value as the new name for the chunk file
		hashedFileName := fmt.Sprintf("%s", hashValue)

		// Rename the chunk file with its hash value
		err = os.Rename(chunkFile.Name(), hashedFileName)
		if err != nil {
			return nil, nil, err
		}

		// Append the hashed file name and hash value to the respective slices
		chunkNames = append(chunkNames, hashedFileName)
		hashValues = append(hashValues, hashValue)

		// Reset the hash for the next iteration
		hasher.Reset()
	}

	// Return the names of the generated chunks and their corresponding hash values
	return chunkNames, hashValues, nil
}
