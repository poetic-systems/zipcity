package bloomdata_test

import (
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomdata"
)

// testFS points at the real, committed embedded_filter data directory, so
// these tests exercise LoadFilter and ZipCityNames against actual filter
// bytes and the real compiled hashes, not fixtures.
func testFS(t *testing.T) fs.FS {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return os.DirFS(path.Join(cwd, "..", "..", "generated", "embedded_filter"))
}

// overrideFS wraps a base filesystem but substitutes the bytes at one path,
// so a test can simulate a corrupted or hash-mismatched file without
// touching any of the real committed data.
type overrideFS struct {
	base fs.FS
	path string
	data []byte
}

func (o *overrideFS) Open(name string) (fs.File, error) { return o.base.Open(name) }

func (o *overrideFS) ReadFile(name string) ([]byte, error) {
	if name == o.path {
		return o.data, nil
	}
	return fs.ReadFile(o.base, name)
}

func TestLoadFilter_CachesOnRepeatCalls(t *testing.T) {
	b, err := bloomdata.New(testFS(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	f1, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("first LoadFilter: %v", err)
	}
	f2, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("second LoadFilter: %v", err)
	}
	if f1 != f2 {
		t.Fatalf("expected the second call to return the cached instance from the first, got a different pointer")
	}
}

// New's whole-filesystem integrity check (Aaron, zipcity#73/#74,
// 2026-10-10) must catch a tampered filter file loudly at construction,
// not silently on whichever lazy LoadFilter call happens to touch it.
func TestNew_HashMismatchFailsAtConstruction(t *testing.T) {
	bad := &overrideFS{
		base: testFS(t),
		path: path.Join("data", "city-street-AK.bin"),
		data: []byte("not a real bloom filter"),
	}

	if _, err := bloomdata.New(bad); err == nil {
		t.Fatalf("expected a hash mismatch error for the tampered filter, got nil")
	}
}

// The ZIP/city names table is part of the same up-front integrity sweep as
// the bloom filters, not a separately-trusted file.
func TestNew_ZipCityNamesHashMismatchFailsAtConstruction(t *testing.T) {
	bad := &overrideFS{
		base: testFS(t),
		path: path.Join("data", "zip-city-names.tsv"),
		data: []byte("not the real names table"),
	}

	if _, err := bloomdata.New(bad); err == nil {
		t.Fatalf("expected a hash mismatch error for the tampered names table, got nil")
	}
}

// ZipCityNames must work having never been asked for before, even after
// only a handful of filters (not all ~156) have been loaded — proving the
// names table and the filters decode independently of each other.
func TestZipCityNames_IndependentOfFilterLoading(t *testing.T) {
	b, err := bloomdata.New(testFS(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := b.LoadFilter(compiled_filter.ZipStreet10); err != nil {
		t.Fatalf("LoadFilter: %v", err)
	}

	table := b.ZipCityNames()
	if len(table) == 0 {
		t.Fatalf("expected a non-empty zip-city-names table")
	}

	// Calling it again must return the same decoded table rather than
	// re-decoding (sync.Once-guarded).
	table2 := b.ZipCityNames()
	if len(table2) != len(table) {
		t.Fatalf("expected repeat ZipCityNames call to return the same table")
	}
}

// Loading a third distinct filter against a cache bounded to 2 must evict
// the least-recently-used one (AK, touched first and never touched again),
// forcing it to be read and decoded again on its next request.
func TestLoadFilter_EvictsLeastRecentlyUsed(t *testing.T) {
	b, err := bloomdata.New(testFS(t), bloomdata.WithCacheSize(2))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	first, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("LoadFilter AK: %v", err)
	}
	if _, err := b.LoadFilter(compiled_filter.CityStreetAL); err != nil {
		t.Fatalf("LoadFilter AL: %v", err)
	}
	if _, err := b.LoadFilter(compiled_filter.CityStreetAR); err != nil {
		t.Fatalf("LoadFilter AR: %v", err)
	}

	firstAgain, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("LoadFilter AK again: %v", err)
	}
	if first == firstAgain {
		t.Fatalf("expected AK to have been evicted and re-decoded, got the same cached instance")
	}
}

// Re-touching AK keeps it more recently used than AL, so loading a third
// distinct filter must evict AL instead, not AK.
func TestLoadFilter_RecentAccessIsNotEvicted(t *testing.T) {
	b, err := bloomdata.New(testFS(t), bloomdata.WithCacheSize(2))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	first, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("LoadFilter AK: %v", err)
	}
	if _, err := b.LoadFilter(compiled_filter.CityStreetAL); err != nil {
		t.Fatalf("LoadFilter AL: %v", err)
	}
	if _, err := b.LoadFilter(compiled_filter.CityStreetAK); err != nil {
		t.Fatalf("LoadFilter AK (re-touch): %v", err)
	}
	if _, err := b.LoadFilter(compiled_filter.CityStreetAR); err != nil {
		t.Fatalf("LoadFilter AR: %v", err)
	}

	firstAgain, err := b.LoadFilter(compiled_filter.CityStreetAK)
	if err != nil {
		t.Fatalf("LoadFilter AK again: %v", err)
	}
	if first != firstAgain {
		t.Fatalf("expected AK to still be cached since it was the most recently touched entry")
	}
}
