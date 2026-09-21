package zipcity

import (
	"fmt"
	"iter"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/addresstables/directionals"
	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
)

var zip5pattern = regexp.MustCompile(`^\d{5}$`)

// FalsePositiveRate is the false positive rate every compiled filter was
// built with. Each key a filter is asked is independently wrong at this
// rate; Asked is how many keys a Match asked, for weighing the chance that
// any of them was.
const FalsePositiveRate = compiled_filter.FalsePositiveRate

// Match is how a street was found: the street itself is in the filter, or
// it is not but directional variants of it are (W FOX PARK DR for FOX PARK
// DR). Every key a filter says yes to is wrong independently at
// FalsePositiveRate, no matter how many keys were asked to get there — but
// Asked, how many were asked (1 for an exact hit; 1 plus the number of
// variants tried otherwise), bounds the chance that *any* reported variant
// is spurious: 1-(1-FalsePositiveRate)^Asked. A caller who does not know
// whether the input's directional was wrong or missing needs that bound,
// not just the variants themselves.
type Match struct {
	Exact    bool
	Variants []string
	Asked    int
}

// Found is true for an exact match or any variant.
func (m Match) Found() bool {
	return m.Exact || len(m.Variants) > 0
}

// The keys hold the Pub 28 abbreviations, so those are the variants tried.
// Puerto Rico's Spanish set is not: TIGER records almost no directionals
// there (see featnames).
var pub28Directionals = slices.Collect(func(yield func(string) bool) {
	for d := range directionals.English() {
		if !yield(d.Short) {
			return
		}
	}
})

// directionalVariants adds each directional the street does not already
// carry, in front and behind: 9200 S tries W 9200 S but not 9200 S N.
func directionalVariants(street string) []string {
	parts := strings.Fields(strings.ToUpper(street))
	if len(parts) == 0 {
		return nil
	}
	variants := []string{}
	if !slices.Contains(pub28Directionals, parts[0]) {
		for _, d := range pub28Directionals {
			variants = append(variants, d+" "+strings.Join(parts, " "))
		}
	}
	if !slices.Contains(pub28Directionals, parts[len(parts)-1]) {
		for _, d := range pub28Directionals {
			variants = append(variants, strings.Join(parts, " ")+" "+d)
		}
	}
	return variants
}

// match tests the street's own key, and the keys of its directional
// variants only when that misses.
func match(f *bloom.BloomFilter, key func(street string) string, street string) Match {
	if f.TestString(key(street)) {
		return Match{Exact: true, Asked: 1}
	}
	variants := directionalVariants(street)
	m := Match{Asked: 1 + len(variants)}
	for _, v := range variants {
		if f.TestString(key(v)) {
			m.Variants = append(m.Variants, v)
		}
	}
	return m
}

func zipStreetFilter(zip, street string) (*bloom.BloomFilter, error) {
	if !zip5pattern.MatchString(zip) {
		return nil, fmt.Errorf("5-digit zip code required")
	}

	if len(street) < 1 {
		return nil, fmt.Errorf("street required")
	}

	filterId, err := compiled_filter.ZipStreetFilterForZip(zip)
	if err != nil {
		return nil, fmt.Errorf("Unable to identify bloom filter for zip: %w", err)
	}

	f, err := compiled_filter.LoadFilter(filterId)
	if err != nil {
		return nil, fmt.Errorf("Unable to load bloom filter: %w", err)
	}
	return f, nil
}

func cityStreetFilter(city, state, street string) (*bloom.BloomFilter, error) {
	if len(city) < 1 {
		return nil, fmt.Errorf("city required")
	}

	if len(state) != 2 {
		return nil, fmt.Errorf("2-letter state abbreviation required")
	}

	if len(street) < 1 {
		return nil, fmt.Errorf("street required")
	}

	filterId, err := compiled_filter.CityStreetFilterForState(state)
	if err != nil {
		return nil, fmt.Errorf("Unable to identify bloom filter for state: %w", err)
	}

	f, err := compiled_filter.LoadFilter(filterId)
	if err != nil {
		return nil, fmt.Errorf("Unable to load bloom filter: %w", err)
	}
	return f, nil
}

func CheckZipAndCity(zip, city string) (bool, error) {
	if !zip5pattern.MatchString(zip) {
		return false, fmt.Errorf("5-digit zip code required")
	}

	if len(city) < 1 {
		return false, fmt.Errorf("city required")
	}

	f, err := compiled_filter.LoadFilter(compiled_filter.ZipCity)
	if err != nil {
		return false, fmt.Errorf("Unable to load bloom filter: %w", err)
	}

	return f.TestString(bloomkeys.KeyZipCity(zip, city)), nil
}

func CheckZipAndStreet(zip, street string) (bool, error) {
	f, err := zipStreetFilter(zip, street)
	if err != nil {
		return false, err
	}

	return f.TestString(bloomkeys.KeyZipStreet(zip, street)), nil
}

// MatchZipAndStreet is CheckZipAndStreet that also looks for the street's
// directional variants when the street itself is not found.
func MatchZipAndStreet(zip, street string) (Match, error) {
	f, err := zipStreetFilter(zip, street)
	if err != nil {
		return Match{}, err
	}

	return match(f, func(s string) string { return bloomkeys.KeyZipStreet(zip, s) }, street), nil
}

func CheckCityStateAndStreet(city, state, street string) (bool, error) {
	f, err := cityStreetFilter(city, state, street)
	if err != nil {
		return false, err
	}

	return f.TestString(bloomkeys.KeyCityStateStreet(city, state, street)), nil
}

// MatchCityStateAndStreet is CheckCityStateAndStreet that also looks for the
// street's directional variants when the street itself is not found.
func MatchCityStateAndStreet(city, state, street string) (Match, error) {
	f, err := cityStreetFilter(city, state, street)
	if err != nil {
		return Match{}, err
	}

	return match(f, func(s string) string { return bloomkeys.KeyCityStateStreet(city, state, s) }, street), nil
}

// CitiesKnownFor yields the state and city names seen for a ZIP Code —
// GeoNames postal cities and the TIGER place names beside its streets — each
// under the state its source placed it in, states ascending and names
// ascending within a state. It is the relation CheckZipAndCity is built
// from, read out for a caller holding a ZIP Code and no city to ask about.
//
// It is a starting point, not an answer: the list is neither complete nor
// preferred-first, and a name's absence from it is not evidence against that
// name. A code nothing was seen for yields nothing. See
// poetic-systems/zipcity#17.
func CitiesKnownFor(zip string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		states := compiled_filter.ZipCityNames()[bloomkeys.Normalize(zip)]
		for _, state := range slices.Sorted(maps.Keys(states)) {
			for _, city := range states[state] {
				if !yield(state, city) {
					return
				}
			}
		}
	}
}
