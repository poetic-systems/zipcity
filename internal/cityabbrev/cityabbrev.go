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
	Zips  int     // ZIP Codes the name has been seen for: a proxy for how likely a person means it
}

// MinFuzzy is the similarity below which Fuzzy does not offer a name.
const MinFuzzy = 0.8

// Match ranks names as readings of abbr: by rule, then ZIP count descending
// (zips may be nil), then score descending, then name. abbr and names are
// expected normalized (bloomkeys.City). A name equal to abbr is not an
// abbreviation of itself and is skipped. The Fuzzy rule is the expensive one
// (an edit distance per name) and the weakest, so it only runs when the
// cheaper rules reached nothing.
func Match(abbr string, names []string, zips func(name string) int) []Candidate {
	abbr = strings.TrimSpace(abbr)
	if abbr == "" {
		return nil
	}
	var out []Candidate
	for _, name := range names {
		if name != abbr {
			if c, ok := reach(abbr, name); ok {
				out = append(out, c)
			}
		}
	}
	if len(out) == 0 {
		for _, name := range names {
			if name != abbr {
				if c, ok := fuzzy(abbr, name); ok {
					out = append(out, c)
				}
			}
		}
	}
	if zips != nil {
		for i := range out {
			out[i].Zips = zips(out[i].City)
		}
	}
	slices.SortFunc(out, func(a, b Candidate) int {
		switch {
		case a.Rule != b.Rule:
			return int(a.Rule) - int(b.Rule)
		case a.Zips != b.Zips:
			return b.Zips - a.Zips
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
		return Candidate{City: name, Rule: Initialism, Score: 1}, true
	case len(abbr) >= 3 && strings.HasPrefix(name, abbr):
		return Candidate{City: name, Rule: Prefix, Score: 1}, true
	case subsequence(abbr, name):
		return Candidate{City: name, Rule: Subsequence, Score: float32(len(abbr)) / float32(len(name))}, true
	}
	return Candidate{}, false
}

// fuzzy offers name when abbr is within MinFuzzy of it. A similarity of
// MinFuzzy needs the lengths within (1-MinFuzzy) of each other, which rules
// most names out before any edit distance is computed.
func fuzzy(abbr, name string) (Candidate, bool) {
	if d := len(abbr) - len(name); float32(max(d, -d)) > (1-MinFuzzy)*float32(max(len(abbr), len(name))) {
		return Candidate{}, false
	}
	if s, err := edlib.StringsSimilarity(abbr, name, edlib.DamerauLevenshtein); err == nil && s >= MinFuzzy {
		return Candidate{City: name, Rule: Fuzzy, Score: s}, true
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
