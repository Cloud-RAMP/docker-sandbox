package comm

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Cloud-RAMP/docker-sandbox/internal/config"
)

type MockContainerRequest struct {
	Payload   string
	Module    string
	Timestamp time.Time
}

type Container struct {
	Port         int
	Mu           *sync.RWMutex
	LoadedModule string
	LastRequest  time.Time
	DataChan     chan MockContainerRequest
	Conn         net.Conn
}

var containers []Container

func init() {
	for i := range config.NUM_CONTAINERS {
		containers = append(containers, Container{
			Port:         config.BASE_PORT + i,
			Mu:           &sync.RWMutex{},
			LoadedModule: "",
			DataChan:     make(chan MockContainerRequest, 10000),
		})
	}
}

func StartCoordinator() error {
	var wg sync.WaitGroup

	for i := range config.NUM_CONTAINERS {
		container := &containers[i]
		addr := fmt.Sprintf("127.0.0.1:%d", container.Port)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to create TCP socket for container %d: %w", i, err)
		}
		defer listener.Close()

		fmt.Printf("Coordinator listening on %s\n", addr)

		wg.Add(1)
		go func(container *Container, listener net.Listener) {
			defer wg.Done()

			container.Conn = nil
			for {
				conn, err := listener.Accept()
				if err != nil {
					fmt.Printf("Failed to accept connection on %s: %v\n", addr, err)
					continue
				}

				container.Mu.Lock()
				// Close any existing connection before handling the new one
				if container.Conn != nil {
					close(container.DataChan)
					container.DataChan = make(chan MockContainerRequest, 10000)
					container.Conn.Close()
				}
				container.Conn = conn
				container.Mu.Unlock()

				go handleConnection(container)
			}
		}(container, listener)
	}

	wg.Wait()
	return nil
}

// External facing function to be called by whoever is feeding reqeusts into the system
//
// In this case, it will just be the benchmark tester
func HandleRequest(req MockContainerRequest) error {
	var container *Container = nil

	for _, c := range containers {
		c.Mu.RLock()
		if c.LoadedModule == req.Module {
			container = &c
			c.Mu.RUnlock()
			break
		}
		c.Mu.RUnlock()
	}

	if container == nil {
		var lruContainer *Container = nil
		var oldestTime time.Time

		for i, c := range containers {
			c.Mu.RLock()
			if lruContainer == nil || containers[i].LastRequest.Before(oldestTime) {
				lruContainer = &containers[i]
				oldestTime = containers[i].LastRequest
			}
			c.Mu.RUnlock()
		}

		// Replace LRU container with this module
		if lruContainer != nil {
			lruContainer.Mu.Lock()
			lruContainer.LoadedModule = req.Module
			container = lruContainer
			sendCode(lruContainer.Conn)
			lruContainer.Mu.Unlock()
		}
	}

	// Send the request to the container
	container.DataChan <- req
	return nil
}

// Once the socket connection with the container has opened, read requests and handle them
func handleConnection(container *Container) {
	defer container.Conn.Close()

	for req := range container.DataChan {
		container.Mu.Lock()
		if err := sendRequest(container.Conn, req.Payload); err != nil {
			container.Mu.Unlock()
			return
		}
		container.Mu.Unlock()
	}
}
