package cache_test

import (
	"github.com/esote/cache"
	"github.com/esote/cache/fifo"
	"github.com/esote/cache/lfu"
)

func ExampleSegmented() {
	// Segmented cache of a low-priority ("probationary") FIFO cache with a
	// high-priority ("protected") LFU cache.
	s := cache.NewSegmented(fifo.New(3), lfu.New(3))

	// 1 and 2 are added to the FIFO cache.
	s.Add(1, "A")
	s.Add(2, "B")

	// 1 is promoted to the LFU cache (removed from FIFO, added to LFU).
	s.Get(1)

	// 3 and 4 are added to the FIFO cache.
	s.Add(3, "C")
	s.Add(4, "D")

	// 5 is added to the FIFO cache, and 2 (the oldest element in the
	// full FIFO cache) is evicted.
	s.Add(5, "E")
}
