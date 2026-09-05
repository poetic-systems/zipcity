package areazip_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/areazip"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

func TestSole(t *testing.T) {
	stateZips := map[string][]string{
		"AS": {"96799"},
		"GU": {"96910", "96913", "96915"},
		"AK": {"99827", "99801"},
		"MP": {"96950", "96951", "96952"},
	}
	countyZips := map[string][]string{
		"MP085": {"96950"},
		"MP100": {"96950", "96951"},
		"AK100": {"99827", "81087"},
	}

	for _, tc := range []struct {
		name       string
		usps       string
		countyfips string
		want       string
	}{
		{"county names one", "MP", "085", "96950"},
		{"county names several", "MP", "100", ""},
		{"county silent, territory names one", "AS", "010", "96799"},
		{"county silent, territory names several", "GU", "010", ""},
		{"county names several, state names several", "AK", "100", ""},
		{"nothing named anywhere", "ZZ", "999", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := areazip.Sole(stateZips, countyZips, tc.usps, tc.countyfips)
			if got != tc.want {
				t.Errorf("Sole(%q, %q) = %q, want %q", tc.usps, tc.countyfips, got, tc.want)
			}
		})
	}
}

// A county's own ZIP Codes take precedence over its state's even when the
// state names exactly one, because the county is the narrower statement.
func TestSolePrefersTheCounty(t *testing.T) {
	stateZips := map[string][]string{"MP": {"96950"}}
	countyZips := map[string][]string{"MP100": {"96951"}}

	if got := areazip.Sole(stateZips, countyZips, "MP", "100"); got != "96951" {
		t.Errorf("Sole = %q, want the county's 96951", got)
	}
}

func sides(zips ...[]string) map[string]*ustigerline.StreetSide {
	out := make(map[string]*ustigerline.StreetSide, len(zips))
	for i, z := range zips {
		out[string(rune('a'+i))] = &ustigerline.StreetSide{Zips: z}
	}

	return out
}

func TestContradicted(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sides   map[string]*ustigerline.StreetSide
		areazip string
		want    bool
	}{
		{"no sides at all", sides(), "96799", false},
		{"every side agrees", sides([]string{"96799"}, []string{"96799"}), "96799", false},
		{"sides name nothing", sides(nil, []string{}), "96799", false},
		{"agreement and silence together", sides([]string{"96799"}, nil), "96799", false},
		{"one side dissents", sides([]string{"96799"}, []string{"96950"}), "96799", true},
		{"a side names the area's ZIP Code and another", sides([]string{"96799", "96950"}), "96799", true},
		{"every side dissents", sides([]string{"96950"}, []string{"96951"}), "96799", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := areazip.Contradicted(tc.sides, tc.areazip); got != tc.want {
				t.Errorf("Contradicted = %v, want %v", got, tc.want)
			}
		})
	}
}

// The Haines Borough shape of poetic-systems/zipcity#16: 344 sides agree, 3
// say nothing, and one published row carries a ZIP Code 2,300 miles away. The
// strict rule refuses the whole area on that one row, and this test is here so
// that anyone loosening it has to say so out loud.
func TestContradictedRefusesOnASingleBadRow(t *testing.T) {
	agreeing := make([][]string, 0, 348)
	for range 344 {
		agreeing = append(agreeing, []string{"99827"})
	}
	for range 3 {
		agreeing = append(agreeing, nil)
	}
	agreeing = append(agreeing, []string{"81087"})

	if !areazip.Contradicted(sides(agreeing...), "99827") {
		t.Error("Contradicted = false; one dissenting side out of 348 must still refuse the area")
	}
}
