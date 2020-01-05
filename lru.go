package cache

import (
	"container/list"
)

type lru struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List

	e     chan<- Pair
	block bool
}

// NewLRU constructs a new least-recently-used cache. Items which are accessed
// least-frequently are evicted first. This is identical to the MRU cache except
// pruning begins at the back.
func NewLRU(capacity int) Cache {
	if capacity <= 0 {
		panic("lru: capacity <= 0")
	}

	return &lru{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

func (lru *lru) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.list.MoveToFront(item)
		value = item.Value.(*Pair).Value
	}
	return
}

func (lru *lru) Add(key, value interface{}) (hit bool) {
	if _, hit = lru.cache[key]; hit {
		return
	}

	if len(lru.cache) >= lru.capacity {
		if item := lru.list.Back(); item != nil {
			send(lru.e, lru.block, *item.Value.(*Pair))
			lru.remove(item)
		}
	}

	lru.cache[key] = lru.list.PushFront(&Pair{key, value})
	return
}

func (lru *lru) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.list.MoveToFront(item)
		item.Value.(*Pair).Value = value
	}
	return
}

func (lru *lru) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = lru.cache[key]; hit {
		lru.remove(item)
	}
	return
}

func (lru *lru) Clear() {
	lru.cache = make(map[interface{}]*list.Element, lru.capacity)
	lru.list = lru.list.Init()
}

func (lru *lru) Len() int {
	return len(lru.cache)
}

func (lru *lru) Eviction(e chan<- Pair, block bool) {
	lru.e, lru.block = e, block
}

func (lru *lru) Dump() []Pair {
	pairs := make([]Pair, 0, len(lru.cache))
	for _, v := range lru.cache {
		pairs = append(pairs, *v.Value.(*Pair))
	}
	return pairs
}

func (lru *lru) remove(item *list.Element) {
	delete(lru.cache, item.Value.(*Pair).Key)
	lru.list.Remove(item)
}
