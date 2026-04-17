package comm

import (
	"fmt"
	"net"
	"os"
)

// Define the Unix socket path
const socketPath = "/tmp/cloud_ramp_socket"

// Message types
const (
	MessageTypeInitialCode = 0 // Sending initial code to the container
	MessageTypeRequest     = 1 // Sending request (coordinator -> container)
	MessageTypeResponse    = 2 // Receiving response (container -> coordinator)
	MessageTypeError       = 5 // Sending/receiving error
)

// SendMessage sends a message to the container
func SendMessage(conn net.Conn, messageType byte, payload string) error {
	// Construct the message: first byte is the type, rest is the payload
	message := append([]byte{messageType}, []byte(payload)...)
	_, err := conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// HandleResponse processes responses from the container
func HandleResponse(data []byte) {
	if len(data) == 0 {
		fmt.Println("Received empty response")
		return
	}

	messageType := data[0]      // First byte is the message type
	payload := string(data[1:]) // Remaining bytes are the payload

	switch messageType {
	case MessageTypeResponse:
		fmt.Printf("Received response from container: %s\n", payload)
	case MessageTypeError:
		fmt.Printf("Received error from container: %s\n", payload)
	default:
		fmt.Printf("Unknown message type: %d, payload: %s\n", messageType, payload)
	}
}

// StartCoordinator starts the coordinator and communicates with the container
func StartCoordinator() error {
	// Remove the socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	// Create a Unix domain socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Coordinator listening on %s\n", socketPath)

	for {
		// Accept a connection from the container
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Container connected")

	// Example: Send initial code to the container
	initialCode := `
module.exports = {
  onMessage: (message) => {
    return "Processed: " + message;
  },
};
`
	if err := SendMessage(conn, MessageTypeInitialCode, initialCode); err != nil {
		fmt.Printf("Failed to send initial code: %v\n", err)
		return
	}
	fmt.Println("Sent initial code to container")

	// Example: Send a request to the container
	request := "Hello from coordinator!"
	if err := SendMessage(conn, MessageTypeRequest, request); err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	fmt.Println("Sent request to container")

	// Read response from the container
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	// Handle the response
	HandleResponse(buffer[:n])
}
