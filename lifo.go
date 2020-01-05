package cache

import "container/list"

type lifo struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List

	e     chan<- Pair
	block bool
}

// NewLIFO constructs a new last-in first-out cache. Items most-recently added
// are evicted first. This is identical to the FIFO cache except pruning begins
// at the front.
func NewLIFO(capacity int) Cache {
	if capacity <= 0 {
		panic("lifo: capacity <= 0")
	}

	return &lifo{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

func (lifo *lifo) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		value = item.Value.(*Pair).Value
	}
	return
}

func (lifo *lifo) Add(key, value interface{}) (hit bool) {
	if _, hit = lifo.cache[key]; hit {
		return
	}

	if len(lifo.cache) >= lifo.capacity {
		if item := lifo.list.Front(); item != nil {
			send(lifo.e, lifo.block, *item.Value.(*Pair))
			lifo.remove(item)
		}
	}

	lifo.cache[key] = lifo.list.PushFront(&Pair{key, value})
	return
}

func (lifo *lifo) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		lifo.list.MoveToFront(item)
		item.Value.(*Pair).Value = value
	}
	return
}

func (lifo *lifo) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lifo.cache[key]; hit {
		lifo.remove(item)
	}
	return
}

func (lifo *lifo) Clear() {
	lifo.cache = make(map[interface{}]*list.Element, lifo.capacity)
	lifo.list = lifo.list.Init()
}

func (lifo *lifo) Len() int {
	return len(lifo.cache)
}

func (lifo *lifo) Eviction(e chan<- Pair, block bool) {
	lifo.e, lifo.block = e, block
}

func (lifo *lifo) Dump() []Pair {
	pairs := make([]Pair, 0, len(lifo.cache))
	for _, v := range lifo.cache {
		pairs = append(pairs, *v.Value.(*Pair))
	}
	return pairs
}

func (lifo *lifo) remove(item *list.Element) {
	delete(lifo.cache, item.Value.(*Pair).Key)
	lifo.list.Remove(item)
}
