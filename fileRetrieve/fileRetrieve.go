package fileRetrieve

import (
	"io"
	"os"
)

// RetrieveChunks retrieves and concatenates chunks to reconstruct the original file
func RetrieveChunks(chunkNames []string, outputFileName string) error {
	// Open the output file for writing
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close() // Defer closing the output file until the surrounding function returns

	// Iterate through each chunk name
	for _, chunkName := range chunkNames {
		// Open each chunk file for reading
		chunkFile, err := os.Open(chunkName)
		if err != nil {
			return err
		}
		defer chunkFile.Close() // Defer closing the chunk file until the surrounding function returns

		// Copy the content of each chunk to the output file
		_, err = io.Copy(outputFile, chunkFile)
		if err != nil {
			return err
		}
	}

	return nil
}
