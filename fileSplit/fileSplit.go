package fileSplit

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// SplitAndHashFile splits a file into chunks, calculates the SHA256 hash for each chunk,
// and returns the names of the created chunks along with their corresponding hash values.
func SplitAndHashFile(inputFile string, chunkSize int64) ([]string, []string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	fileSize := fileInfo.Size()

	chunkNames := make([]string, 0)
	hashValues := make([]string, 0)

	hasher := sha256.New()

	for i := int64(0); i < fileSize; i += chunkSize {
		chunkName := fmt.Sprintf("%s_chunk%d", inputFile, i/chunkSize+1)
		chunkFile, err := os.Create(chunkName)
		if err != nil {
			return nil, nil, err
		}

		// Create a multi-writer to both write to the file and calculate the hash
		multiWriter := io.MultiWriter(chunkFile, hasher)

		// Copy the specified chunkSize bytes from the original file to the chunk file
		_, err = io.CopyN(multiWriter, file, chunkSize)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		chunkFile.Close()

		// Add the name of the created chunk file to the slice
		chunkNames = append(chunkNames, chunkName)

		// Add the hash value of the chunk to the slice
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		hashValues = append(hashValues, hashValue)

		// Reset the hash for the next iteration
		hasher.Reset()
	}

	return chunkNames, hashValues, nil
}
