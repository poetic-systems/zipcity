package cityabbrev

import "testing"

func TestMatch(t *testing.T) {
	names := []string{"NEW YORK", "NEWARK", "GLENDALE", "GLEN ELLYN", "BROOKLYN", "SPRINGFIELD", "SPRING", "LOS ANGELES"}
	for _, tc := range []struct {
		abbr, want string
		rule       Rule
	}{
		{"NY", "NEW YORK", Initialism},
		{"LA", "LOS ANGELES", Initialism},
		{"GLNDL", "GLENDALE", Subsequence},
		{"BRKLYN", "BROOKLYN", Subsequence},
		{"SPRINGF", "SPRINGFIELD", Prefix},
		{"GLENDAL", "GLENDALE", Prefix},
		{"GLNEDALE", "GLENDALE", Fuzzy},
		{"BK", "BROOKLYN", Subsequence},
		{"BKYN", "BROOKLYN", Subsequence},
	} {
		got := Match(tc.abbr, names, nil)
		if len(got) == 0 || got[0].City != tc.want || got[0].Rule != tc.rule {
			t.Errorf("Match(%q) = %+v, want first %s via %s", tc.abbr, got, tc.want, tc.rule)
		}
	}
}

// With no rule reaching a name beyond the subsequence, the name seen for the
// most ZIP Codes is the likelier reading.
func TestMatchRanksByZips(t *testing.T) {
	zips := map[string]int{"BROOKLYN": 40, "BUCKLEY": 2, "BECKLEY": 5}
	names := []string{"BUCKLEY", "BROOKLYN", "BECKLEY"}
	got := Match("BK", names, func(n string) int { return zips[n] })
	if len(got) != 3 || got[0].City != "BROOKLYN" || got[1].City != "BECKLEY" || got[2].City != "BUCKLEY" {
		t.Errorf("Match(BK) = %+v, want BROOKLYN, BECKLEY, BUCKLEY by ZIP count", got)
	}
}
