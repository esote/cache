// Package cache implements many kinds of caching policies.
//
// Cache replacement algorithms currently implemented:
//
//	FIFO: first-in first-out
//	LFU: least-frequently used
//	LIFO: last-in first-out
//	LRU: least-recently used
//	MRU: most-recently used
//	RR: random-replacement
package cache

import "sync"

// Cache represents a cache implementation.
type Cache interface {
	// Get value in the cache.
	Get(key interface{}) (value interface{}, hit bool)

	// Add value to the cache.
	Add(key, value interface{}) (hit bool)

	// Set value in the cache.
	Set(key, value interface{}) (hit bool)

	// Delete value from the cache. Returns whether the delete hit.
	Delete(key interface{}) (hit bool)

	// Clear all values in the cache.
	Clear()

	// Len returns the number of items in the cache.
	Len() int
}

// Locked wraps a cache in mutex locks.
func Locked(cache Cache) Cache {
	return &lockedCache{
		cache: cache,
	}
}

type lockedCache struct {
	cache Cache
	mu    sync.Mutex
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
