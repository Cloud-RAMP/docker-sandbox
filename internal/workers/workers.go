package workers

import (
	"bufio"
	"fmt"
	"os/exec"
	"path"
	"sync"

	"github.com/Cloud-RAMP/docker-sandbox/internal/config"
)

func StartWorkers() error {
	var wg sync.WaitGroup
	for i := range config.NUM_CONTAINERS {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			containerName := fmt.Sprintf("container-sandbox-%d", workerID)

			// Mount the directory, not the socket file itself
			cmd := exec.Command(
				"docker", "run", "-d",
				"--name", containerName,
				"-p", fmt.Sprintf("%d:%d", config.BASE_PORT+workerID, config.BASE_PORT+workerID),
				"container-sandbox", fmt.Sprintf("%d", config.BASE_PORT+workerID),
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error starting container %d: %v\nOutput: %s\n", workerID, err, string(output))
				return
			}

			fmt.Printf("Container %d started successfully\n", workerID)
		}(i)
	}

	wg.Wait()
	return nil
}

func StartWorkersLocal() error {
	var wg sync.WaitGroup

	for i := range config.NUM_CONTAINERS {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			scriptPath := path.Join(config.ROOT_DIR_PATH, "container", "index.js")
			port := fmt.Sprintf("%d", config.BASE_PORT+workerID)

			cmd := exec.Command("node", scriptPath, port)

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Printf("Error creating stdout pipe for worker %d: %v\n", workerID, err)
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				fmt.Printf("Error creating stderr pipe for worker %d: %v\n", workerID, err)
				return
			}

			if err := cmd.Start(); err != nil {
				fmt.Printf("Error starting node worker %d: %v\n", workerID, err)
				return
			}

			// Stream stdout
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					fmt.Printf("[worker %d stdout] %s\n", workerID, scanner.Text())
				}
			}()
			// Stream stderr
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					fmt.Printf("[worker %d stderr] %s\n", workerID, scanner.Text())
				}
			}()

			if err := cmd.Wait(); err != nil {
				fmt.Printf("Node worker %d exited with error: %v\n", workerID, err)
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func StopWorkers() {
	for i := range config.NUM_CONTAINERS {
		containerName := fmt.Sprintf("container-sandbox-%d", i)
		removeCmd := exec.Command("docker", "rm", "-f", containerName)
		if output, err := removeCmd.CombinedOutput(); err == nil {
			fmt.Printf("Removed existing container %s: %s", containerName, string(output))
		}
	}
}
