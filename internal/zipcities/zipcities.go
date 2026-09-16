// Package zipcities holds the ZIP Code to city name relation as a table, in
// the form it is written to and read back from a generated file.
//
// The bloom filters can only refuse. ZipCityExists("20170", "HERNDON") answers
// a question the caller already had a candidate for; a caller holding a ZIP
// Code and no city has nothing to ask. This is the same relation the zip-city
// filter is built from, kept in a form that can be read out.
//
// What it holds is names we have seen for a ZIP Code — GeoNames postal cities
// and the TIGER place names taken off the face beside a side — not the cities
// in it, each under the state the source placed it in. The list is neither
// complete nor preferred-first, and a name's absence from it is not evidence
// against that name. See poetic-systems/zipcity#17.
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

// Table maps a ZIP Code to the states it reaches, and each state to the city
// names seen in it, uppercase and without repeats.
//
// A ZIP Code is nearly always in one state, but not always: a dozen or so
// straddle a line, and the military codes GeoNames files under no state at
// all. So the state hangs off the names rather than the code, and a caller
// holding a ZIP Code reads out cities and the state each is in together.
type Table map[string]map[string][]string

// Add records a name for a ZIP Code in a state, spelled the way the zip-city
// filter keys it (bloomkeys.City): uppercase, diacritics folded per Project
// US@ Appendix A, punctuation dropped and ST/MT/FT spelled out per
// Publication 28. The table is normalized rather than kept as the sources
// spell it so that a name read out of it is the name a Project US@ address
// carries, and so the same city is not listed twice because two sources
// wrote it differently. CAÑO MARTIN PEÑA is here as CANO MARTIN PENA and
// ST. ALBANS as SAINT ALBANS, the renderings the filter already answers to.
// See poetic-systems/zipcity#24.
func (t Table) Add(zip, state, city string) {
	zip, state, city = bloomkeys.Normalize(zip), bloomkeys.Normalize(state), bloomkeys.City(city)
	if zip == "" || city == "" {
		return
	}
	if t[zip] == nil {
		t[zip] = map[string][]string{}
	}
	if !slices.Contains(t[zip][state], city) {
		t[zip][state] = append(t[zip][state], city)
	}
}

// Encode writes the table as one line per ZIP Code and state: the code, the
// state, then the names seen in it, tab separated.
//
// A line per code and state rather than a row per name because those are the
// repeated part, and text rather than gob because the artifact is committed
// and a reviewer should be able to read a diff of it. The order is fixed —
// codes ascending, states ascending, names ascending — so two generations
// over the same data write the same bytes.
func Encode(w io.Writer, t Table) error {
	out := bufio.NewWriter(w)
	for _, zip := range slices.Sorted(maps.Keys(t)) {
		for _, state := range slices.Sorted(maps.Keys(t[zip])) {
			names := slices.Sorted(slices.Values(t[zip][state]))
			if _, err := fmt.Fprintf(out, "%s\t%s\t%s\n", zip, state, strings.Join(names, "\t")); err != nil {
				return err
			}
		}
	}

	return out.Flush()
}

// Decode reads back what Encode wrote.
func Decode(r io.Reader) (Table, error) {
	t := make(Table)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 3 || fields[0] == "" {
			return nil, fmt.Errorf("line %d: %q is not a ZIP Code, a state and at least one city name", line, scanner.Text())
		}
		if t[fields[0]] == nil {
			t[fields[0]] = map[string][]string{}
		}
		t[fields[0]][fields[1]] = append(t[fields[0]][fields[1]], fields[2:]...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}
