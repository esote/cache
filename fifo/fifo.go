package fifo

import "container/list"

// FIFO implements a first-in first-out cache. This is identical to the LIFO
// cache except pruning begins at the back. FIFO satisfies the Cache interface.
//
// Items least-recently added are evicted first. All operations are O(1).
type FIFO struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List
}

type pair struct {
	k, v interface{}
}

// New constructs a new FIFO cache. Panics if capacity <= 0.
func New(capacity int) *FIFO {
	if capacity <= 0 {
		panic("fifo: capacity <= 0")
	}

	return &FIFO{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

// Get value in the FIFO cache.
func (fifo *FIFO) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		value = item.Value.(*pair).v
	}
	return
}

// Add value to the FIFO cache.
func (fifo *FIFO) Add(key, value interface{}) (hit bool) {
	if _, hit = fifo.cache[key]; hit {
		return
	}

	if len(fifo.cache) >= fifo.capacity {
		if item := fifo.list.Back(); item != nil {
			fifo.remove(item)
		}
	}

	fifo.cache[key] = fifo.list.PushFront(&pair{key, value})
	return
}

// Set value in the FIFO cache. The item is moved to the front of the cache.
func (fifo *FIFO) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		fifo.list.MoveToFront(item)
		item.Value.(*pair).v = value
	}
	return
}

// Delete value from the FIFO cache. Returns whether the delete hit.
func (fifo *FIFO) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		fifo.remove(item)
	}
	return
}

// Clear all values in the FIFO cache.
func (fifo *FIFO) Clear() {
	fifo.cache = make(map[interface{}]*list.Element, fifo.capacity)
	fifo.list = fifo.list.Init()
}

// Len returns the number of items in the FIFO cache.
func (fifo *FIFO) Len() int {
	return len(fifo.cache)
}

func (fifo *FIFO) remove(item *list.Element) {
	delete(fifo.cache, item.Value.(*pair).k)
	fifo.list.Remove(item)
}
