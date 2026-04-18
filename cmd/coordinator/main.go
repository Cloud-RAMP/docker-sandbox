package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Cloud-RAMP/docker-sandbox/internal/comm"
	"github.com/Cloud-RAMP/docker-sandbox/internal/config"
	"github.com/Cloud-RAMP/docker-sandbox/internal/workers"
)

func main() {
	var wg sync.WaitGroup

	workers.StopWorkers()

	wg.Add(1)
	go func() {
		if err := comm.StartCoordinator(); err != nil {
			fmt.Println(err.Error())
		}
		wg.Done()
	}()

	time.Sleep(500 * time.Millisecond)
	go workers.StartWorkers()
	time.Sleep(1000 * time.Millisecond)

	fmt.Println()
	for i := range 5 {
		req := comm.MockContainerRequest{
			Payload: fmt.Sprintf("%d", i),
			Module:  fmt.Sprintf("%d", i%config.NUM_CONTAINERS),
		}

		if err := comm.HandleRequest(req); err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	wg.Wait()
}
