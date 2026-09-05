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
		{"N", "NORTH", false},
		{"S", "SOUTH", false},
		{"E", "EAST", false},
		{"W", "WEST", false},
		{"NE", "NORTHEAST", false},
		{"NW", "NORTHWEST", false},
		{"SE", "SOUTHEAST", false},
		{"SW", "SOUTHWEST", false},
		{"N", "NORTE", true},
		{"S", "SUR", true},
		{"E", "ESTE", true},
		{"O", "OESTE", true},
		{"NE", "NORESTE", true},
		{"NO", "NOROESTE", true},
		{"SE", "SUDESTE", true},
		{"SO", "SUDOESTE", true},
		// A word that arrives already expanded is left alone, and so is one
		// the table does not name.
		{"NORTH", "NORTH", false},
		{"NORTE", "NORTE", true},
		{"Q", "Q", false},
	}

	for _, tc := range cases {
		got := directionals.Expand(tc.In, tc.Spanish)
		if got != tc.Want {
			t.Errorf("Expand('%s', %t) wanted: '%s' got: '%s'", tc.In, tc.Spanish, tc.Want, got)
		}
	}
}

func TestAbbreviate(t *testing.T) {
	cases := []struct {
		In      string
		Want    string
		Spanish bool
	}{
		{"NORTH", "N", false},
		{"SOUTHWEST", "SW", false},
		{"NORTE", "N", true},
		{"OESTE", "O", true},
		{"NOROESTE", "NO", true},
		// An abbreviation that arrives abbreviated is left alone, and so is a
		// word the table does not name.
		{"SW", "SW", false},
		{"NO", "NO", true},
		{"Q", "Q", false},
	}

	for _, tc := range cases {
		got := directionals.Abbreviate(tc.In, tc.Spanish)
		if got != tc.Want {
			t.Errorf("Abbreviate('%s', %t) wanted: '%s' got: '%s'", tc.In, tc.Spanish, tc.Want, got)
		}
	}
}
