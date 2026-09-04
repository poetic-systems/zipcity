//go:build ignore
// +build ignore

// Command generate writes the TIGER/Line test fixtures from a local download
// cache.
//
// The fixtures are one connected slice of a single county rather than a
// sample of each file taken on its own: the readers exist to follow the joins
// between the files, so a fixture that breaks a join tests nothing. It is
// built outward from the faces that belong to a place, because a face with no
// place names no city and the city path is the one worth exercising.
//
//	faces  -> the first placeCount places, and every face that belongs to them
//	edges  -> every edge bounded by one of those faces, giving the TLIDs
//	featnames, addr -> every record for one of those TLIDs
//	place  -> the places the kept faces name
//	state  -> the one state the county is in
//
// Run it from the module root, with the county's files already downloaded:
//
//	go run internal/tigerfixture/build/generate.go
package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/poetic-systems/zipcity/internal/tigerfixture"
)

const (
	vintage = "tl_2025"
	county  = "01001" // Autauga County, Alabama
	state   = "01"

	// Places are the knob on how much of the county comes along: every face
	// of a kept place is kept, and the edges, names and address ranges follow
	// from those. Six is enough to reach a street whose two sides carry
	// different ZIP Codes without pulling in a county's worth of geometry.
	placeCount = 6

	cachedir = "data/us_census_tiger"
	destdir  = "internal/ustigerline/testdata/us_census_tiger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "tigerfixture: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	places, faces, err := chooseFaces()
	if err != nil {
		return err
	}

	if _, err := subset("faces", county, func(fields map[string]any) bool {
		return faces[text(fields["TFID"])]
	}); err != nil {
		return err
	}

	tlids := map[string]bool{}
	if _, err := subset("edges", county, func(fields map[string]any) bool {
		if !faces[text(fields["TFIDL"])] && !faces[text(fields["TFIDR"])] {
			return false
		}
		tlids[text(fields["TLID"])] = true
		return true
	}); err != nil {
		return err
	}

	for _, set := range []string{"featnames", "addr"} {
		if _, err := subset(set, county, func(fields map[string]any) bool {
			return tlids[text(fields["TLID"])]
		}); err != nil {
			return err
		}
	}

	if _, err := subset("place", state, func(fields map[string]any) bool {
		return places[text(fields["PLACEFP"])]
	}); err != nil {
		return err
	}

	_, err = subset("state", "us", func(fields map[string]any) bool {
		return text(fields["STATEFP"]) == state
	})
	return err
}

// chooseFaces picks the places to keep and the faces that belong to them,
// taking the lowest place codes so that a rebuild against the same vintage
// produces the same fixture.
func chooseFaces() (places, faces map[string]bool, err error) {
	byPlace := map[string][]string{}
	if err := tigerfixture.Scan(source("faces", county), func(fields map[string]any) {
		placefp, tfid := text(fields["PLACEFP"]), text(fields["TFID"])
		if placefp != "" && tfid != "" {
			byPlace[placefp] = append(byPlace[placefp], tfid)
		}
	}); err != nil {
		return nil, nil, err
	}

	codes := slices.Sorted(maps.Keys(byPlace))
	if len(codes) > placeCount {
		codes = codes[:placeCount]
	}

	places, faces = map[string]bool{}, map[string]bool{}
	for _, code := range codes {
		places[code] = true
		for _, tfid := range byPlace[code] {
			faces[tfid] = true
		}
	}
	if len(faces) == 0 {
		return nil, nil, fmt.Errorf("no face in %s names a place", county)
	}
	return places, faces, nil
}

func subset(set, prefix string, keep tigerfixture.Keep) (int, error) {
	src, dst := source(set, prefix), filepath.Join(destdir, set, name(set, prefix))
	kept, err := tigerfixture.Subset(src, dst, keep)
	if err != nil {
		return 0, err
	}
	fmt.Printf("%-10s %6d records -> %s\n", set, kept, dst)
	return kept, nil
}

func source(set, prefix string) string {
	return filepath.Join(cachedir, set, name(set, prefix))
}

func name(set, prefix string) string {
	return fmt.Sprintf("%s_%s_%s.zip", vintage, prefix, set)
}

// text renders a DBF field as the readers do, so that a fixture predicate and
// the code under test agree about what a field says.
func text(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
