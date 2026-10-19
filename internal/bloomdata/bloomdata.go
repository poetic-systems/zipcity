package bloomdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"path"
	"slices"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
	"github.com/poetic-systems/zipcity/internal/zipcities"
)

type BloomData struct {
	FS             fs.FS
	Hash           []byte
	bloom_filters  map[compiled_filter.CompiledFilter]*bloom.BloomFilter
	zip_city_names zipcities.Table
	conflicted     *bloom.BloomFilter
}

// ConflictedAbbreviationFilter names the filter of word pairs around
// abbreviations that read more than one way (CT, MT, NE, ST, DR). It is
// not a compiled_filter: it is built from street and city names together,
// after the zip-city names are written, and is hashed after them.
const ConflictedAbbreviationFilter = "conflicted-abbreviation"

func New(bloomDir fs.FS) (*BloomData, error) {
	b := &BloomData{
		FS: bloomDir,
	}
	err := b.init()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *BloomData) init() error {
	readFS, ok := b.FS.(fs.ReadFileFS)
	if !ok {
		return fmt.Errorf("not a readable filesystem")
	}

	h := sha256.New()
	h.Reset()

	var initerr error

	// Keep a map of the raw, zero-allocation byte arrays
	filternames := slices.Collect(compiled_filter.All())
	slices.Sort(filternames)
	b.bloom_filters = maps.Collect(func(yield func(compiled_filter.CompiledFilter, *bloom.BloomFilter) bool) {
		for _, name := range filternames {

			keypath := path.Join("data", bloomfilename.Filename(string(name)))

			// ReadFile allocates each byte slice once during boot
			data, err := readFS.ReadFile(keypath)
			if err != nil {
				initerr = err
				return
			}

			filter := &bloom.BloomFilter{}
			reader := bytes.NewReader(data)
			tee := io.TeeReader(reader, h)
			_, err = filter.ReadFrom(tee)
			if err != nil {
				initerr = fmt.Errorf("Failed to read %s bloom filter: %w", name, err)
				return
			}

			if !yield(name, filter) {
				return
			}
		}
	})
	if initerr != nil {
		return initerr
	}

	zipCityNames, err := readFS.ReadFile(path.Join("./data", "zip-city-names.tsv"))
	if err != nil {
		return err
	}
	reader := bytes.NewReader(zipCityNames)
	tee := io.TeeReader(reader, h)
	t, err := zipcities.Decode(tee)
	if err != nil {
		return fmt.Errorf("Failed to read zip-city names: %w", err)
	}

	b.zip_city_names = t

	// Data generated before the conflicted-abbreviation filter existed has no
	// such file, and its hash is of what it does have.
	conflictedData, err := readFS.ReadFile(path.Join("data", bloomfilename.Filename(ConflictedAbbreviationFilter)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return err
	default:
		filter := &bloom.BloomFilter{}
		_, err = filter.ReadFrom(io.TeeReader(bytes.NewReader(conflictedData), h))
		if err != nil {
			return fmt.Errorf("Failed to read %s bloom filter: %w", ConflictedAbbreviationFilter, err)
		}
		b.conflicted = filter
	}

	computedSha256 := hex.EncodeToString(h.Sum(nil))

	// check the hash of filter files
	if compiled_filter.DataFilesSHA256 != computedSha256 {
		return fmt.Errorf("bloom filter FS data hash mismatch")
	}

	return nil
}

// ZipCityNames is the ZIP Code to city name table CheckZipAndCity and
// CitiesKnownFor both read exactly, decoded on first use so a caller who
// only asks the street filters pays for the bytes and nothing more. The
// file is written by the same generation that reads it back, so failing to
// read it is a build defect, and fails the way an unreadable filter does.
func (b *BloomData) ZipCityNames() zipcities.Table {
	return b.zip_city_names
}

// LoadFilter restores the compiled filter in memory
func (b *BloomData) LoadFilter(name compiled_filter.CompiledFilter) (*bloom.BloomFilter, error) {
	filter, ok := b.bloom_filters[name]
	if !ok {
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	return filter, nil
}

// ConflictedAbbreviation is the conflicted-abbreviation filter, or nil when
// the data was generated without one.
func (b *BloomData) ConflictedAbbreviation() *bloom.BloomFilter {
	return b.conflicted
}
