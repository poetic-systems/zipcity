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
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/poetic-systems/addresstables/streetsuffixes"
	"github.com/poetic-systems/addresstables/ustigerfile/featuretypes"
	"github.com/poetic-systems/zipcity/internal/ustigerline/directionals"
	"github.com/poetic-systems/zipcity/internal/ustigerline/fieldutil"
	"github.com/poetic-systems/zipcity/internal/ustigerline/qualifiers"
)

// We want to look for indicators of a Spanish street name
var spanishMarkerRegex = regexp.MustCompile(`(?i)\b(de|del|las|los|san|santa|don|doña|val|valle|villa|plaza)\b`)

var spanishPrefixOverrides = map[string]string{
	"AVE": "AVENIDA",
}

// Pub 28 spells out no interstate form, so the key follows what
// go-projectusat/pkg/highways produces for one: INTERSTATE 5, not Appendix D's
// INTERSTATE HIGHWAY 5. See #37.
var pub28PrefixOverrides = map[string]string{
	"I-": "INTERSTATE",
}

// These exceptions often mean something else when they occur in the base street name
// rather than a coded prefix or suffix type
var basePartExceptions = []string{
	"ST", // it often means Saint
	"DR", // it often means Doctor
}

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

// English rows are yielded last so they win a Short collision (AVE is both 124
// AVENIDA and 125 AVENUE); ApplySpanishPrefixOverrides picks the Spanish reading.
var featnameShortMap = maps.Collect(func(yield func(string, FeatnameInfo) bool) {
	for _, spanish := range []bool{true, false} {
		for f := range featuretypes.All() {
			if f.Spanish == spanish && !yield(f.Short, f) {
				return
			}
		}
	}
})

// The longest Short is several words ("BUREAU OF INDIAN AFFAIRS HIGHWAY"), so
// the base name is matched as windows of whole tokens, longest window first.
var featnameShortMaxWords = slices.Max(slices.Collect(func(yield func(int) bool) {
	for f := range featuretypes.All() {
		if !yield(strings.Count(f.Short, " ") + 1) {
			return
		}
	}
}))

var pub28StreetSuffixes = maps.Collect(func(yield func(string, string) bool) {
	for s := range streetsuffixes.All() {
		for _, a := range s.Alt {
			if !yield(a, s.Short) {
				return
			}
		}
	}
})

func ApplySpanishPrefixOverrides(info FeatnameInfo, full string, defaultSpanish bool) (string, bool) {
	// In Puerto Rico, where the legal language for street names is Spanish, we can assume
	// that an English prefix that collides with a Spanish one is a data ingestion error.
	// Otherwise, as a special case for "AVE", we check the rest of the street name text to
	// see if it looks like Spanish.
	short := strings.ToUpper(info.Short)
	if defaultSpanish || (short == "AVE" && spanishMarkerRegex.MatchString(full)) {
		spanish, ok := spanishPrefixOverrides[short]
		if ok {
			return spanish, true
		}
	}
	return info.Full, false
}

func Pub28FeatureName(attr map[string]any) string {
	rawsfp, ok := attr["__STATEFP"]
	sfp := fieldutil.AsString(rawsfp)

	base := ""
	// attr['NAME'] will contain the text between all prefix and suffix values
	rawname, ok := attr["NAME"]
	if ok {
		base = fieldutil.AsString(rawname)
	}
	base = strings.ToUpper(base)

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

	// The NAME may contain secondary "prefixes" that should be expanded per pub 28.
	// However, examples like MT ST HELENS RD mean we have to be careful of certain collisions
	// Matching whole tokens, longest window first, lets "CO RD" and other multi-word
	// feature name values expand while "ST" never matches inside "FOREST".
	baseparts := strings.Fields(base)
	expandedbaseparts := slices.Collect(func(yield func(string) bool) {
		for i := 0; i < len(baseparts); i++ {
			v := baseparts[i]
			for w := min(featnameShortMaxWords, len(baseparts)-i); w > 0; w-- {
				short := strings.Join(baseparts[i:i+w], " ")
				info, ok := featnameShortMap[short]
				// We skip replacing exceptions because they often mean something else when
				// they are in the base street name
				if ok && !slices.Contains(basePartExceptions, short) {
					v, _ = ApplySpanishPrefixOverrides(info, base, sfp == "72")
					i += w - 1
					break
				}
			}
			if !yield(v) {
				return
			}
		}
	})
	base = strings.Join(expandedbaseparts, " ")

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
		// The prefix type is rendered from the code's Full form rather than
		// copied from FULLNAME, and the difference is not cosmetic. Across
		// San Juan, Ponce, Mayagüez, Bayamón, Caguas and Guaynabo, FULLNAME
		// spells 26 distinct leading types: ten abbreviated in TIGER's own
		// vocabulary (CLL 39,050 times, AVE, CAM, QBDA, CARR, SEC, PSO, BLVD,
		// PLZ, CNL) and the rest written out (EXPRESO, CALLEJÓN, AUTOPISTA,
		// RÍO, ...), with no rule that says which a given record will use.
		// Project US@ (p. 26) forbids abbreviating a street name, so a caller
		// sends CALLE LOIZA; a key copied from FULLNAME would hold CLL LOIZA
		// and the lookup would miss a street that is in the data. See #25.
		//
		// "72" is Puerto Rico, where the legal language for street names is Spanish
		prefixtype, _ = ApplySpanishPrefixOverrides(prefixInfo, base, sfp == "72")
		if p28, ok := pub28PrefixOverrides[prefixInfo.Short]; ok {
			prefixtype = p28
		}
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
			prefixdirectional = directionals.Pub28(pdir)
		}
	}

	suffixdirectional := ""
	// attr['SUFDIRABRV'] will contain a string abbreviation for any postdirectional
	rawsdir, ok := attr["SUFDIRABRV"]
	if ok {
		sdir := fieldutil.AsString(rawsdir)
		if len(sdir) > 0 {
			suffixdirectional = directionals.Pub28(sdir)
		}
	}

	// if len(prefixtype) > 0 && prefixtype != prefixInfo.Full {
	// 	fmt.Printf("(Spanish?: %t): %s %s | %s | %v\n", prefixInfo.Spanish, prefixtype, prefixInfo.Full, base, attr["FULLNAME"])
	// }

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
