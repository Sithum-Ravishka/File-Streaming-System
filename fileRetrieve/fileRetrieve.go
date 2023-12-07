package fileRetrieve

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// RetrieveChunksAndVerify retrieves and concatenates chunks to reconstruct the original file,
// verifying the integrity of each chunk using the provided hash values.
func RetrieveChunksAndVerify(chunkNames []string, hashValues []string, outputFileName string) error {
	outputFile, err := os.Create(outputFileName)

	// Error handle
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Use SHA-256 as the hashing algorithum
	hasher := sha256.New()

	for i, chunkName := range chunkNames {
		chunkFile, err := os.Open(chunkName)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// Create a multi-reader to both read from the file and calculate the hash
		multiReader := io.TeeReader(chunkFile, hasher)

		_, err = io.Copy(outputFile, multiReader)
		if err != nil {
			return err
		}

		// Verify the hash of the chunk
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		if hashValue != hashValues[i] {
			return fmt.Errorf("hash verification failed for chunk %s", chunkName)
		}

		// Reset the hash for the next iteration
		hasher.Reset()
	}

	return nil
}
