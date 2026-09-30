// Package zipcities holds the ZIP Code to city name relation as a table, in
// the form it is written to and read back from a generated file.
//
// The other bloom filters can only refuse. CheckZipAndCity("20170", "HERNDON")
// answers a question the caller already had a candidate for; a caller
// holding a ZIP Code and no city has nothing to ask. This table is what
// CheckZipAndCity itself reads for an exact answer, kept in a form that can
// also be read out directly.
//
// What it holds is names we have seen for a ZIP Code — GeoNames postal cities,
// the TIGER place names taken off the face beside a side, and the Postal
// Service's own delivery unit for the code — not the cities in it, each under
// the state the source placed it in. So the question the table answers is
// "has a source ever written this name on this ZIP Code", not "is this a
// valid last line": the list is neither complete nor preferred-first, and a
// name's absence from it is not evidence against that name. What it can say
// is which source offered a name, which is what Source is for. See
// poetic-systems/zipcity#17 and poetic-systems/zipcity#48.
package zipcities

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/poetic-systems/zipcity/internal/bloomkeys"
)

// Source is where a name came from, recorded per name because the sources are
// not equally good and a caller has to be able to tell them apart.
//
// Seen is a name some source associated with the ZIP Code — a GeoNames postal
// city or a TIGER place beside one of its streets. USPS is the delivery unit
// the Postal Service's ZIP Locale Detail table names for the code, which is
// the name most likely to be on an envelope and still wrong about one time in
// seven, because the unit that delivers a code is often in the next town over.
type Source string

const (
	Seen Source = "seen"
	USPS Source = "usps"
)

// sources are the sources a line may name, in the order Encode writes them.
var sources = []Source{Seen, USPS}

// Name is one city name and the source that offered it. A name both sources
// offer is recorded once per source, because "GeoNames and the Postal Service
// agree" is a different fact from either one alone.
type Name struct {
	City   string
	Source Source
}

// Table maps a ZIP Code to the states it reaches, and each state to the names
// recorded in it, uppercase and without repeats within a source.
//
// A ZIP Code is nearly always in one state, but not always: a dozen or so
// straddle a line, and the military codes GeoNames files under no state at
// all. So the state hangs off the names rather than the code, and a caller
// holding a ZIP Code reads out cities and the state each is in together.
type Table map[string]map[string][]Name

// Add records a name for a ZIP Code in a state, from a source, spelled the way
// CheckZipAndCity keys it (bloomkeys.City): uppercase, diacritics folded per
// Project US@ Appendix A, punctuation dropped and ST/MT/FT spelled out per
// Publication 28. The table is normalized rather than kept as the sources
// spell it so that a name read out of it is the name a Project US@ address
// carries, and so the same city is not listed twice because two sources
// wrote it differently. CAÑO MARTIN PEÑA is here as CANO MARTIN PENA and
// ST. ALBANS as SAINT ALBANS, the renderings CheckZipAndCity already answers
// to. See poetic-systems/zipcity#24.
func (t Table) Add(zip, state, city string, source Source) {
	zip, state, city = bloomkeys.Normalize(zip), bloomkeys.Normalize(state), bloomkeys.City(city)
	if zip == "" || city == "" {
		return
	}
	if t[zip] == nil {
		t[zip] = map[string][]Name{}
	}
	name := Name{City: city, Source: source}
	if !slices.Contains(t[zip][state], name) {
		t[zip][state] = append(t[zip][state], name)
	}
}

// Names are the city names recorded for a ZIP Code in a state, whatever their
// source, ascending and without repeats. A name two sources offered is one
// name to a caller who only asked what the names are.
func (t Table) Names(zip, state string) []string {
	names := []string{}
	for _, name := range t[zip][state] {
		names = append(names, name.City)
	}
	slices.Sort(names)

	return slices.Compact(names)
}

// NamesFrom are the city names one source recorded for a ZIP Code in a state,
// ascending. NamesFrom(zip, state, USPS) is what the Postal Service calls the
// code, which a caller weighing several names wants to know.
func (t Table) NamesFrom(zip, state string, source Source) []string {
	names := []string{}
	for _, name := range t[zip][state] {
		if name.Source == source {
			names = append(names, name.City)
		}
	}
	slices.Sort(names)

	return names
}

// Encode writes the table as one line per ZIP Code, state and source: the
// code, the state, the source, then the names that source recorded, tab
// separated.
//
// A line per code, state and source rather than a row per name because those
// are the repeated part, and text rather than gob because the artifact is
// committed and a reviewer should be able to read a diff of it. The source is
// a column rather than a marker inside a name so the names stay the tail of
// the line and adding one moves one field. The order is fixed — codes
// ascending, states ascending, sources in the order they are declared, names
// ascending — so two generations over the same data write the same bytes.
func Encode(w io.Writer, t Table) error {
	out := bufio.NewWriter(w)
	for _, zip := range slices.Sorted(maps.Keys(t)) {
		for _, state := range slices.Sorted(maps.Keys(t[zip])) {
			for _, source := range sources {
				names := t.NamesFrom(zip, state, source)
				if len(names) == 0 {
					continue
				}
				if _, err := fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", zip, state, source, strings.Join(names, "\t")); err != nil {
					return err
				}
			}
		}
	}

	return out.Flush()
}

// Decode reads back what Encode wrote. A line naming a source this package
// does not know is an error rather than a name filed under nothing.
func Decode(r io.Reader) (Table, error) {
	t := make(Table)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 4 || fields[0] == "" {
			return nil, fmt.Errorf("line %d: %q is not a ZIP Code, a state, a source and at least one city name", line, scanner.Text())
		}
		source := Source(fields[2])
		if !slices.Contains(sources, source) {
			return nil, fmt.Errorf("line %d: %q is not a source this package knows", line, fields[2])
		}
		if t[fields[0]] == nil {
			t[fields[0]] = map[string][]Name{}
		}
		for _, city := range fields[3:] {
			t[fields[0]][fields[1]] = append(t[fields[0]][fields[1]], Name{City: city, Source: source})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}
