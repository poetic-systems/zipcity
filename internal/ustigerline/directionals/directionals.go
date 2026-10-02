// Package directionals reads the PREDIRABRV and SUFDIRABRV fields of a TIGER
// FEATNAMES record.
//
// The rows live in addresstables, where they are the Publication 28 set: TIGER
// writes the same eight English and eight Spanish directionals. What is here
// is the expansion a reader of those two fields needs, keyed both ways — a
// file writes the abbreviation, but the same lookup has to leave a word that
// arrives already expanded alone.
//
// Which of the two sets to read is the caller's decision rather than the
// field's. N is NORTH on a mainland street and NORTE on a Puerto Rican one,
// and only the feature type beside it says which.
//
// Rows are uppercase, as the whole of addresstables is; the published tables
// print them in title case.
package directionals

import (
	"iter"
	"maps"

	table "github.com/poetic-systems/addresstables/directionals"
)

// fullMap keys both the abbreviation and the word to the word.
func fullMap(rows iter.Seq[table.Directional]) map[string]string {
	return maps.Collect(func(yield func(string, string) bool) {
		for d := range rows {
			if !yield(d.Short, d.Full) {
				return
			}
			if !yield(d.Full, d.Full) {
				return
			}
		}
	})
}

// shortMap keys both the word and the abbreviation to the abbreviation.
func shortMap(rows iter.Seq[table.Directional]) map[string]string {
	return maps.Collect(func(yield func(string, string) bool) {
		for d := range rows {
			if !yield(d.Full, d.Short) {
				return
			}
			if !yield(d.Short, d.Short) {
				return
			}
		}
	})
}

var (
	spanishFullMap  = fullMap(table.Spanish())
	spanishShortMap = shortMap(table.Spanish())
	englishFullMap  = fullMap(table.English())
	englishShortMap = shortMap(table.English())
	infoTable       = maps.Collect(func(yield func(string, table.Directional) bool) {
		for d := range table.All() {
			if !yield(d.Full, d) {
				return
			}
			if !yield(d.Short, d) {
				return
			}
		}
	})
)

func Expand(abrev string, isSpanish bool) string {
	if isSpanish {
		r, ok := spanishFullMap[abrev]
		if ok {
			return r
		}
		return abrev
	}

	r, ok := englishFullMap[abrev]
	if ok {
		return r
	}
	return abrev
}

func Abbreviate(full string, isSpanish bool) string {
	if isSpanish {
		r, ok := spanishShortMap[full]
		if ok {
			return r
		}
		return full
	}

	r, ok := englishShortMap[full]
	if ok {
		return r
	}
	return full
}

// Pub28 abbreviates src to its Publication 28 form without crossing
// languages. USPS Pub 28 §295 and Project US@ p.25 both forbid translating a
// Puerto Rico directional between Spanish and English (NO must stay NO, not
// become NW), so a Spanish row's own Short is returned rather than its
// English counterpart's.
func Pub28(src string) string {
	d, ok := infoTable[src]
	if ok {
		return d.Short
	}

	// the unmatched case just returns src unchanged so we don't
	// require error handling
	return src
}
