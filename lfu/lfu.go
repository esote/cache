// Package lfu provides a least-frequently-used cache.
package lfu

import (
	"container/list"
	"math"
)

// LFU implements a least-frequently-used cache. LFU satisfies the Cache
// interface.
//
// Items which are accessed the least are evicted first. All operations are
// O(1). This is an implementation of the O(1) eviction scheme given by Shah,
// Mitra, and Matani in http://dhruvbird.com/lfu.pdf. Item frequency is limited
// to 2^(64) - 1.
type LFU struct {
	capacity int

	cache map[interface{}]*pair
	list  *list.List
}

type pair struct {
	k, v interface{}
	el   *list.Element
}

type header struct {
	entries   map[*pair]bool
	frequency uint64
}

// New constructs a new LFU cache. Panics if capacity <= 0.
func New(capacity int) *LFU {
	if capacity <= 0 {
		panic("lfu: capacity <= 0")
	}

	return &LFU{
		capacity: capacity,
		cache:    make(map[interface{}]*pair, capacity),
		list:     list.New(),
	}
}

// Get value in the LFU cache.
func (lfu *LFU) Get(key interface{}) (value interface{}, hit bool) {
	var item *pair
	if item, hit = lfu.cache[key]; hit {
		lfu.increment(item)
		value = item.v
	}
	return
}

// Add value to the LFU cache.
func (lfu *LFU) Add(key, value interface{}) (hit bool) {
	if _, hit = lfu.cache[key]; hit {
		return
	}

	if len(lfu.cache) >= lfu.capacity {
		if item := lfu.list.Front(); item != nil {
			entry := item.Value.(*header).any()
			delete(lfu.cache, entry.k)
			lfu.remove(item, entry)
		}
	}

	item := &pair{
		k: key,
		v: value,
	}

	lfu.cache[key] = item
	lfu.increment(item)
	return
}

// Set value in the LFU cache. The item's frequency is incremented.
func (lfu *LFU) Set(key, value interface{}) (hit bool) {
	var item *pair
	if item, hit = lfu.cache[key]; hit {
		item.v = value
		lfu.increment(item)
	}
	return
}

// Delete value from the LFU cache. Returns whether the delete hit.
func (lfu *LFU) Delete(key interface{}) (hit bool) {
	var item *pair
	if item, hit = lfu.cache[key]; hit {
		delete(lfu.cache, item.k)
		lfu.remove(item.el, item)
	}
	return
}

// Clear all values in the LFU cache.
func (lfu *LFU) Clear() {
	lfu.cache = make(map[interface{}]*pair, lfu.capacity)
	lfu.list = lfu.list.Init()
}

// Len returns the number of items in the LFU cache.
func (lfu *LFU) Len() int {
	return len(lfu.cache)
}

// Select any element from header entries. Should only be called when entries
// exist.
func (hdr *header) any() *pair {
	for entry := range hdr.entries {
		return entry
	}

	panic("lfu: hdr: no entries")
}

func (lfu *LFU) increment(item *pair) {
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
			entries:   make(map[*pair]bool),
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

func (lfu *LFU) remove(el *list.Element, item *pair) {
	hdr := el.Value.(*header)
	delete(hdr.entries, item)
	if len(hdr.entries) == 0 {
		lfu.list.Remove(el)
	}
}
