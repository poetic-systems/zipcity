package matchcache_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/matchcache"
)

func TestGetMiss(t *testing.T) {
	c := matchcache.New[string, int](2)
	if _, ok := c.Get("a"); ok {
		t.Fatal("Get on empty cache reported a hit")
	}
}

func TestAddThenGet(t *testing.T) {
	c := matchcache.New[string, int](2)
	c.Add("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(%q) = %v, %v, want 1, true", "a", v, ok)
	}
}

func TestEvictsLeastRecentlyUsed(t *testing.T) {
	c := matchcache.New[string, int](2)
	c.Add("a", 1)
	c.Add("b", 2)
	c.Get("a") // touch a, making b the least recently used
	c.Add("c", 3)

	if _, ok := c.Get("b"); ok {
		t.Fatal("b should have been evicted")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("a should still be cached, got %v, %v", v, ok)
	}
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Fatalf("c should be cached, got %v, %v", v, ok)
	}
}

func TestNegativeSizeDisablesCache(t *testing.T) {
	c := matchcache.New[string, int](-1)
	c.Add("a", 1)
	if _, ok := c.Get("a"); ok {
		t.Fatal("a cache disabled with a negative size should never report a hit")
	}
}

func TestZeroSizeDisablesCache(t *testing.T) {
	c := matchcache.New[string, int](0)
	c.Add("a", 1)
	if _, ok := c.Get("a"); ok {
		t.Fatal("a cache constructed with size 0 should never report a hit")
	}
}
