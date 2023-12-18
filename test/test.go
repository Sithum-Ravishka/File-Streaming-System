package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
)

const (
	ProtocolSendFile    = "SEND_FILE"
	ProtocolRequestFile = "REQUEST_FILE"
	DefaultServerPort   = ":8080"
	DefaultClientTarget = "localhost:8080"
)

func kkk() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <mode> [port/target]")
		fmt.Println("Modes:")
		fmt.Println("  server: Start as a server")
		fmt.Println("  client: Start as a client")
		return
	}

	mode := strings.ToLower(os.Args[1])

	switch mode {
	case "server":
		port := DefaultServerPort
		if len(os.Args) == 3 {
			port = ":" + os.Args[2]
		}
		startServer(port)
	case "client":
		target := DefaultClientTarget
		if len(os.Args) == 3 {
			target = os.Args[2]
		}
		connectToPeer(target, generateUserID())
	default:
		fmt.Println("Invalid mode. Use 'server' or 'client'.")
	}
}

func startServer(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started. Listening on", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Generate a unique userID for the connected peer
	userID := generateUserID()
	fmt.Println("New connection from", conn.RemoteAddr(), "with UserID:", userID)

	for {
		fmt.Println("Enter file name to send or type 'retrieve' to retrieve a file:")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			return
		}

		command = strings.TrimSpace(command)

		switch command {
		case "retrieve":
			handleFileRetrieve(conn, userID)
		default:
			handleFileSend(conn, command, userID)
		}
	}
}

// Modify connectToPeer function to send chunks using the fileSplit package
func connectToPeer(target string, userID string) {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to peer at", target)

	for {
		fmt.Println("Enter file name to send or type 'retrieve' to retrieve a file:")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			return
		}

		command = strings.TrimSpace(command)

		switch command {
		case "retrieve":
			handleFileRetrieve(conn, userID)
		default:
			handleFileSend(conn, command, userID)
		}
	}
}

// Modify handleFileSend function to split the file into chunks of 128 KB and send them
func handleFileSend(conn net.Conn, fileName, userID string) {
	fmt.Println("File transfer initiated. Sending file:", fileName)

	// Create a folder for the user if it doesn't exist
	userFolder := fmt.Sprintf("%s", userID)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		// Create the directory with read-write-execute permissions for the owner
		err := os.Mkdir(userFolder, 0700)
		if err != nil {
			fmt.Println("Error creating user folder:", err)
			return
		}
	}

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Retrieve information about the input file
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	// Get the total size of the input file
	fileSize := fileInfo.Size()

	// Create Merkle Tree storage
	store := memory.NewMemoryStorage()
	mt, err := merkletree.NewMerkleTree(context.Background(), store, 32)
	if err != nil {
		fmt.Println("Error initializing Merkle tree:", err)
		return
	}

	// Send the protocol and userID to the peer
	conn.Write([]byte(ProtocolSendFile + "\n"))
	conn.Write([]byte(userID + "\n"))

	// Buffer to read and send chunks
	buffer := make([]byte, 1024)

	for i := int64(0); i < fileSize; i += 1024 {
		// Read a chunk from the file
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// Create a new chunk file with a name based on the index in the specified directory
		chunkFilePath := fmt.Sprintf("%s/chunk%d", userFolder, i/(1024)+1)
		chunkFile, err := os.Create(chunkFilePath)
		if err != nil {
			fmt.Println("Error creating chunk file:", err)
			return
		}

		// Write the chunk data to the file
		_, err = chunkFile.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error writing chunk to file:", err)
			return
		}

		// Close the chunk file after writing
		chunkFile.Close()

		// Add the chunk file to the Merkle Tree
		index := big.NewInt(i / (128 * 1024))
		mt.Add(context.Background(), index, hash(chunkFile))

		// Send the chunk name to the peer
		conn.Write([]byte(chunkFilePath + "\n"))

		// Open the chunk file for reading
		chunkFile, err = os.Open(chunkFilePath)
		if err != nil {
			fmt.Println("Error opening chunk file:", err)
			return
		}

		// Send the chunk to the peer
		io.Copy(conn, chunkFile)

		fmt.Println("Chunk sent successfully:", chunkFilePath)
		fmt.Printf("merkle chunk hash%s\n", mt.Root().Hex())
	}

	// Send the Merkle Root to the peer
	conn.Write([]byte("MerkleRoot:\n" + mt.Root().String() + "\n"))
	fmt.Printf("merkle root:%s\n", mt.Root().Hex())
	fmt.Println("File sent successfully.")
}

// Hash function to generate hash of a file
func hash(file *os.File) *big.Int {
	// Implement your hash generation logic here
	// For example, you can use cryptographic hash functions like SHA-256
	// For simplicity, let's assume a basic hash using the file name
	hash := new(big.Int)
	hash.SetString(strings.ReplaceAll(file.Name(), "/", ""), 10)
	return hash
}

// Modify handleFileRetrieve function to receive chunks and reconstruct the file
func handleFileRetrieve(conn net.Conn, userID string) {
	fmt.Println("Enter file name to retrieve:")
	fileName, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading file name:", err)
		return
	}

	fileName = strings.TrimSpace(fileName)

	conn.Write([]byte(ProtocolRequestFile + "\n"))
	conn.Write([]byte(userID + "\n")) // Send the userID to the server
	conn.Write([]byte(fileName + "\n"))

	// Receive the chunks
	for {
		chunkName, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading chunk name:", err)
			return
		}

		chunkName = strings.TrimSpace(chunkName)

		// Check if the transfer is complete
		if chunkName == "TRANSFER_COMPLETE" {
			fmt.Println("File retrieval complete.")
			break
		}

		// Create the chunk file for writing
		chunkFile, err := os.Create(chunkName)
		if err != nil {
			fmt.Println("Error creating chunk file:", err)
			return
		}
		defer chunkFile.Close()

		// Receive the chunk from the peer
		io.Copy(chunkFile, conn)

		fmt.Println("Chunk received successfully:", chunkName)
	}
}

// ReconstructFile reconstructs a file from its chunks
// func ReconstructFile(fileName string) error {
//  // ... (implementation of file reconstruction, combining chunks into the original file)
// }

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
