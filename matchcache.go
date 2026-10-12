package zipcity

import (
	"container/list"
	"sync"
)

// DefaultMatchCacheSize is how many distinct Match results New's ZipCity
// keeps per cache when New is not given WithMatchCacheSize. go-projectusat
// generates several candidate readings per address and often asks the same
// (zip, street) or (city, state, street) pair more than once while doing
// so (poetic-systems/zipcity#75) — 4096 is an assumed starting point sized
// to cover a request's worth of candidates many times over, not a measured
// one.
const DefaultMatchCacheSize = 4096

type zipStreetKey struct {
	zip, street string
}

type cityStateStreetKey struct {
	city, state, street string
}

// matchCache is a fixed-size, least-recently-used cache of Match results,
// the same shape MatchZipAndStreet and MatchCityStateAndStreet each need —
// every key a caller supplies is used exactly as given, no normalization,
// since the filter data underneath a (zip, street) or (city, state, street)
// triple cannot change within a process.
type matchCache[K comparable] struct {
	mu    sync.Mutex
	size  int
	ll    *list.List // front = most recently used
	index map[K]*list.Element
}

type matchCacheEntry[K comparable] struct {
	key   K
	value Match
}

func newMatchCache[K comparable](size int) *matchCache[K] {
	return &matchCache[K]{
		size:  size,
		ll:    list.New(),
		index: map[K]*list.Element{},
	}
}

func (c *matchCache[K]) get(key K) (Match, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.index[key]
	if !ok {
		return Match{}, false
	}
	c.ll.MoveToFront(el)
	return el.Value.(matchCacheEntry[K]).value, true
}

func (c *matchCache[K]) add(key K, value Match) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.index[key]; ok {
		el.Value = matchCacheEntry[K]{key: key, value: value}
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(matchCacheEntry[K]{key: key, value: value})
	c.index[key] = el
	if c.ll.Len() > c.size {
		oldest := c.ll.Back()
		c.ll.Remove(oldest)
		delete(c.index, oldest.Value.(matchCacheEntry[K]).key)
	}
}
