package conflictedabrev

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/poetic-systems/zipcity/internal/bloomkeys"
)

type SourceValue struct {
	Source string
	Value  string
}

type KnownConflict struct {
	Abrev    string
	Possible []SourceValue
}

var knownConflicts = []KnownConflict{
	{
		Abrev: "CT",
		Possible: []SourceValue{
			{
				Source: "StreetType",
				Value:  "COURT",
			},
			{
				Source: "Region",
				Value:  "CONNECTICUT",
			},
		},
	},
	{
		Abrev: "MT",
		Possible: []SourceValue{
			{"CityName", "MOUNT"},
			{"Region", "MONTANA"},
		},
	},
	{
		Abrev: "NE",
		Possible: []SourceValue{
			{"Directional", "NORTHEAST"},
			{"Region", "NEBRASKA"},
		},
	},
	{
		Abrev: "ST",
		Possible: []SourceValue{
			{"CityName", "SAINT"},
			{"StreetType", "STREET"},
			{"Other", "STATE"},
		},
	},
	{
		Abrev: "DR",
		Possible: []SourceValue{
			{"StreetType", "DRIVE"},
			{"Other", "DOCTOR"},
		},
	},
}

var conflictMap = maps.Collect(func(yield func(string, KnownConflict) bool) {
	for _, kc := range knownConflicts {
		if !yield(kc.Abrev, kc) {
			return
		}
		for _, sv := range kc.Possible {
			if !yield(sv.Value, kc) {
				return
			}
		}
	}
})

func Info(token string) *KnownConflict {
	kc, ok := conflictMap[token]
	if !ok {
		return nil
	}

	return &KnownConflict{
		Abrev:    kc.Abrev,
		Possible: kc.Possible[:],
	}
}

// Keys yields the filter key of each word pair in name that touches a
// conflicted word: the word with the one before it, and the word with the
// one after it. The generator adds these to the conflicted-abbreviation
// filter for every street and city name it reads; Readings asks for the same
// keys, so the two cannot drift apart.
func Keys(name string) iter.Seq[string] {
	return func(yield func(string) bool) {
		words := strings.Fields(name)
		for i, w := range words {
			if Info(w) == nil {
				continue
			}
			if i > 0 && !yield(bloomkeys.KeyConflictedAbbreviationPreviousWord(w, words[i-1])) {
				return
			}
			if i < len(words)-1 && !yield(bloomkeys.KeyConflictedAbbreviationNextWord(w, words[i+1])) {
				return
			}
		}
	}
}

// Tester is the part of a bloom filter Readings needs.
type Tester interface {
	TestString(string) bool
}

// Readings returns the readings of a conflicted word (the abbreviation itself,
// then each spelling it may stand for) that appear beside prev or next in the
// names the filter was built from. Either neighbor may be empty when the word
// starts or ends its name. The result is empty for a word that is not
// conflicted, and for one whose neighbors the filter has never seen it with.
//
// Each reading is a bloom filter hit, so it is wrong at the filter's false
// positive rate; more than one reading is a real ambiguity or a false
// positive, and the filter cannot say which.
func Readings(f Tester, word, prev, next string) []string {
	kc := Info(bloomkeys.Normalize(word))
	if kc == nil {
		return nil
	}
	candidates := []string{kc.Abrev}
	for _, sv := range kc.Possible {
		candidates = append(candidates, sv.Value)
	}
	candidates = slices.Compact(candidates)

	var seen []string
	for _, c := range candidates {
		if (prev != "" && f.TestString(bloomkeys.KeyConflictedAbbreviationPreviousWord(c, prev))) ||
			(next != "" && f.TestString(bloomkeys.KeyConflictedAbbreviationNextWord(c, next))) {
			seen = append(seen, c)
		}
	}
	return seen
}
