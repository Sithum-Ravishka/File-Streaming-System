package fileSplit

import (
	"fmt"
	"io"
	"os"
)

// SplitFile splits a file into chunks and returns the names of the created chunks.
func SplitFile(inputFile string, chunkSize int64) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()

	chunkNames := make([]string, 0)

	for i := int64(0); i < fileSize; i += chunkSize {
		remaining := fileSize - i
		toCopy := chunkSize
		if remaining < chunkSize {
			toCopy = remaining
		}

		chunkName := fmt.Sprintf("%s_chunk%d", inputFile, i/chunkSize+1)
		chunkFile, err := os.Create(chunkName)
		if err != nil {
			return nil, err
		}

		// Copy the specified toCopy bytes from the original file to the chunk file
		_, err = io.CopyN(chunkFile, file, toCopy)
		if err != nil && err != io.EOF {
			chunkFile.Close()
			return nil, err
		}

		chunkFile.Close()
		chunkNames = append(chunkNames, chunkName)
	}

	return chunkNames, nil
}
