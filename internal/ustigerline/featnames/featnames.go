// Package featnames renders the street name a TIGER FEATNAMES record spells
// out in parts.
//
// A record does carry a rendered name, in FULLNAME, but it is one abbreviated
// rendering and is not always complete, so the name is built here from the
// base name and the six fields around it: PREQUAL, PREDIRABRV, PRETYP, SUFTYP,
// SUFDIRABRV, and SUFQUAL.
//
// The rows those codes name are Appendix D of the Census Bureau's technical
// documentation and live in addresstables. What is here is the lookup and the
// assembly order, which is what a reader of the file needs and what the table
// does not say.
package featnames

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/streetsuffixes"
	"github.com/poetic-systems/addresstables/ustigerfile/featuretypes"
	"github.com/poetic-systems/zipcity/internal/ustigerline/directionals"
	"github.com/poetic-systems/zipcity/internal/ustigerline/fieldutil"
	"github.com/poetic-systems/zipcity/internal/ustigerline/qualifiers"
)

// FeatnameInfo is one Appendix D row. Rows are uppercase, as the whole of
// addresstables is; the published table prints them in title case.
type FeatnameInfo = featuretypes.FeatureType

var featnameMap = maps.Collect(func(yield func(string, FeatnameInfo) bool) {
	for f := range featuretypes.All() {
		if !yield(f.Code, f) {
			return
		}
	}
})

// The TIGER files seem to frequently code prefix types in Puerto Rico as
// English even though Project US@ says that the prefix position is for
// Spanish street types. It is relatively rare to have a English street type
// in a prefix position so frequent occurances in Puerto Rico seem very strange.
var spanishCollisions = maps.Collect(func(yield func(string, FeatnameInfo) bool) {
	collect := make(map[string][]FeatnameInfo)
	for f := range featuretypes.All() {
		c, ok := collect[f.Short]
		if !ok {
			c = make([]FeatnameInfo, 0)
		}
		c = append(c, f)
		collect[f.Short] = c
	}

	for k, c := range collect {
		if len(c) > 1 {
			for _, f := range c {
				if f.Spanish {
					if !yield(k, f) {
						return
					}
				}
			}
		}
	}
})

var pub28StreetSuffixes = maps.Collect(func(yield func(string, string) bool) {
	for s := range streetsuffixes.All() {
		for _, a := range s.Alt {
			if !yield(a, s.Short) {
				return
			}
		}
	}
})

func Pub28FeatureName(attr map[string]any) string {
	isSpanish := false
	sfp, ok := attr["__STATEFP"]
	if ok && sfp == "72" {
		// 72 is Puerto Rico, default to Spanish
		fmt.Printf("State FIPS code is 72 (Puerto Rico). Defaulting to Spanish.\n")
		isSpanish = true
	}

	base := ""
	// attr['NAME'] will contain the text between all prefix and suffix values
	rawname, ok := attr["NAME"]
	if ok {
		base = fieldutil.AsString(rawname)
	}
	base = strings.ToUpper(base)

	prefixqualifier := ""
	// attr['PREQUAL'] will contain a numeric code for qualifiers
	rawpq, ok := attr["PREQUAL"]
	if ok {
		pq := fieldutil.AsString(rawpq)
		prefixQualifierInfo, err := qualifiers.Info(pq)
		if err == nil && prefixQualifierInfo.Prefix {
			prefixqualifier = prefixQualifierInfo.Full
		}
	}

	prefixtype := ""
	// attr['PRETYP'] will contain a numeric code for feature names
	pt := ""
	rawpt, ok := attr["PRETYP"]
	if ok {
		pt = fieldutil.AsString(rawpt)
	}
	prefixInfo, ok := featnameMap[pt]
	if ok && prefixInfo.Prefix {
		// if sfp == "72" {
		// 	// 72 is Puerto Rico - an English prefix should be very uncommon there because
		// 	// the prefix position is proper in Spanish; an English word in that position
		// 	// would just be confusing. Unfortunately, the TIGER files use the English
		// 	// prefix type code frequently for these streets. So we are taking the liberty
		// 	// of overriding them.
		// 	c, ok := spanishCollisions[prefixInfo.Short]
		// 	if ok {
		// 		prefixInfo = c
		// 	}
		// }
		if prefixInfo.Spanish {
			isSpanish = true
		}
		prefixtype = prefixInfo.Full
	}

	suffixqualifier := ""
	// attr['SUFQUAL'] will contain a numeric code for qualifiers
	rawsq, ok := attr["SUFQUAL"]
	if ok {
		sq := fieldutil.AsString(rawsq)
		suffixQualifierInfo, err := qualifiers.Info(sq)
		if err == nil && suffixQualifierInfo.Suffix {
			suffixqualifier = suffixQualifierInfo.Full
		}
	}

	suffixtype := ""
	// attr['SUFTYP'] will contain a numeric code for feature names
	st := ""
	rawst, ok := attr["SUFTYP"]
	if ok {
		st = fieldutil.AsString(rawst)
	}
	suffixInfo, ok := featnameMap[st]
	if ok && suffixInfo.Suffix {
		if suffixInfo.Spanish {
			isSpanish = true
		}

		// only abbreviate known pub28 suffixes
		p28, ok := pub28StreetSuffixes[suffixInfo.Full]
		if ok {
			suffixtype = p28
		} else {
			suffixtype = suffixInfo.Full
		}
	}

	// handle directionals

	prefixdirectional := ""
	// attr['PREDIRABRV'] will contain a string abbreviation for any predirectional
	rawpdir, ok := attr["PREDIRABRV"]
	if ok {
		pdir := fieldutil.AsString(rawpdir)
		if len(pdir) > 0 {
			pd := directionals.Expand(pdir, isSpanish)
			prefixdirectional = directionals.Pub28(pd)
		}
	}

	suffixdirectional := ""
	// attr['SUFDIRABRV'] will contain a string abbreviation for any postdirectional
	rawsdir, ok := attr["SUFDIRABRV"]
	if ok {
		sdir := fieldutil.AsString(rawsdir)
		if len(sdir) > 0 {
			sd := directionals.Expand(sdir, isSpanish)
			suffixdirectional = directionals.Pub28(sd)
		}
	}

	// According to USPS Pub 28 Puerto Rican addresses begin with the street type.
	// Additionally, directional prefixes are noted as rare (and "Ó " is not a
	// valid directional prefix.)
	// https://www2.census.gov/geo/tiger/rd_2ktiger/tgrrd2k.pdf lists "Ó" as one
	// of the characters that it previously used square brackets to indicate.
	// On that basis, the roughly 77 street records in puerto rico starting with
	// "Ó " are believed to result from the migration of pre-2000 ASCII-to-UTF-8
	// diacritical encodings in Puerto Rican/Spanish street records, which persist
	// as literal strings in annual TIGER/Line roll-forwards.
	base, _ = strings.CutPrefix(base, "Ó ")

	if sfp == "72" && len(prefixtype) > 0 && !isSpanish {
		fmt.Printf("(Spanish?: %t): %v\n", isSpanish, attr)
	}

	// The full concatenation order in TIGER files is:
	//   Prefix Qualifier (e.g., Old, New)
	//   Prefix Directional (e.g., North, East)
	//   Prefix Type (e.g., State Route, County Road)
	//   Base Name
	//   Suffix Type
	//   Suffix Directional
	//   Suffix Qualifier

	return fieldutil.JoinNonEmpty([]string{
		prefixqualifier,
		prefixdirectional,
		prefixtype,
		base,
		suffixtype,
		suffixdirectional,
		suffixqualifier,
	}, " ")
}
