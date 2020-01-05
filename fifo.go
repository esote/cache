package cache

import "container/list"

type fifo struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List

	e     chan<- Pair
	block bool
}

// NewFIFO constructs a new first-in first-out cache. Items least-recently added
// are evicted first. This is identical to the LIFO cache except pruning begins
// at the back. All operations are O(1).
func NewFIFO(capacity int) Cache {
	if capacity <= 0 {
		panic("fifo: capacity <= 0")
	}

	return &fifo{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

func (fifo *fifo) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		value = item.Value.(*Pair).Value
	}
	return
}

func (fifo *fifo) Add(key, value interface{}) (hit bool) {
	if _, hit = fifo.cache[key]; hit {
		return
	}

	if len(fifo.cache) >= fifo.capacity {
		if item := fifo.list.Back(); item != nil {
			send(fifo.e, fifo.block, *item.Value.(*Pair))
			fifo.remove(item)
		}
	}

	fifo.cache[key] = fifo.list.PushFront(&Pair{key, value})
	return
}

func (fifo *fifo) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		fifo.list.MoveToFront(item)
		item.Value.(*Pair).Value = value
	}
	return
}

func (fifo *fifo) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = fifo.cache[key]; hit {
		fifo.remove(item)
	}
	return
}

func (fifo *fifo) Clear() {
	fifo.cache = make(map[interface{}]*list.Element, fifo.capacity)
	fifo.list = fifo.list.Init()
}

func (fifo *fifo) Len() int {
	return len(fifo.cache)
}

func (fifo *fifo) Eviction(e chan<- Pair, block bool) {
	fifo.e, fifo.block = e, block
}

func (fifo *fifo) Dump() []Pair {
	pairs := make([]Pair, 0, len(fifo.cache))
	for _, v := range fifo.cache {
		pairs = append(pairs, *v.Value.(*Pair))
	}
	return pairs
}

func (fifo *fifo) remove(item *list.Element) {
	delete(fifo.cache, item.Value.(*Pair).Key)
	fifo.list.Remove(item)
}
