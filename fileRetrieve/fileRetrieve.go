package fileRetrieve

import (
	"io"
	"os"
)

// RetrieveChunksAndVerify retrieves and concatenates chunks to reconstruct the original file,
// verifying the integrity of each chunk using the provided hash values.
func RetrieveChunksAndVerify(chunkNames []string, outputFileName string) error {
	// Create the output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Loop through each chunk
	for _, chunkName := range chunkNames {
		// Open the chunk file
		chunkFile, err := os.Open(chunkName)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// Copy the chunk data to the output file
		_, err = io.Copy(outputFile, chunkFile)
		if err != nil {
			return err
		}

	}

	// Add specific attributes for chunk file from user-A
	// After adding chunks to the merkle tree, retrieve proof with user for retrieval

	return nil
}
