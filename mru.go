package cache

import "container/list"

type mru struct {
	capacity int

	cache map[interface{}]*list.Element
	list  *list.List

	e     chan<- Pair
	block bool
}

// NewMRU constructs a new most-recently-used cache. Items which are accessed
// most-recently are evicted first. This is identical to LRU cache except
// pruning begins at the front.
func NewMRU(capacity int) Cache {
	if capacity <= 0 {
		panic("mru: capacity <= 0")
	}

	return &mru{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element, capacity),
		list:     list.New(),
	}
}

func (mru *mru) Get(key interface{}) (value interface{}, hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.list.MoveToFront(item)
		value = item.Value.(*Pair).Value
	}
	return
}

func (mru *mru) Add(key, value interface{}) (hit bool) {
	if _, hit = mru.cache[key]; hit {
		return
	}

	if len(mru.cache) >= mru.capacity {
		if item := mru.list.Front(); item != nil {
			send(mru.e, mru.block, *item.Value.(*Pair))
			mru.remove(item)
		}
	}

	mru.cache[key] = mru.list.PushFront(&Pair{key, value})
	return
}

func (mru *mru) Set(key, value interface{}) (hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.list.MoveToFront(item)
		item.Value.(*Pair).Value = value
	}
	return
}

func (mru *mru) Delete(key interface{}) (hit bool) {
	var item *list.Element
	if item, hit = mru.cache[key]; hit {
		mru.remove(item)
	}
	return
}

func (mru *mru) Clear() {
	mru.cache = make(map[interface{}]*list.Element, mru.capacity)
	mru.list = mru.list.Init()
}

func (mru *mru) Len() int {
	return len(mru.cache)
}

func (mru *mru) Eviction(e chan<- Pair, block bool) {
	mru.e, mru.block = e, block
}

func (mru *mru) Dump() []Pair {
	pairs := make([]Pair, 0, len(mru.cache))
	for _, v := range mru.cache {
		pairs = append(pairs, *v.Value.(*Pair))
	}
	return pairs
}

func (mru *mru) remove(item *list.Element) {
	delete(mru.cache, item.Value.(*Pair).Key)
	mru.list.Remove(item)
}
