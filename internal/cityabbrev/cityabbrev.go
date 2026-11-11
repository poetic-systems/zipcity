// Package cityabbrev ranks the city names of one state as readings of an
// abbreviation, for a caller who has an abbreviated city and no ZIP Code to
// look it up by. See poetic-systems/zipcity#83 and
// PortobelloAuth/go-projectusat#262.
//
// Each Rule is one way a person shortens a name. A candidate is the name plus
// the strongest rule that reached it, so a caller can weigh "NY is the
// initialism of NEW YORK" differently from "GLNDALE is within one edit of
// GLENDALE". This is exploration: the rules and their order are a starting
// point to argue with, not settled.
package cityabbrev

import (
	"slices"
	"strings"

	"github.com/hbollon/go-edlib"
)

// Rule is how an abbreviation reaches a name, strongest first.
type Rule int

const (
	Initialism  Rule = iota // NY -> NEW YORK, LA -> LOS ANGELES
	Prefix                  // SPRINGFLD is not one, SPRING is: SPRING -> SPRINGFIELD
	Subsequence             // GLNDL -> GLENDALE, BKLYN -> BROOKLYN: first letter kept, rest in order
	Fuzzy                   // GLNDALE -> GLENDALE, within a Damerau-Levenshtein threshold
)

var ruleNames = [...]string{"initialism", "prefix", "subsequence", "fuzzy"}

func (r Rule) String() string { return ruleNames[r] }

// Candidate is a city name and the strongest rule that reached it.
type Candidate struct {
	City  string
	Rule  Rule
	Score float32 // similarity in [0,1]; 1 for the exact rules
}

// MinFuzzy is the similarity below which Fuzzy does not offer a name.
const MinFuzzy = 0.8

// Match ranks names as readings of abbr: by rule, then score descending, then
// name. abbr and names are expected normalized (bloomkeys.City). A name equal
// to abbr is not an abbreviation of itself and is skipped.
func Match(abbr string, names []string) []Candidate {
	abbr = strings.TrimSpace(abbr)
	var out []Candidate
	for _, name := range names {
		if name == abbr || abbr == "" {
			continue
		}
		if c, ok := reach(abbr, name); ok {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b Candidate) int {
		switch {
		case a.Rule != b.Rule:
			return int(a.Rule) - int(b.Rule)
		case a.Score != b.Score:
			if a.Score > b.Score {
				return -1
			}
			return 1
		}
		return strings.Compare(a.City, b.City)
	})
	return out
}

func reach(abbr, name string) (Candidate, bool) {
	switch {
	case initialism(name) == abbr:
		return Candidate{name, Initialism, 1}, true
	case len(abbr) >= 3 && strings.HasPrefix(name, abbr):
		return Candidate{name, Prefix, 1}, true
	case len(abbr) >= 3 && subsequence(abbr, name):
		return Candidate{name, Subsequence, float32(len(abbr)) / float32(len(name))}, true
	}
	if s, err := edlib.StringsSimilarity(abbr, name, edlib.DamerauLevenshtein); err == nil && s >= MinFuzzy {
		return Candidate{name, Fuzzy, s}, true
	}
	return Candidate{}, false
}

func initialism(name string) string {
	words := strings.Fields(name)
	if len(words) < 2 {
		return ""
	}
	var b strings.Builder
	for _, w := range words {
		b.WriteByte(w[0])
	}
	return b.String()
}

// subsequence reports whether abbr starts with name's first letter and the
// rest of abbr's letters appear in name in order, spaces ignored.
func subsequence(abbr, name string) bool {
	if abbr[0] != name[0] {
		return false
	}
	rest := name[1:]
	for i := 1; i < len(abbr); i++ {
		j := strings.IndexByte(rest, abbr[i])
		if j < 0 {
			return false
		}
		rest = rest[j+1:]
	}
	return true
}
