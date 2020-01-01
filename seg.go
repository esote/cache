package cache

// Segmented combines caches into a prioritized list.
type Segmented struct {
	caches []Cache
}

// NewSegmented constructs a segmented cache. Argument order matters: caches
// specified first have lower priority than caches specified last. Cache values
// move from lower to higher priority caches as they are accessed.
//
// Panics if no caches are specified.
func NewSegmented(caches ...Cache) *Segmented {
	if len(caches) == 0 {
		panic("seg: no caches specified")
	}

	all := make([]Cache, 0, len(caches))
	for _, c := range caches {
		if s, ok := c.(*Segmented); ok {
			all = append(all, s.caches...)
		} else {
			all = append(all, c)
		}
	}
	return &Segmented{all}
}

// Get value from the segmented cache. Caches are checked in reverse order to
// give higher-priority caches the fast path. Upon a cache hit, the value is
// moved to the next higher-priority cache.
func (s *Segmented) Get(key interface{}) (value interface{}, hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if value, hit = s.caches[i].Get(key); hit {
			if i != len(s.caches)-1 {
				_ = s.caches[i].Delete(key)
				_ = s.caches[i+1].Add(key, value)
			}
			break
		}
	}
	return
}

// Add value to the segmented cache. The value is added to the lowest-priority
// cache.
func (s *Segmented) Add(key, value interface{}) (hit bool) {
	return s.caches[0].Add(key, value)
}

// Set value in the segmented cache. Caches are checked in reverse order to give
// higher-priority caches the fast path. Returns on the first Set cache hit.
func (s *Segmented) Set(key, value interface{}) (hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if hit = s.caches[i].Set(key, value); hit {
			break
		}
	}
	return
}

// Delete value from the cache. Caches are checked in reverse order to give
// higher-priority caches the fast path. Returns on the first Delete cache hit.
func (s *Segmented) Delete(key interface{}) (hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if hit = s.caches[i].Delete(key); hit {
			break
		}
	}
	return
}

// Clear all values in all the internal caches.
func (s *Segmented) Clear() {
	for _, c := range s.caches {
		c.Clear()
	}
}

// Len returns the sum of the lengths of all internal caches.
func (s *Segmented) Len() int {
	var sum int
	for _, c := range s.caches {
		sum += c.Len()
	}
	return sum
}
