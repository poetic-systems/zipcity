// Package matchcache is a small fixed-size, least-recently-used cache keyed
// by an arbitrary comparable type. zipcity uses it to memoize
// MatchZipAndStreet and MatchCityStateAndStreet results (see
// poetic-systems/zipcity#75).
package matchcache

import (
	"container/list"
	"sync"
)

// Cache is a fixed-size, least-recently-used cache. A Cache whose size is
// zero or negative is disabled: Add is a no-op and Get always misses.
type Cache[K comparable, V any] struct {
	mu    sync.Mutex
	size  int
	ll    *list.List // front = most recently used
	index map[K]*list.Element
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// New returns a Cache holding up to size entries. size <= 0 disables the
// cache.
func New[K comparable, V any](size int) *Cache[K, V] {
	return &Cache[K, V]{
		size:  size,
		ll:    list.New(),
		index: map[K]*list.Element{},
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.size <= 0 {
		var zero V
		return zero, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.index[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.ll.MoveToFront(el)
	return el.Value.(entry[K, V]).value, true
}

func (c *Cache[K, V]) Add(key K, value V) {
	if c.size <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.index[key]; ok {
		el.Value = entry[K, V]{key: key, value: value}
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(entry[K, V]{key: key, value: value})
	c.index[key] = el
	if c.ll.Len() > c.size {
		oldest := c.ll.Back()
		c.ll.Remove(oldest)
		delete(c.index, oldest.Value.(entry[K, V]).key)
	}
}
