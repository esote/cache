package rr

import (
	"encoding/binary"
	"io"
	"math/rand"
)

// RR implements a random-replacement cache. RR satisfies the Cache interface.
//
// Items are evicted from random indices. All operations are O(1).
type RR struct {
	capacity int

	cache map[interface{}]int
	list  []*pair

	r *rand.Rand
}

type pair struct {
	k, v interface{}
}

// New constructs a new RR cache. Panics if capacity <= 0. The random reader
// will be used as the source for randomness. If reading from this source fails,
// RR will panic. If the random reader is nil, "math/rand" will be used.
func New(capacity int, random io.Reader) *RR {
	if capacity <= 0 {
		panic("rr: capacity <= 0")
	}

	rr := &RR{
		capacity: capacity,
		cache:    make(map[interface{}]int, capacity),
		list:     make([]*pair, capacity),
	}

	if random == nil {
		rr.r = rand.New(rand.NewSource(1).(rand.Source64))
	} else {
		rr.r = rand.New(rand.Source64(&src{random}))
	}

	return rr
}

// Get value in the RR cache.
func (rr *RR) Get(key interface{}) (value interface{}, hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		value = rr.list[n].v
	}
	return
}

// Add value to the RR cache.
func (rr *RR) Add(key, value interface{}) (hit bool) {
	if _, hit = rr.cache[key]; hit {
		return
	}

	item := &pair{key, value}

	if len(rr.cache) >= rr.capacity {
		// Swap values
		n := rr.r.Intn(len(rr.cache))
		delete(rr.cache, rr.list[n].k)
		rr.list[n] = item
		rr.cache[key] = n
	} else {
		rr.list[len(rr.cache)] = item
		rr.cache[key] = len(rr.cache)
	}

	return
}

// Set value in the RR cache.
func (rr *RR) Set(key, value interface{}) (hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		rr.list[n].v = value
	}
	return
}

// Delete value from the RR cache. Returns whether the delete hit.
func (rr *RR) Delete(key interface{}) (hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		delete(rr.cache, key)
		rr.list[n], rr.list[len(rr.cache)] = rr.list[len(rr.cache)], nil
		rr.cache[rr.list[n].k] = n
	}
	return
}

// Clear all values in the cache.
func (rr *RR) Clear() {
	rr.cache = make(map[interface{}]int, rr.capacity)
	rr.list = make([]*pair, rr.capacity)
}

// Len returns the number of items in the RR cache.
func (rr *RR) Len() int {
	return len(rr.cache)
}

type src struct {
	r io.Reader
}

func (s *src) Int63() int64 {
	return int64(s.Uint64() & (1<<63 - 1))
}

func (s *src) Seed(seed int64) {
	// nop
	_ = seed
}

func (s *src) Uint64() uint64 {
	var buf [8]byte
	if _, err := io.ReadFull(s.r, buf[:]); err != nil {
		panic("rr: read random failed: " + err.Error())
	}
	return binary.LittleEndian.Uint64(buf[:])
}
