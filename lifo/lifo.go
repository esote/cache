// Package lifo provides a last-in first-out cache.
package lifo

import "container/list"

// LIFO implements a last-in first-out cache. This is identical to the FIFO
// cache except pruning begins at the front. LIFO satisfies the Cache interface.
//
// Items most-recently added are evicted first. All operations are O(1).
type LIFO struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List
}

type pair struct {
	k, v interface{}
}

// New constructs a new LIFO cache. Panics if capacity <= 0.
func New(capacity int) *LIFO {
	if capacity <= 0 {
		panic("lifo: capacity <= 0")
	}

	return &LIFO{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

// Get value in LIFO cache.
func (lifo *LIFO) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		value = item.Value.(*pair).v
	}
	return
}

// Add value to the LIFO cache.
func (lifo *LIFO) Add(key, value interface{}) (hit bool) {
	if _, hit = lifo.cache[key]; hit {
		return
	}

	if len(lifo.cache) >= lifo.capacity {
		if item := lifo.list.Front(); item != nil {
			lifo.remove(item)
		}
	}

	lifo.cache[key] = lifo.list.PushFront(&pair{key, value})
	return
}

// Set value in the LIFO cache. The item is moved to the front of the cache.
func (lifo *LIFO) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		lifo.list.MoveToFront(item)
		item.Value.(*pair).v = value
	}
	return
}

// Delete value from the LRU cache. Returns whether the delete hit.
func (lifo *LIFO) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		lifo.remove(item)
	}
	return
}

// Clear all values in the LIFO cache.
func (lifo *LIFO) Clear() {
	lifo.cache = make(map[interface{}]*list.Element, lifo.capacity)
	lifo.list = lifo.list.Init()
}

// Len returns the number of items in the LIFO cache.
func (lifo *LIFO) Len() int {
	return len(lifo.cache)
}

func (lifo *LIFO) remove(item *list.Element) {
	delete(lifo.cache, item.Value.(*pair).k)
	lifo.list.Remove(item)
}
