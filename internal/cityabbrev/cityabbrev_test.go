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
	} {
		got := Match(tc.abbr, names)
		if len(got) == 0 || got[0].City != tc.want || got[0].Rule != tc.rule {
			t.Errorf("Match(%q) = %+v, want first %s via %s", tc.abbr, got, tc.want, tc.rule)
		}
	}
}
