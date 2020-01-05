package cache

import (
	"container/list"
	"math"
)

type lfu struct {
	capacity int

	cache map[interface{}]*elPair
	list  *list.List

	e     chan<- Pair
	block bool
}

type elPair struct {
	Pair
	el *list.Element
}

type header struct {
	entries   map[*elPair]bool
	frequency uint64
}

// NewLFU constructs a new least-frequently-used cache. Items which are accessed
// the least are evicted first. This is an implementation of the O(1) eviction
// scheme given by Shah, Mitra, and Matani in http://dhruvbird.com/lfu.pdf. Item
// frequency is limited to 2^(64) - 1.
func NewLFU(capacity int) Cache {
	if capacity <= 0 {
		panic("lfu: capacity <= 0")
	}

	return &lfu{
		capacity: capacity,
		cache:    make(map[interface{}]*elPair, capacity),
		list:     list.New(),
	}
}

func (lfu *lfu) Get(key interface{}) (value interface{}, hit bool) {
	var item *elPair
	if item, hit = lfu.cache[key]; hit {
		lfu.increment(item)
		value = item.Value
	}
	return
}

func (lfu *lfu) Add(key, value interface{}) (hit bool) {
	if _, hit = lfu.cache[key]; hit {
		return
	}

	if len(lfu.cache) >= lfu.capacity {
		if item := lfu.list.Front(); item != nil {
			entry := item.Value.(*header).any()
			send(lfu.e, lfu.block, entry.Pair)
			delete(lfu.cache, entry.Key)
			lfu.remove(item, entry)
		}
	}

	item := &elPair{
		Pair: Pair{Key: key, Value: value},
		el:   nil,
	}

	lfu.cache[key] = item
	lfu.increment(item)
	return
}

func (lfu *lfu) Set(key, value interface{}) (hit bool) {
	var item *elPair
	if item, hit = lfu.cache[key]; hit {
		item.Value = value
		lfu.increment(item)
	}
	return
}

func (lfu *lfu) Delete(key interface{}) (hit bool) {
	var item *elPair
	if item, hit = lfu.cache[key]; hit {
		delete(lfu.cache, item.Key)
		lfu.remove(item.el, item)
	}
	return
}

func (lfu *lfu) Clear() {
	lfu.cache = make(map[interface{}]*elPair, lfu.capacity)
	lfu.list = lfu.list.Init()
}

func (lfu *lfu) Len() int {
	return len(lfu.cache)
}

func (lfu *lfu) Eviction(e chan<- Pair, block bool) {
	lfu.e, lfu.block = e, block
}

func (lfu *lfu) Dump() []Pair {
	pairs := make([]Pair, 0, len(lfu.cache))
	for _, v := range lfu.cache {
		pairs = append(pairs, v.Pair)
	}
	return pairs
}

// Select any element from header entries. Should only be called when entries
// exist.
func (hdr *header) any() *elPair {
	for entry := range hdr.entries {
		return entry
	}

	panic("lfu: hdr: no entries")
}

func (lfu *lfu) increment(item *elPair) {
	var frequency uint64
	var next *list.Element

	current := item.el
	if current == nil {
		next = lfu.list.Front()
	} else {
		frequency = current.Value.(*header).frequency
		next = current.Next()
	}

	if frequency != math.MaxUint64 {
		frequency++
	}

	if next == nil || next.Value.(*header).frequency != frequency {
		hdr := &header{
			frequency: frequency,
			entries:   make(map[*elPair]bool),
		}

		if current == nil {
			next = lfu.list.PushFront(hdr)
		} else {
			next = lfu.list.InsertAfter(hdr, current)
		}
	}

	item.el = next
	next.Value.(*header).entries[item] = true

	if current != nil {
		lfu.remove(current, item)
	}
}

func (lfu *lfu) remove(el *list.Element, item *elPair) {
	hdr := el.Value.(*header)
	delete(hdr.entries, item)
	if len(hdr.entries) == 0 {
		lfu.list.Remove(el)
	}
}
