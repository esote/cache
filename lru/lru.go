package lru

import (
	"container/list"
)

// LRU implements a least-recently-used cache. This is identical to the MRU
// cache except pruning begins at the back. LRU satisfies the Cache interface.
//
// Items which are accessed least-recently are evicted first. All operations are
// O(1).
type LRU struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List
}

type pair struct {
	k, v interface{}
}

// New constructs a new LRU cache. Panics if capacity <= 0.
func New(capacity int) *LRU {
	if capacity <= 0 {
		panic("lru: capacity <= 0")
	}

	return &LRU{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

// Get value in the LRU cache.
func (lru *LRU) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.list.MoveToFront(item)
		value = item.Value.(*pair).v
	}
	return
}

// Add value to the LRU cache.
func (lru *LRU) Add(key, value interface{}) (hit bool) {
	if _, hit = lru.cache[key]; hit {
		return
	}

	if len(lru.cache) >= lru.capacity {
		if item := lru.list.Back(); item != nil {
			lru.remove(item)
		}
	}

	lru.cache[key] = lru.list.PushFront(&pair{key, value})
	return
}

// Set value in the LRU cache. The item is moved to the front of the cache.
func (lru *LRU) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.list.MoveToFront(item)
		item.Value.(*pair).v = value
	}
	return
}

// Delete value from the LRU cache. Returns whether the delete hit.
func (lru *LRU) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.remove(item)
	}
	return
}

// Clear all values in the LRU cache.
func (lru *LRU) Clear() {
	lru.cache = make(map[interface{}]*list.Element, lru.capacity)
	lru.list = lru.list.Init()
}

// Len returns the number of items in the LRU cache.
func (lru *LRU) Len() int {
	return len(lru.cache)
}

func (lru *LRU) remove(item *list.Element) {
	delete(lru.cache, item.Value.(*pair).k)
	lru.list.Remove(item)
}
