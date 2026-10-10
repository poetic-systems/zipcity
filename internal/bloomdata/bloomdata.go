package bloomdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"sync"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
	"github.com/poetic-systems/zipcity/internal/zipcities"
)

// BloomData is the per-filter-lazy replacement for the old eager-decode-all
// struct: New only validates the filesystem it is handed, and every filter
// — plus the ZIP/city names table — is read from disk and decoded on its own
// first use, not at construction. A caller who only ever touches a handful
// of states or ZIP prefixes now pays for exactly those filters' bytes, not
// all ~156.
type BloomData struct {
	FS fs.FS

	mu     sync.Mutex
	loaded map[compiled_filter.CompiledFilter]*bloom.BloomFilter

	namesOnce  sync.Once
	namesTable zipcities.Table
}

func New(bloomDir fs.FS) (*BloomData, error) {
	if _, ok := bloomDir.(fs.ReadFileFS); !ok {
		return nil, fmt.Errorf("not a readable filesystem")
	}
	return &BloomData{
		FS:     bloomDir,
		loaded: map[compiled_filter.CompiledFilter]*bloom.BloomFilter{},
	}, nil
}

// ZipCityNames is the ZIP Code to city name table CheckZipAndCity and
// CitiesKnownFor both read exactly, decoded on first use so a caller who
// only asks the street filters pays for the bytes and nothing more. The
// file is written by the same generation that reads it back, so failing to
// read, hash, or decode it is a build defect — not a condition a caller can
// recover from — so unlike LoadFilter this panics rather than returning an
// error, the same way the old eager init() did for every filter. This
// preserves ZipCityNames' existing no-error signature (CheckZipAndCity,
// CitiesKnownFor, CitiesRecommendedFor and ZipsKnownFor's zipsByStateCity
// have no error channel today), deferring the cost to first use without
// touching any of those four public call sites.
func (b *BloomData) ZipCityNames() zipcities.Table {
	b.namesOnce.Do(func() {
		readFS := b.FS.(fs.ReadFileFS)

		data, err := readFS.ReadFile(path.Join("data", "zip-city-names.tsv"))
		if err != nil {
			panic(fmt.Errorf("Failed to read zip-city names: %w", err))
		}

		computed := sha256.Sum256(data)
		computedHex := hex.EncodeToString(computed[:])
		if computedHex != compiled_filter.ZipCityNamesSHA256 {
			panic(fmt.Errorf("zip-city names hash mismatch: compiled %q, computed %q", compiled_filter.ZipCityNamesSHA256, computedHex))
		}

		t, err := zipcities.Decode(bytes.NewReader(data))
		if err != nil {
			panic(fmt.Errorf("Failed to decode zip-city names: %w", err))
		}

		b.namesTable = t
	})
	return b.namesTable
}

// LoadFilter restores one compiled filter in memory, reading and decoding
// it from the filesystem on first request and caching the result for
// later callers. Unlike ZipCityNames, a bad read, hash mismatch, or decode
// failure here is returned as an error rather than panicking: a caller
// that only needs a subset of filters should not have one bad filter take
// down every other lookup the way the old eager init() did.
func (b *BloomData) LoadFilter(name compiled_filter.CompiledFilter) (*bloom.BloomFilter, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if filter, ok := b.loaded[name]; ok {
		return filter, nil
	}

	expectedHash, ok := compiled_filter.DataFileSHA256[name]
	if !ok {
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}

	readFS := b.FS.(fs.ReadFileFS)
	keypath := path.Join("data", bloomfilename.Filename(string(name)))

	data, err := readFS.ReadFile(keypath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s bloom filter: %w", name, err)
	}

	computed := sha256.Sum256(data)
	computedHex := hex.EncodeToString(computed[:])
	if computedHex != expectedHash {
		return nil, fmt.Errorf("%s bloom filter hash mismatch: compiled %q, computed %q", name, expectedHash, computedHex)
	}

	filter := &bloom.BloomFilter{}
	if _, err := filter.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("Failed to decode %s bloom filter: %w", name, err)
	}

	b.loaded[name] = filter
	return filter, nil
}
