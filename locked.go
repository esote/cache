package cache

import "sync"

type locked struct {
	cache Cache
	mu    sync.Mutex
}

// NewLocked wraps a cache in mutex locks.
func NewLocked(cache Cache) Cache {
	return &locked{
		cache: cache,
	}
}

func (l *locked) Get(key interface{}) (value interface{}, hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Get(key)
}

func (l *locked) Add(key, value interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Add(key, value)
}

func (l *locked) Set(key, value interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Set(key, value)
}

func (l *locked) Delete(key interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Delete(key)
}

func (l *locked) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Clear()
}

func (l *locked) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Len()
}

func (l *locked) Eviction(e chan<- Pair, block bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Eviction(e, block)
}

func (l *locked) Dump() []Pair {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Dump()
}
