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

func main() {
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

// startServer function to handle incoming connections
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
	go startServer(DefaultServerPort) // Start server in a goroutine

	for {
		conn, err := net.Dial("tcp", target)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			fmt.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connected to peer at", target)

		for {
			fmt.Println("Enter file name to send or type 'retrieve' to retrieve a file:")
			command, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				fmt.Println("Error reading command:", err)
				break
			}

			command = strings.TrimSpace(command)

			switch command {
			case "retrieve":
				handleFileRetrieve(conn, userID)
			default:
				handleFileSend(conn, command, userID)
			}
		}

		// If the connection is lost, retry after a delay
		fmt.Println("Connection lost. Retrying in 5 seconds...")
		conn.Close()
		time.Sleep(5 * time.Second)
	}
}

// Modify handleFileSend function to split the file into chunks of 128 KB and send them
func handleFileSend(conn net.Conn, fileName, userID string) {
	fmt.Println("File transfer initiated. Sending file:", fileName)

	// Create a folder for the user if it doesn't exist
	userFolder := fmt.Sprintf("%s", userID)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		os.Mkdir(userFolder, os.ModeDir)
	}

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Ensure the input file is closed when the function completes
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
	mt, _ := merkletree.NewMerkleTree(context.Background(), store, 32)

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
		defer chunkFile.Close()

		// Write the chunk data to the file
		_, err = chunkFile.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error writing chunk to file:", err)
			return
		}

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
		defer chunkFile.Close()

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
	fmt.Println("File retrieval initiated. Waiting for file name...")

	// Receive the protocol and userID from the peer
	protocol, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading protocol:", err)
		return
	}
	protocol = strings.TrimSpace(protocol)

	if protocol != ProtocolSendFile {
		fmt.Println("Invalid protocol. Expected SEND_FILE.")
		return
	}

	senderUserID, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading sender UserID:", err)
		return
	}
	senderUserID = strings.TrimSpace(senderUserID)

	// Ensure the folder for the user exists
	userFolder := fmt.Sprintf("%s", userID)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		os.Mkdir(userFolder, os.ModeDir)
	}

	// Receive chunks and reconstruct the file
	for {
		chunkFilePath, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading chunk file path:", err)
			return
		}
		chunkFilePath = strings.TrimSpace(chunkFilePath)

		// If the peer signals the end of the file transfer
		if chunkFilePath == "MerkleRoot:" {
			// Receive and display the Merkle Root
			merkleRoot, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Error reading Merkle Root:", err)
				return
			}
			merkleRoot = strings.TrimSpace(merkleRoot)
			fmt.Printf("Received Merkle Root: %s\n", merkleRoot)
			break
		}

		// Receive and save the chunk
		chunkFilePath = fmt.Sprintf("%s/%s", userFolder, chunkFilePath)
		receiveChunkAndSave(conn, chunkFilePath)
	}

	// Reconstruct data and save it as a JPG file
	retrieveData(userID)

	fmt.Println("File retrieval completed.")
}

// Function to retrieve data and save it as a JPG file
func retrieveData(userID string) {
	// Open the file for writing
	file, err := os.Create(fmt.Sprintf("%s/data.jpg", userID))
	if err != nil {
		fmt.Println("Error creating data file:", err)
		return
	}
	defer file.Close()

	// Read each chunk file and write its content to the data file
	for i := 1; ; i++ {
		chunkFilePath := fmt.Sprintf("%s/chunk%d", userID, i)
		chunkFile, err := os.Open(chunkFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				// No more chunks
				break
			}
			fmt.Println("Error opening chunk file:", err)
			return
		}
		defer chunkFile.Close()

		// Copy the content of the chunk file to the data file
		_, err = io.Copy(file, chunkFile)
		if err != nil {
			fmt.Println("Error writing chunk to data file:", err)
			return
		}
	}

	fmt.Println("Data retrieved and saved as data.jpg.")
}

// Function to receive a chunk and save it to a file
func receiveChunkAndSave(conn net.Conn, filePath string) {
	// Open the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating chunk file:", err)
		return
	}
	defer file.Close()

	// Receive and write the chunk data to the file
	_, err = io.Copy(file, conn)
	if err != nil {
		fmt.Println("Error receiving chunk:", err)
		return
	}

	fmt.Println("Chunk received successfully:", filePath)
}

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
