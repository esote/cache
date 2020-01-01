package cache

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/esote/cache/fifo"
	"github.com/esote/cache/lfu"
	"github.com/esote/cache/lifo"
	"github.com/esote/cache/lru"
	"github.com/esote/cache/mru"
	"github.com/esote/cache/rr"
)

type cache struct {
	name  string
	cache Cache
}

func freshCaches(capacity int) []cache {
	return []cache{
		{"FIFO", fifo.New(capacity)},
		{"LFU", lfu.New(capacity)},
		{"LIFO", lifo.New(capacity)},
		{"LRU", lru.New(capacity)},
		{"MRU", mru.New(capacity)},
		{"RR", rr.New(capacity, nil)},
	}
}

func BenchmarkAddFrequencies(b *testing.B) {
	dists := []struct {
		name string
		f    func([]uint16)
	}{
		{"Uniform", uniformBuf},
		{"Normal", normalBuf},
		{"Zipf", zipfBuf},
	}

	buf := make([]uint16, math.MaxUint16/8)

	for _, d := range dists {
		caches := freshCaches(1 << 10)
		for _, c := range caches {
			b.Run(c.name+"_Add_"+d.name, func(b *testing.B) {
				benchmarkAdd(b, c.cache, buf, d.f)
			})
		}
		fmt.Println()
	}

	for _, d := range dists {
		caches := freshCaches(1 << 10)
		for _, c := range caches {
			b.Run(c.name+"_Get_"+d.name, func(b *testing.B) {
				benchmarkGet(b, c.cache, buf, d.f)
			})
		}
		fmt.Println()
	}
}

func benchmarkAdd(b *testing.B, c Cache, buf []uint16, dist func([]uint16)) {
	var curs int
	dist(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(buf[curs], nil)
		curs++
		if curs == len(buf) {
			b.StopTimer()
			dist(buf)
			curs = 0
			b.StartTimer()
		}
	}
}

func benchmarkGet(b *testing.B, c Cache, buf []uint16, dist func([]uint16)) {
	var curs int
	in := make([]uint16, 1<<10)
	dist(in)
	for i := range in {
		c.Add(in[i], nil)
	}
	dist(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(buf[curs])
		curs++
		if curs == len(buf) {
			b.StopTimer()
			dist(buf)
			curs = 0
			b.StartTimer()
		}
	}
}

const (
	sd  = math.MaxUint16 / 8
	med = math.MaxUint16 / 2
	s   = 1.2
)

var zipf = rand.NewZipf(rand.New(rand.NewSource(1)), s, 1, math.MaxUint16-1)

func uniformBuf(buf []uint16) {
	for i := range buf {
		buf[i] = uint16(rand.Intn(math.MaxUint16 + 1))
	}
}

func normUint16() uint16 {
	r := float64(-1)
	for ; r < 0 || r > math.MaxUint16; r = rand.NormFloat64()*sd + med {
	}
	return uint16(r)
}

func normalBuf(buf []uint16) {
	for i := range buf {
		buf[i] = normUint16()
	}
}

func zipfBuf(buf []uint16) {
	for i := range buf {
		buf[i] = uint16(zipf.Uint64())
	}
}

func TestCoherent(t *testing.T) {
	caches := freshCaches(3)

	for _, c := range caches {
		if err := testCoherent(c.cache); err != nil {
			t.Fatal(err)
		}
	}
}

func testCoherent(c Cache) error {
	if c.Add(1, 1) {
		return errors.New("add 1->1")
	}

	if value, hit := c.Get(1); !hit || value != 1 {
		return fmt.Errorf("get 1, got %d", value)
	}

	if c.Add(2, 2) {
		return errors.New("add 2->2")
	}

	if c.Add(3, 3) {
		return errors.New("add 3->3")
	}

	if value, hit := c.Get(3); !hit || value != 3 {
		return fmt.Errorf("get 3, got %d", value)
	}

	if !c.Delete(2) {
		return errors.New("delete 3")
	}

	if value, hit := c.Get(2); hit {
		return fmt.Errorf("get 2, got %d", value)
	}

	if !c.Set(1, "A") {
		return errors.New("set 1->A")
	}

	if value, hit := c.Get(1); !hit || value != "A" {
		return fmt.Errorf("get 1, got %d", value)
	}

	if !c.Delete(1) {
		return errors.New("delete 1")
	}

	if l := c.Len(); l != 1 {
		return fmt.Errorf("len %d", l)
	}

	if !c.Add(3, "B") {
		return errors.New("add 3->B")
	}

	c.Clear()

	if l := c.Len(); l != 0 {
		return errors.New("len")
	}

	return nil
}
