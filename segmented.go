package cache

type segmented struct {
	caches    []Cache
	evictions []chan Pair
}

// NewSegmented constructs a new segmented cache. The eviction channel of the
// input caches will be replaced. While this segmented cache is unclosed, using
// the input caches is undefined behavior. This function panics if no caches are
// specified.
//
// Arguments specified first are designated "lower" caches and hold
// lower-priority items. Items are added to the lowest internal cache. When
// items are accessed, they are moved to the next higher cache. When an internal
// cache becomes full, evicted values move to the next lower cache, until being
// completely evicted and passed through the registered eviction channel of the
// segmented cache.
//
// Internal caches are checked in reverse order to give higher caches the fast
// path.
func NewSegmented(caches ...Cache) Closer {
	if len(caches) == 0 {
		panic("segmented: no caches specified")
	}

	s := &segmented{
		caches:    make([]Cache, len(caches)),
		evictions: make([]chan Pair, len(caches)-1),
	}

	copy(s.caches, caches)

	for i := 1; i < len(s.caches); i++ {
		s.evictions[i-1] = make(chan Pair, 1)
		s.caches[i].Eviction(s.evictions[i-1], true)
	}

	return s
}

func (s *segmented) Get(key interface{}) (value interface{}, hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if value, hit = s.caches[i].Get(key); hit {
			if i != len(s.caches)-1 {
				_ = s.caches[i].Delete(key)
				_ = s.caches[i+1].Add(key, value)
				s.trickle(i + 1)
			}
			break
		}
	}
	return
}

func (s *segmented) Add(key, value interface{}) (hit bool) {
	return s.caches[0].Add(key, value)
}

func (s *segmented) Set(key, value interface{}) (hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if hit = s.caches[i].Set(key, value); hit {
			break
		}
	}
	return
}

func (s *segmented) Delete(key interface{}) (hit bool) {
	for i := len(s.caches) - 1; i >= 0; i-- {
		if hit = s.caches[i].Delete(key); hit {
			break
		}
	}
	return
}

func (s *segmented) Clear() {
	for _, c := range s.caches {
		c.Clear()
	}
}

func (s *segmented) Len() int {
	var sum int
	for _, c := range s.caches {
		sum += c.Len()
	}
	return sum
}

func (s *segmented) Eviction(e chan<- Pair, block bool) {
	s.caches[0].Eviction(e, block)
}

func (s *segmented) Dump() []Pair {
	pairs := make([]Pair, 0, s.Len())
	for _, c := range s.caches {
		pairs = append(pairs, c.Dump()...)
	}
	return pairs
}

func (s *segmented) Close() error {
	s.caches[0].Eviction(nil, true)
	for i := 1; i < len(s.caches); i++ {
		s.caches[i].Eviction(nil, true)
		close(s.evictions[i-1])
	}
	s.caches = nil
	s.evictions = nil
	return nil
}

// Catch eviction values trickling down the internal caches.
func (s *segmented) trickle(i int) {
	for i > 0 {
		select {
		case p := <-s.evictions[i-1]:
			_ = s.caches[i-1].Add(p.Key, p.Value)
			i--
		default:
			return
		}
	}
}
