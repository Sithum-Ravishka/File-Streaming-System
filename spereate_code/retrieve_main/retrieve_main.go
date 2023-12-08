package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func RetrieveChunksAndVerify(directory string, outputFileName string) error {
	files, err := filepath.Glob(filepath.Join(directory, "data.jpg_chunk*"))
	if err != nil {
		return err
	}

	var chunkNames []string
	var hashValues []string

	hasher := sha256.New()

	for _, file := range files {
		chunkNames = append(chunkNames, file)

		chunkFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// Verify the hash of the chunk while copying to the output file
		_, err = io.Copy(io.MultiWriter(hasher, os.Stdout), chunkFile)
		if err != nil {
			return err
		}

		// Reset the hash for the next iteration
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))
		hashValues = append(hashValues, hashValue)
		hasher.Reset()
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for i, chunkName := range chunkNames {
		chunkFile, err := os.Open(chunkName)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		_, err = io.Copy(outputFile, chunkFile)
		if err != nil {
			return err
		}

		// Verify the hash of the chunk
		if hashValues[i] != fmt.Sprintf("%x", sha256.Sum256([]byte(chunkName))) {
			return fmt.Errorf("hash verification failed for chunk %s", chunkName)
		}
	}

	return nil
}

func main() {
	// Replace this value with the actual directory where chunks are stored
	chunkDirectory := "./"

	outputFileName := "restored_data.jpg"

	err := RetrieveChunksAndVerify(chunkDirectory, outputFileName)
	if err != nil {
		fmt.Println("Error retrieving and verifying chunks:", err)
		return
	}
}
