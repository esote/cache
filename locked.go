package cache

import "sync"

type lockedCache struct {
	cache Cache
	mu    sync.Mutex
}

// NewLocked wraps a cache in mutex locks.
func NewLocked(cache Cache) Cache {
	return &lockedCache{
		cache: cache,
	}
}

func (l *lockedCache) Get(key interface{}) (value interface{}, hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Get(key)
}

func (l *lockedCache) Add(key, value interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Add(key, value)
}

func (l *lockedCache) Set(key, value interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Set(key, value)
}

func (l *lockedCache) Delete(key interface{}) (hit bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Delete(key)
}

func (l *lockedCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Clear()
}

func (l *lockedCache) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Len()
}

func (l *lockedCache) Eviction(e chan<- Pair, block bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Eviction(e, block)
}

func (l *lockedCache) Dump() []Pair {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache.Dump()
}
