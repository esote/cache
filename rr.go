package cache

import (
	"encoding/binary"
	"io"
	"math/rand"
)

type rr struct {
	capacity int

	cache map[interface{}]int
	list  []*Pair

	e     chan<- Pair
	block bool

	r *rand.Rand
}

// NewRR constructs a new random-replacement cache. Items are evicted based on
// random indices read from rnd. If reading from this source fails, the cache
// will panic. If rnd is nil, "math/rand" will be used.
func NewRR(capacity int, rnd io.Reader) Cache {
	if capacity <= 0 {
		panic("rr: capacity <= 0")
	}

	rr := &rr{
		capacity: capacity,
		cache:    make(map[interface{}]int, capacity),
		list:     make([]*Pair, capacity),
	}

	if rnd == nil {
		rr.r = rand.New(rand.NewSource(1).(rand.Source64))
	} else {
		rr.r = rand.New(rand.Source64(&src{rnd}))
	}

	return rr
}

func (rr *rr) Get(key interface{}) (value interface{}, hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		value = rr.list[n].Value
	}
	return
}

func (rr *rr) Add(key, value interface{}) (hit bool) {
	if _, hit = rr.cache[key]; hit {
		return
	}

	item := &Pair{key, value}

	if len(rr.cache) >= rr.capacity {
		// Swap values
		n := rr.r.Intn(len(rr.cache))
		send(rr.e, rr.block, *rr.list[n])
		delete(rr.cache, rr.list[n].Key)
		rr.list[n] = item
		rr.cache[key] = n
	} else {
		rr.list[len(rr.cache)] = item
		rr.cache[key] = len(rr.cache)
	}

	return
}

func (rr *rr) Set(key, value interface{}) (hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		rr.list[n].Value = value
	}
	return
}

func (rr *rr) Delete(key interface{}) (hit bool) {
	var n int
	if n, hit = rr.cache[key]; hit {
		delete(rr.cache, key)
		rr.list[n], rr.list[len(rr.cache)] = rr.list[len(rr.cache)], nil
		rr.cache[rr.list[n].Key] = n
	}
	return
}

func (rr *rr) Clear() {
	rr.cache = make(map[interface{}]int, rr.capacity)
	rr.list = make([]*Pair, rr.capacity)
}

func (rr *rr) Len() int {
	return len(rr.cache)
}

func (rr *rr) Eviction(e chan<- Pair, block bool) {
	rr.e, rr.block = e, block
}

func (rr *rr) Dump() []Pair {
	pairs := make([]Pair, 0, len(rr.cache))
	for _, v := range rr.list {
		pairs = append(pairs, *v)
	}
	return pairs
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
