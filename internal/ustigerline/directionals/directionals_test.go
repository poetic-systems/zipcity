package directionals_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline/directionals"
)

func TestExpand(t *testing.T) {
	cases := []struct {
		In      string
		Want    string
		Spanish bool
	}{
		{"N", "North", false},
		{"S", "South", false},
		{"E", "East", false},
		{"W", "West", false},
		{"NE", "Northeast", false},
		{"NW", "Northwest", false},
		{"SE", "Southeast", false},
		{"SW", "Southwest", false},
		{"N", "Norte", true},
		{"S", "Sur", true},
		{"E", "Este", true},
		{"O", "Oeste", true},
		{"NE", "Noreste", true},
		{"NO", "Noroeste", true},
		{"SE", "Sudeste", true},
		{"SO", "Sudoeste", true},
	}

	for _, tc := range cases {
		got := directionals.Expand(tc.In, tc.Spanish)
		if got != tc.Want {
			t.Fatalf("Wanted: '%s' Got: '%s", tc.Want, got)
		}
	}
}
