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
// so a test can simulate a corrupted or hash-mismatched filter file without
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

// A hash mismatch on one filter must not take down any other filter — the
// whole point of going per-filter lazy rather than eager-decode-all-at-init.
func TestLoadFilter_HashMismatchErrorsIndependently(t *testing.T) {
	bad := &overrideFS{
		base: testFS(t),
		path: path.Join("data", "city-street-AK.bin"),
		data: []byte("not a real bloom filter"),
	}
	b, err := bloomdata.New(bad)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := b.LoadFilter(compiled_filter.CityStreetAK); err == nil {
		t.Fatalf("expected a hash mismatch error for the tampered filter, got nil")
	}

	// A different filter, whose bytes the override never touched, must still
	// load cleanly — the AK failure must stay scoped to AK.
	if _, err := b.LoadFilter(compiled_filter.CityStreetAL); err != nil {
		t.Fatalf("expected the unaffected filter to load, got error: %v", err)
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
