package bloomdata

import (
	"bytes"
	"container/list"
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

// DefaultCacheSize is how many decoded bloom filters BloomData keeps in
// memory at once when New is not given WithCacheSize. Aaron (zipcity#73/#74,
// 2026-10-10): a bad filter file must fail loudly at startup rather than
// silently in production, but decoding all ~156 filters into memory
// unconditionally is the cost this cache exists to avoid. 16 is an assumed
// starting point, not a measured one — most callers look up a small, local
// cluster of ZIP/state filters per request.
const DefaultCacheSize = 16

// BloomData is backed by a filesystem whose integrity is checked once, in
// full, at construction: every compiled filter and the ZIP/city names table
// are hashed against compiled_filter.DataFileSHA256 before New returns, so a
// corrupt or truncated data file fails the build/startup step instead of
// surfacing as a silent wrong answer on whichever lazy first use happens to
// hit it. Decoding into a *bloom.BloomFilter is still deferred to first use
// (that's the expensive, discardable step), and the decoded result is held
// in an LRU cache bounded to cacheSize entries, so a caller touching many
// distinct filters over a long-lived process does not retain all of them.
type BloomData struct {
	FS fs.FS

	mu        sync.Mutex
	cacheSize int
	cache     *list.List // front = most recently used
	index     map[compiled_filter.CompiledFilter]*list.Element

	namesOnce  sync.Once
	namesTable zipcities.Table
}

type cacheEntry struct {
	name   compiled_filter.CompiledFilter
	filter *bloom.BloomFilter
}

// Option configures New. WithCacheSize is currently the only one.
type Option func(*BloomData)

// WithCacheSize bounds how many decoded bloom filters New's BloomData keeps
// before evicting the least-recently-used one. n must be at least 1.
func WithCacheSize(n int) Option {
	return func(b *BloomData) {
		b.cacheSize = n
	}
}

func New(bloomDir fs.FS, opts ...Option) (*BloomData, error) {
	readFS, ok := bloomDir.(fs.ReadFileFS)
	if !ok {
		return nil, fmt.Errorf("not a readable filesystem")
	}

	for name, expectedHash := range compiled_filter.DataFileSHA256 {
		keypath := path.Join("data", bloomfilename.Filename(string(name)))
		if err := verifyFileHash(readFS, keypath, expectedHash); err != nil {
			return nil, fmt.Errorf("%s bloom filter failed integrity check: %w", name, err)
		}
	}
	if err := verifyFileHash(readFS, path.Join("data", "zip-city-names.tsv"), compiled_filter.ZipCityNamesSHA256); err != nil {
		return nil, fmt.Errorf("zip-city names failed integrity check: %w", err)
	}

	b := &BloomData{
		FS:        readFS,
		cacheSize: DefaultCacheSize,
		cache:     list.New(),
		index:     map[compiled_filter.CompiledFilter]*list.Element{},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b, nil
}

func verifyFileHash(readFS fs.ReadFileFS, keypath, expectedHash string) error {
	data, err := readFS.ReadFile(keypath)
	if err != nil {
		return fmt.Errorf("Failed to read %s: %w", keypath, err)
	}

	computed := sha256.Sum256(data)
	computedHex := hex.EncodeToString(computed[:])
	if computedHex != expectedHash {
		return fmt.Errorf("hash mismatch: compiled %q, computed %q", expectedHash, computedHex)
	}
	return nil
}

// ZipCityNames is the ZIP Code to city name table CheckZipAndCity and
// CitiesKnownFor both read exactly, decoded on first use so a caller who
// only asks the street filters pays for the bytes and nothing more. Its
// hash was already checked in New, so a read or decode failure here is a
// build defect rather than a condition a caller can recover from — same as
// the old eager init() — so unlike LoadFilter this panics rather than
// returning an error. This preserves ZipCityNames' existing no-error
// signature (CheckZipAndCity, CitiesKnownFor, CitiesRecommendedFor and
// ZipsKnownFor's zipsByStateCity have no error channel today).
func (b *BloomData) ZipCityNames() zipcities.Table {
	b.namesOnce.Do(func() {
		readFS := b.FS.(fs.ReadFileFS)

		data, err := readFS.ReadFile(path.Join("data", "zip-city-names.tsv"))
		if err != nil {
			panic(fmt.Errorf("Failed to read zip-city names: %w", err))
		}

		t, err := zipcities.Decode(bytes.NewReader(data))
		if err != nil {
			panic(fmt.Errorf("Failed to decode zip-city names: %w", err))
		}

		b.namesTable = t
	})
	return b.namesTable
}

// LoadFilter decodes one compiled filter on first request and keeps it in
// an LRU cache bounded to cacheSize entries, re-decoding it if it was since
// evicted. Its hash was already checked in New, so unlike that eager pass a
// read or decode failure here only errors out the one filter requested — a
// caller that only needs a subset of filters should not have one bad
// filter take down every other lookup.
func (b *BloomData) LoadFilter(name compiled_filter.CompiledFilter) (*bloom.BloomFilter, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if elem, ok := b.index[name]; ok {
		b.cache.MoveToFront(elem)
		return elem.Value.(*cacheEntry).filter, nil
	}

	if _, ok := compiled_filter.DataFileSHA256[name]; !ok {
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}

	readFS := b.FS.(fs.ReadFileFS)
	keypath := path.Join("data", bloomfilename.Filename(string(name)))

	data, err := readFS.ReadFile(keypath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s bloom filter: %w", name, err)
	}

	filter := &bloom.BloomFilter{}
	if _, err := filter.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("Failed to decode %s bloom filter: %w", name, err)
	}

	elem := b.cache.PushFront(&cacheEntry{name: name, filter: filter})
	b.index[name] = elem
	for b.cache.Len() > b.cacheSize {
		oldest := b.cache.Back()
		b.cache.Remove(oldest)
		delete(b.index, oldest.Value.(*cacheEntry).name)
	}

	return filter, nil
}
