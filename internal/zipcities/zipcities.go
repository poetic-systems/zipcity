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
// in it. The list is neither complete nor preferred-first, and a name's
// absence from it is not evidence against that name. See
// poetic-systems/zipcity#17.
package zipcities

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// Table maps a ZIP Code to the city names seen for it, uppercase and without
// repeats.
type Table map[string][]string

// Add records a name for a ZIP Code. Both are uppercased, so the keys are the
// keys the filter is built from and a caller cannot get a different answer by
// asking in a different case.
func (t Table) Add(zip, city string) {
	zip, city = strings.ToUpper(zip), strings.ToUpper(city)
	if zip == "" || city == "" {
		return
	}
	if !slices.Contains(t[zip], city) {
		t[zip] = append(t[zip], city)
	}
}

// Encode writes the table as one line per ZIP Code: the code, then its names,
// tab separated.
//
// A line per ZIP Code rather than a row per pair because the code is the
// repeated part, and text rather than gob because the artifact is committed
// and a reviewer should be able to read a diff of it. The order is fixed —
// codes ascending, names ascending — so two generations over the same data
// write the same bytes.
func Encode(w io.Writer, t Table) error {
	out := bufio.NewWriter(w)
	for _, zip := range slices.Sorted(maps.Keys(t)) {
		names := slices.Clone(t[zip])
		slices.Sort(names)
		if _, err := fmt.Fprintf(out, "%s\t%s\n", zip, strings.Join(names, "\t")); err != nil {
			return err
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
		if len(fields) < 2 || fields[0] == "" {
			return nil, fmt.Errorf("line %d: %q is not a ZIP Code and at least one city name", line, scanner.Text())
		}
		t[fields[0]] = fields[1:]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}
