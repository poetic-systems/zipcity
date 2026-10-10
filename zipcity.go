package zipcity

import (
	"fmt"
	"iter"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/addresstables/directionals"
	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomdata"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
	"github.com/poetic-systems/zipcity/internal/zipcities"
	"github.com/poetic-systems/zipcity/pkg/filterfs"
	"github.com/poetic-systems/zipcity/pkg/filterfs/embedded"
)

var zipcityFS filterfs.InitFunc

var filters = sync.OnceValue(func() *bloomdata.BloomData {
	fmt.Println("Initializing zipcity filesystem")
	if zipcityFS == nil {
		// panic(fmt.Errorf("you must register a zipcity filterfs.InitFunc"))
		RegisterFS(embedded.PrepareFS)
	}
	files, err := zipcityFS.PrepareFS()
	if err != nil {
		panic(err)
	}
	b, err := bloomdata.New(files)
	if err != nil {
		panic(err)
	}

	return b
})

func RegisterFS(ffs filterfs.InitFn) {
	zipcityFS = ffs
}

var zip5pattern = regexp.MustCompile(`^\d{5}$`)

// ZipStreetFalsePositiveRate and CityStreetFalsePositiveRate are the false
// positive rates the zip-street and city-street compiled filters were each
// built with. Each key a filter is asked is independently wrong at its own
// filter's rate; Asked is how many keys a Match asked, for weighing the
// chance that any of them was. They are separate constants, rather than one
// shared rate, so that one filter type's rate can change without changing
// the other's API.
const (
	ZipStreetFalsePositiveRate  = compiled_filter.ZipStreetFalsePositiveRate
	CityStreetFalsePositiveRate = compiled_filter.CityStreetFalsePositiveRate
)

// Match is how a street was found: the street itself is in the filter, or
// it is not but directional variants of it are (W FOX PARK DR for FOX PARK
// DR). Every key a filter says yes to is wrong independently at the rate
// its filter was built with — ZipStreetFalsePositiveRate for
// MatchZipAndStreet, CityStreetFalsePositiveRate for
// MatchCityStateAndStreet — no matter how many keys were asked to get
// there. But Asked, how many were asked (1 for an exact hit; 1 plus the
// number of variants tried otherwise), bounds the chance that *any*
// reported variant is spurious: 1-(1-rate)^Asked. A caller who does not
// know whether the input's directional was wrong or missing needs that
// bound, not just the variants themselves.
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

	f, err := filters().LoadFilter(filterId)
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

	f, err := filters().LoadFilter(filterId)
	if err != nil {
		return nil, fmt.Errorf("Unable to load bloom filter: %w", err)
	}
	return f, nil
}

// CheckZipAndCity reports whether a city name has been seen for a ZIP Code,
// state-blind, the way the zip-city filter's key used to be. It is now an
// exact answer, with no false positive rate to weigh, read straight from
// compiled_filter.ZipCityNames() — the same table CitiesKnownFor reads. See
// poetic-systems/zipcity#55.
func CheckZipAndCity(zip, city string) (bool, error) {
	if !zip5pattern.MatchString(zip) {
		return false, fmt.Errorf("5-digit zip code required")
	}

	if len(city) < 1 {
		return false, fmt.Errorf("city required")
	}

	name := bloomkeys.City(city)
	for _, names := range filters().ZipCityNames()[bloomkeys.Normalize(zip)] {
		if slices.ContainsFunc(names, func(n zipcities.Name) bool { return n.City == name }) {
			return true, nil
		}
	}
	return false, nil
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

// CitiesKnownFor yields the state and city names known for a ZIP Code —
// GeoNames postal cities, the TIGER place names beside its streets, and the
// Postal Service's own delivery unit for the code — each under the state its
// source placed it in, states ascending and names ascending within a state. A
// name more than one source offered is yielded once. It is the relation
// CheckZipAndCity is built from, read out for a caller holding a ZIP Code and
// no city to ask about.
//
// It is a starting point, not an answer: the list is neither complete nor
// preferred-first, and a name's absence from it is not evidence against that
// name. A code nothing was seen for yields nothing. A caller who needs to
// weigh the names rather than just enumerate them wants
// CitiesRecommendedFor. See poetic-systems/zipcity#17.
func CitiesKnownFor(zip string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		zip = bloomkeys.Normalize(zip)
		table := filters().ZipCityNames()
		for _, state := range slices.Sorted(maps.Keys(table[zip])) {
			for _, city := range table.Names(zip, state) {
				if !yield(state, city) {
					return
				}
			}
		}
	}
}

// CitiesRecommendedFor yields the state and city names the Postal Service's
// ZIP Locale Detail table gives for a ZIP Code: the delivery units that carry
// its mail, states ascending and names ascending within a state.
//
// These are a subset of what CitiesKnownFor yields, marked out because they
// are not the same kind of claim. A GeoNames postal city or a TIGER place is a
// name somebody associated with the code; this is the name the Postal Service
// itself puts on it, which is the one most likely to appear on an envelope.
// It is still not a guarantee — a code is often delivered from an office in
// the next town, so the recommended name is not always the name the addressee
// writes — but a caller choosing between several names should know which one
// came from here. Most codes yield one name, a few yield several where more
// than one office delivers into them, and one yields none where only stations
// and branches do. See poetic-systems/zipcity#48.
func CitiesRecommendedFor(zip string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		zip = bloomkeys.Normalize(zip)
		table := filters().ZipCityNames()
		for _, state := range slices.Sorted(maps.Keys(table[zip])) {
			for _, city := range table.NamesFrom(zip, state, zipcities.USPS) {
				if !yield(state, city) {
					return
				}
			}
		}
	}
}

// zipsByStateCity inverts compiled_filter.ZipCityNames() — ZIP Code to state
// to city names — into state and city to the ZIP Codes seen there, keyed
// the way CheckZipAndCity keys a city (bloomkeys.Normalize the state,
// bloomkeys.City the name) so a caller's spelling and case do not matter.
// Built once on first use and sorted then, rather than on every call, since
// the table it inverts is itself fixed for the life of the process.
var zipsByStateCity = sync.OnceValue(func() map[string][]string {
	inverted := map[string][]string{}
	for zip, states := range filters().ZipCityNames() {
		for state, names := range states {
			for _, name := range names {
				key := state + ":" + name.City
				inverted[key] = append(inverted[key], zip)
			}
		}
	}
	// A name both sources offered lands here once per source, so the codes are
	// compacted as well as sorted: which sources named a code is not a fact
	// about the list of codes.
	for key := range inverted {
		slices.Sort(inverted[key])
		inverted[key] = slices.Compact(inverted[key])
	}
	return inverted
})

// ZipsKnownFor yields the ZIP Codes a city name has been seen for in a
// state, ascending. It is CitiesKnownFor read the other way, for a caller
// holding a city and state and no ZIP Code to ask about.
//
// It is a starting point, not an answer: the list is neither complete nor
// preferred-first, and a code's absence from it is not evidence against
// that code. A city nothing was seen for yields nothing. See
// poetic-systems/addressparsers#17.
func ZipsKnownFor(state, city string) iter.Seq[string] {
	return func(yield func(string) bool) {
		key := bloomkeys.Normalize(state) + ":" + bloomkeys.City(city)
		for _, zip := range zipsByStateCity()[key] {
			if !yield(zip) {
				return
			}
		}
	}
}
