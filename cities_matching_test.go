package zipcity_test

import (
	"testing"

	"github.com/poetic-systems/zipcity"
)

func TestCitiesMatching(t *testing.T) {
	zc, err := zipcity.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ state, abbr, want string }{
		{"NY", "NY", "NEW YORK"},
		{"CA", "LA", "LOS ANGELES"},
		{"NY", "BKLYN", "BROOKLYN"},
		{"CA", "GLNDALE", "GLENDALE"},
	} {
		got := zc.CitiesMatching(tc.state, tc.abbr)
		t.Logf("%s %s -> %v", tc.state, tc.abbr, got[:min(len(got), 5)])
		found := false
		for _, c := range got {
			found = found || c.City == tc.want
		}
		if !found {
			t.Errorf("CitiesMatching(%s,%s) missing %s", tc.state, tc.abbr, tc.want)
		}
	}
}

func BenchmarkCitiesMatching(b *testing.B) {
	zc, err := zipcity.New()
	if err != nil {
		b.Fatal(err)
	}
	zc.CitiesMatching("CA", "GLNDALE") // build the index
	b.ResetTimer()
	for range b.N {
		zc.CitiesMatching("CA", "GLNDALE")
	}
}

func BenchmarkCitiesMatchingFuzzyFallback(b *testing.B) {
	zc, err := zipcity.New()
	if err != nil {
		b.Fatal(err)
	}
	zc.CitiesMatching("CA", "GLNEDALE") // build the index
	b.ResetTimer()
	for range b.N {
		zc.CitiesMatching("CA", "GLNEDALE") // transposition: only Fuzzy reaches it
	}
}
