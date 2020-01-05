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
//
// They all operate in constant time, with the exception of the Dump function
// which has a runtime of O(n) where n is the size of the cache.
package cache

import "io"

// Cache represents a cache implementation. Unless specified otherwise, all
// caches will panic if constructed with a capacity <= 0.
type Cache interface {
	// Get value in the cache.
	Get(key interface{}) (value interface{}, hit bool)

	// Add value to the cache. If the value already exists, nothing is
	// changed.
	Add(key, value interface{}) (hit bool)

	// Set value in the cache. This serves as an optimization of Delete+Add.
	Set(key, value interface{}) (hit bool)

	// Delete value from the cache. Returns whether the delete hit.
	Delete(key interface{}) (hit bool)

	// Clear all values in the cache.
	Clear()

	// Len returns the number of items in the cache.
	Len() int

	// Eviction registers a channel through which evicted key-value pairs
	// will be sent. Only pairs automatically evicted will be sent, not
	// those manually removed with Delete.
	Eviction(e chan<- Pair, block bool)

	// Dump the contents of the cache in no particular order.
	Dump() []Pair
}

// Pair represents the key-value pair in a cache.
type Pair struct {
	Key, Value interface{}
}

// Closer represents a cache should be closed when it will no longer be
// used.
type Closer interface {
	Cache
	io.Closer
}

func send(e chan<- Pair, block bool, p Pair) {
	if e == nil {
		return
	}
	if block {
		e <- p
	} else {
		select {
		case e <- p:
		default:
		}
	}
}
