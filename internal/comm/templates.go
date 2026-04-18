package comm

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/Cloud-RAMP/docker-sandbox/internal/config"
)

// Message types
const (
	MessageTypeInitialCode       = 0 // Sending initial code to the container
	MessageTypeRequest           = 1 // Sending request (coordinator -> container)
	MessageTypeResponse          = 2 // Receiving response (container -> coordinator)
	MessageTypeContainerRequest  = 3 // Receive request (container -> coordinator)
	MessageTypeContainerResponse = 4 // Send response (coordinator -> container)
	MessageTypeError             = 5 // Sending/receiving error
)

// SendMessage sends a message to the container
func SendMessage(conn net.Conn, messageType byte, payload string) error {
	payloadBytes := []byte(payload)
	length := uint32(len(payloadBytes))

	// Construct: [type (1 byte)][length (4 bytes)][payload]
	message := make([]byte, 1+4+len(payloadBytes))
	message[0] = messageType
	binary.BigEndian.PutUint32(message[1:5], length)
	copy(message[5:], payloadBytes)

	if conn == nil {
		fmt.Println("Conn nil when sending message!")
		return fmt.Errorf("bad")
	}

	_, err := conn.Write(message)
	if err != nil {
		return err
	}
	return nil
}

// Send initial code to a container
func sendCode(conn net.Conn) error {
	codePath := path.Join(config.ROOT_DIR_PATH, "resources/userCode.js")
	initialCodeBytes, err := os.ReadFile(codePath)
	if err != nil {
		fmt.Printf("Failed to read initial code from %s: %v\n", codePath, err)
		return err
	}
	initialCode := string(initialCodeBytes)

	if err := SendMessage(conn, MessageTypeInitialCode, initialCode); err != nil {
		return err
	}

	// Read code ACK from container
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		return err
	}

	return nil
}

func sendRequest(conn net.Conn, request string) error {
	if err := SendMessage(conn, MessageTypeRequest, request); err != nil {
		return err
	}

	// Read response from the container
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	if buffer[0] != byte(3) {
		fmt.Println("Invalid response from external request", string(buffer))
		return fmt.Errorf("")
	}

	// send message back, response to external request
	if err := SendMessage(conn, MessageTypeContainerResponse, ""); err != nil {
		return err
	}

	buffer = make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		return err
	}
	if buffer[0] != byte(6) {
		fmt.Println("Invalid done message", string(buffer))
		return fmt.Errorf("")
	}

	return nil
}
