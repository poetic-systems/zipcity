// Package featnames renders the street name a TIGER FEATNAMES record spells
// out in parts.
//
// A record does carry a rendered name, in FULLNAME, but it is one abbreviated
// rendering and is not always complete, so the name is built here from the
// base name and the four codes around it — PREQUAL, PRETYP, SUFTYP, SUFQUAL —
// plus the two directional fields.
//
// The rows those codes name are Appendix D of the Census Bureau's technical
// documentation and live in addresstables. What is here is the lookup and the
// assembly order, which is what a reader of the file needs and what the table
// does not say.
package featnames

import (
	"maps"
	"strings"

	table "github.com/poetic-systems/addresstables/ustigerfile/featuretypes"
	"github.com/poetic-systems/zipcity/internal/ustigerline/directionals"
	"github.com/poetic-systems/zipcity/internal/ustigerline/fieldutil"
	"github.com/poetic-systems/zipcity/internal/ustigerline/qualifiers"
)

// FeatnameInfo is one Appendix D row. Rows are uppercase, as the whole of
// addresstables is; the published table prints them in title case.
type FeatnameInfo = table.FeatureType

var featnameMap = maps.Collect(func(yield func(string, FeatnameInfo) bool) {
	for f := range table.All() {
		if !yield(f.Code, f) {
			return
		}
	}
})

func ExpandFeatureName(attr map[string]any) string {
	isSpanish := false

	base := ""
	// attr['NAME'] will contain the text between all prefix and suffix values
	rawname, ok := attr["NAME"]
	if ok {
		base = fieldutil.AsString(rawname)
	}

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
		suffixtype = suffixInfo.Full
	}

	// handle directionals

	prefixdirectional := ""
	// attr['PREDIRABRV'] will contain a string abreviation for any predirectional
	rawpdir, ok := attr["PREDIRABRV"]
	if ok {
		pdir := fieldutil.AsString(rawpdir)
		if len(pdir) > 0 {
			pd := directionals.Expand(pdir, isSpanish)
			prefixdirectional = pd
		}
	}

	suffixdirectional := ""
	// attr['SUFDIRABRV'] will contain a string abreviation for any postdirectional
	rawsdir, ok := attr["SUFDIRABRV"]
	if ok {
		sdir := fieldutil.AsString(rawsdir)
		if len(sdir) > 0 {
			// FIXME: Exactly what we do with directionals is TBD.
			// See https://github.com/poetic-systems/zipcity/issues/4.
			// Expanding, abbreviating, and aliasing (duplicating without the directional)
			// all have significant concerns.
			sd := directionals.Expand(sdir, isSpanish)
			suffixdirectional = sd
		}
	}

	// According to USPS Pub 28 Puerto Rican addresses begin with the street type.
	// Additionally, directional prefixes are noted as rare (and "Ó " is not a
	// valid directional prefix.)
	// https://www2.census.gov/geo/tiger/rd_2ktiger/tgrrd2k.pdf lists "Ó" as one
	// of the characters that it previously used square brackets to indicate.
	// On that basis, the large number of streets in puerto rico starting with
	// "Ó " are believed to result from the migration of pre-2000 ASCII-to-UTF-8
	// diacritical encodings in Puerto Rican/Spanish street records, which persist
	// as literal strings in annual TIGER/Line roll-forwards.
	base, _ = strings.CutPrefix(base, "Ó ")

	// TODO: determine if we need to uppercase and remove diacritics preemptively

	// The full concatenation order in TIGER files is:
	//   Prefix Qualifier (e.g., Old, New)
	//   Prefix Directional (e.g., North, East)
	//   Prefix Type (e.g., State Route, County Road)
	//   Base Name
	//   Suffix Type
	//   Suffix Directional
	//   Suffix Qualifier

	return strings.Join([]string{
		prefixqualifier,
		prefixdirectional,
		prefixtype,
		base,
		suffixtype,
		suffixdirectional,
		suffixqualifier,
	}, " ")
}
