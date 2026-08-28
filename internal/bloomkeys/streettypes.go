package bloomkeys

import "strings"

// Puerto Rico street names carry their type at the front — CALLE LOIZA, not
// LOIZA ST — and the TIGER data spells that leading type in a vocabulary that
// is neither the caller's nor the standard's.
//
// Project US@ page 26 says a street name MUST NOT be abbreviated, so a caller
// following the specification sends CALLE. The TIGER FULLNAME column holds CLL.
// Nothing is missing from the data: the street is indexed under a spelling no
// conforming caller will ever produce, and the lookup returns a definitive
// false for a real address.
//
// The abbreviations are also applied inconsistently. Across six municipios
// (San Juan, Ponce, Mayaguez, Bayamon, Caguas, Guaynabo) TIGER abbreviates CLL,
// CAM, QBDA, CARR, SEC, PSO, BLVD, PLZ and CNL, while writing EXPRESO,
// AUTOPISTA, CALLEJON, RIO, CANO, VIA, RAMAL, MARGINAL, PUENTE, BULEVAR and
// PASAJE out in full — and it keeps the diacritics on the ones it spells out.
// There is no rule a caller could apply in advance to guess which form is held.
//
// So the equivalence is a property of this dataset rather than of addresses,
// which is why it is resolved here instead of being pushed onto callers. A
// caller cannot discover it without reading the shapefiles.
//
// Each group is one street type and every spelling of it seen in the data or
// likely to arrive from a caller. Groups are deliberately absent for types with
// a single observed spelling — AUTOPISTA, VIA, RAMAL, MARGINAL, PUENTE, PASAJE —
// because an alias that never matches anything only costs a filter probe.
//
// VER, ENT and VIS are also left out. They appear a handful of times each and
// their expansions are a guess; a wrong guess here would trade a false negative
// for a false positive, which is the worse direction.
var streetTypeGroups = [][]string{
	{"CALLE", "CLL"},
	{"AVENIDA", "AVE"},
	{"CAMINO", "CAM"},
	{"QUEBRADA", "QBDA"},
	{"CARRETERA", "CARR"},
	{"SECTOR", "SEC"},
	{"PASEO", "PSO"},
	{"BOULEVARD", "BULEVAR", "BLVD"},
	{"PLAZA", "PLZ"},
	{"CANAL", "CNL"},
	{"EXPRESO", "EXPRESSWAY", "EXPY"},

	// Accented forms are folded here rather than by a general diacritic pass.
	// The leading type is a closed vocabulary, so RIO can be mapped to RÍO
	// without guessing. The rest of the name is not: nothing derives LOÍZA
	// from LOIZA, and both are present in TIGER as separate records. Folding
	// the name half needs the index rebuilt with folded keys — see
	// https://github.com/poetic-systems/zipcity/issues/1.
	{"RIO", "RÍO"},
	{"CALLEJON", "CALLEJÓN"},
	{"CANO", "CAÑO"},
}

// alternates maps a leading street type to the other spellings of that type.
var alternates = func() map[string][]string {
	m := make(map[string][]string, 32)
	for _, group := range streetTypeGroups {
		for _, spelling := range group {
			for _, other := range group {
				if other != spelling {
					m[spelling] = append(m[spelling], other)
				}
			}
		}
	}
	return m
}()

// StreetSpellings returns the spellings of a street name worth looking up, the
// given one first.
//
// Only a leading type is substituted. A trailing AVE is an English suffix on a
// mainland street and means something the caller already spelled the way the
// data holds it, so rewriting it would add a probe that cannot match.
//
// Every returned spelling costs one filter probe, and the probes are for
// mutually exclusive keys, so k of them raise the chance of a spurious true
// from 1% to roughly 1-0.99^k. That is the trade this function makes: at most
// three probes, against a class of false negative that currently swallows 71%
// of the street names in San Juan.
func StreetSpellings(street string) []string {
	spellings := []string{street}

	normalized := normalize(street)
	head, tail, found := strings.Cut(normalized, " ")
	if !found {
		return spellings
	}

	for _, other := range alternates[head] {
		spellings = append(spellings, other+" "+tail)
	}

	return spellings
}
