package bloomkeys

import (
	"slices"
	"testing"
)

func TestStreetSpellingsKeepsTheGivenSpellingFirst(t *testing.T) {
	for _, street := range []string{"CALLE LOIZA", "CLL LOIZA", "FOX PARK DR", "MAIN"} {
		got := StreetSpellings(street)
		if len(got) == 0 || got[0] != street {
			t.Errorf("StreetSpellings(%q) = %q, want the given spelling first", street, got)
		}
	}
}

func TestStreetSpellingsSubstitutesTheLeadingType(t *testing.T) {
	for _, tc := range []struct {
		street string
		want   string
	}{
		{"CALLE LOIZA", "CLL LOIZA"},
		{"CLL LOIZA", "CALLE LOIZA"},
		{"AVENIDA ASHFORD", "AVE ASHFORD"},
		{"CARRETERA CAIMITO", "CARR CAIMITO"},
		{"RIO PIEDRAS", "RÍO PIEDRAS"},
		{"CALLEJON LA FE", "CALLEJÓN LA FE"},
	} {
		if got := StreetSpellings(tc.street); !slices.Contains(got, tc.want) {
			t.Errorf("StreetSpellings(%q) = %q, want it to include %q", tc.street, got, tc.want)
		}
	}
}

// BOULEVARD is the one type TIGER writes three ways, so a caller who uses any
// of them has to reach the other two.
func TestStreetSpellingsCoversAWholeGroup(t *testing.T) {
	got := StreetSpellings("BOULEVARD DE LOS ARBOLES")
	for _, want := range []string{"BLVD DE LOS ARBOLES", "BULEVAR DE LOS ARBOLES"} {
		if !slices.Contains(got, want) {
			t.Errorf("StreetSpellings(%q) = %q, want it to include %q", "BOULEVARD DE LOS ARBOLES", got, want)
		}
	}
}

// A trailing type is an English suffix the caller has already spelled the way
// the data holds it. Rewriting it would buy a filter probe that cannot match.
func TestStreetSpellingsLeavesATrailingTypeAlone(t *testing.T) {
	for _, street := range []string{"FOX PARK DR", "BLUE RIDGE AVE", "W 9000 S"} {
		if got := StreetSpellings(street); len(got) != 1 {
			t.Errorf("StreetSpellings(%q) = %q, want the input unchanged", street, got)
		}
	}
}

// Every extra spelling is another filter probe against a mutually exclusive
// key, and each probe carries the filter's 1% false positive rate. Keeping the
// count small is what keeps the compounded rate near it.
func TestStreetSpellingsStaysCheap(t *testing.T) {
	for _, group := range streetTypeGroups {
		for _, spelling := range group {
			if got := StreetSpellings(spelling + " EXAMPLE"); len(got) > 3 {
				t.Errorf("StreetSpellings(%q ...) returned %d spellings: %q", spelling, len(got), got)
			}
		}
	}
}

// The caller's own casing must not decide whether the substitution happens.
func TestStreetSpellingsIgnoresCase(t *testing.T) {
	if got := StreetSpellings("Calle Loiza"); !slices.Contains(got, "CLL LOIZA") {
		t.Errorf("StreetSpellings(%q) = %q, want it to include %q", "Calle Loiza", got, "CLL LOIZA")
	}
}
