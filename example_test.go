package cache_test

import (
	"fmt"
	"sync"

	"github.com/esote/cache"
)

func ExampleNewSegmented() {
	// Segmented cache of a low-priority ("probationary") FIFO cache with a
	// high-priority ("protected") LFU cache.
	s := cache.NewSegmented(cache.NewFIFO(2), cache.NewLFU(2))

	// Register an eviction channel for the segmented cache, which prints
	// out evicted values.
	evicted := make(chan cache.Pair, 1)
	s.Eviction(evicted, true)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for pair := range evicted {
			fmt.Printf("<%v, %v>\n", pair.Key, pair.Value)
		}
	}()

	// 1 and 2 are added to the FIFO cache.
	s.Add(1, "A")
	s.Add(2, "B")

	// 1 is promoted to the LFU cache (removed from FIFO, added to LFU).
	s.Get(1)

	// 2 is promoted to LFU, and (per LFU eviction policy) are seen as more
	// frequently accessed than 1.
	s.Get(2)
	s.Get(2)

	// 3 and 4 are added to the FIFO cache.
	s.Add(3, "C")
	s.Add(4, "D")

	// 3 is promoted to LFU, and 1 (being the least-frequently-used item in
	// the LFU cache) is demoted to the FIFO cache.
	s.Get(3)

	// 5 is added to the FIFO cache, and 2 (the oldest element in the
	// full FIFO cache) is evicted.
	s.Add(5, "E")

	_ = s.Close()
	close(evicted)
	wg.Wait()
	// Output: <4, D>
}
