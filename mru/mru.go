package mru

import (
	"container/list"
)

// MRU implements a most-recently-used cache. This is identical to LRU cache
// except pruning begins at the front. MRU satisfies the Cache interface.
//
// Items which are accessed most-recently are evicted first. All operations are
// O(1).
type MRU struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List
}

type pair struct {
	k, v interface{}
}

// New constructs a new MRU cache. Panics if capacity <= 0.
func New(capacity int) *MRU {
	if capacity <= 0 {
		panic("mru: capacity <= 0")
	}

	return &MRU{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

// Get value in the MRU cache.
func (mru *MRU) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.list.MoveToFront(item)
		value = item.Value.(*pair).v
	}
	return
}

// Add value to the MRU cache.
func (mru *MRU) Add(key, value interface{}) (hit bool) {
	if _, hit = mru.cache[key]; hit {
		return
	}

	if len(mru.cache) >= mru.capacity {
		if item := mru.list.Front(); item != nil {
			mru.remove(item)
		}
	}

	mru.cache[key] = mru.list.PushFront(&pair{key, value})
	return
}

// Set value in the MRU cache.
func (mru *MRU) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.list.MoveToFront(item)
		item.Value.(*pair).v = value
	}
	return
}

// Delete value from the MRU cache. Returns whether the delete hit.
func (mru *MRU) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.remove(item)
	}
	return
}

// Clear all values in the MRU cache.
func (mru *MRU) Clear() {
	mru.cache = make(map[interface{}]*list.Element, mru.capacity)
	mru.list = mru.list.Init()
}

// Len returns the number of items in the MRU cache.
func (mru *MRU) Len() int {
	return len(mru.cache)
}

func (mru *MRU) remove(item *list.Element) {
	delete(mru.cache, item.Value.(*pair).k)
	mru.list.Remove(item)
}
