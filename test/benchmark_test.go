package test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Cloud-RAMP/docker-sandbox/internal/comm"
	"github.com/Cloud-RAMP/docker-sandbox/internal/config"
	"github.com/Cloud-RAMP/docker-sandbox/internal/workers"
)

func setup(t testing.TB) {
	t.Helper()

	// workers.StopWorkers()

	go func() {
		err := comm.StartCoordinator()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	workers.StartWorkers()

	time.Sleep(1500 * time.Millisecond)
}

// BenchmarkSimpleSingleModule-8   	   10000	    229567 ns/op	    2071 B/op	       4 allocs/op
// Send all requests to one module
func BenchmarkSimpleSingleModule(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for b.Loop() {
		req := comm.MockContainerRequest{
			Payload: "",
			Module:  "0",
		}

		if err := comm.HandleRequest(req); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimpleModuleEviction-8   	    1474	    794842 ns/op	    5086 B/op	      19 allocs/op
func BenchmarkSimpleModuleEviction(b *testing.B) {
	setup(b)

	i := 0
	b.ResetTimer()
	for b.Loop() {
		req := comm.MockContainerRequest{
			Payload: "",
			Module:  fmt.Sprintf("%d", i%(config.NUM_CONTAINERS)),
		}

		if err := comm.HandleRequest(req); err != nil {
			b.Fatal(err)
		}

		i++
	}
}

// BenchmarkZipf-8   	    1807	    641591 ns/op	    4147 B/op	      14 allocs/op
func BenchmarkZipf(b *testing.B) {
	setup(b)

	rng := rand.New(rand.NewSource(42)) // deterministic benchmark distribution
	zipf := rand.NewZipf(rng, 1.2, 1, uint64(config.NUM_CONTAINERS))

	b.ResetTimer()
	for b.Loop() {
		idx := zipf.Uint64()
		req := comm.MockContainerRequest{
			Payload: "",
			Module:  fmt.Sprintf("%d", idx),
		}

		if err := comm.HandleRequest(req); err != nil {
			b.Fatal(err)
		}
	}
}
