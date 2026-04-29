package test

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sync"
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

var setupOnce sync.Once

// BenchmarkParellelSingleModule-8   	  137491	    297432 ns/op	    2137 B/op	       5 allocs/op
func BenchmarkParellelSingleModule(b *testing.B) {
	setupOnce.Do(func() { setup(b) })

	b.SetParallelism(8) // 8x GOMAXPROCS goroutines
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := comm.MockContainerRequest{
				Payload: "",
				Module:  "0",
			}

			if err := comm.HandleRequest(req); err != nil {
				b.Fatal(err)
			}
		}
	})
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

func BenchmarkDockerLatencyPercentiles(b *testing.B) {
	setupOnce.Do(func() { setup(b) })

	req := comm.MockContainerRequest{
		Payload: "",
		Module:  "0",
	}

	// warmup
	for range 10 {
		if err := comm.HandleRequest(req); err != nil {
			b.Fatalf("Warmup failed: %v", err)
		}
	}

	latencies := make([]time.Duration, 0, b.N)
	var mu sync.Mutex

	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()
			err := comm.HandleRequest(req)
			elapsed := time.Since(start)

			if err != nil {
				b.Errorf("Failed to handle request: %v", err)
				return
			}

			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()
		}
	})

	b.StopTimer()

	slices.Sort(latencies)

	n := len(latencies)
	if n == 0 {
		b.Fatal("No latencies recorded")
	}

	p50 := latencies[n*50/100]
	p95 := latencies[n*95/100]
	p99 := latencies[n*99/100]
	pMax := latencies[n-1]

	b.ReportMetric(float64(p50.Microseconds()), "p50_us")
	b.ReportMetric(float64(p95.Microseconds()), "p95_us")
	b.ReportMetric(float64(p99.Microseconds()), "p99_us")
	b.ReportMetric(float64(pMax.Microseconds()), "pMax_us")

	b.Logf("p50:  %v", p50)
	b.Logf("p95:  %v", p95)
	b.Logf("p99:  %v", p99)
	b.Logf("pMax: %v", pMax)
}
