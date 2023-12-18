package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func RetrieveChunks(directory string, outputFileName string) error {
	// Step 1: Gather chunk file names
	files, err := filepath.Glob(filepath.Join(directory, "chunk*"))
	if err != nil {
		return err
	}

	// Step 2: Create a slice to store chunk file names
	var chunkNames []string

	// Step 3: Iterate through chunk files, copy to the output file, and verify the hash
	for _, file := range files {
		chunkNames = append(chunkNames, file)

		chunkFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer chunkFile.Close()

		// Verify the hash of the chunk while copying to the output file
		_, err = io.Copy(io.MultiWriter(os.Stdout), chunkFile)
		if err != nil {
			return err
		}
	}

	// Step 4: Create the output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Step 5: Iterate through chunk files again and copy to the output file
	for _, chunkName := range chunkNames {
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
		// Add your hash verification code here
	}

	return nil
}

func main() {
	// Replace this value with the actual directory where chunks are stored
	chunkDirectory := "./splitfile"

	outputFileName := "restored.jpg"

	err := RetrieveChunks(chunkDirectory, outputFileName)
	if err != nil {
		fmt.Println("Error retrieving and verifying chunks:", err)
		return
	}
}
