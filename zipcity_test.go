package zipcity_test

import (
	"slices"
	"testing"

	"github.com/poetic-systems/zipcity"
)

var testData = []struct {
	Street string
	City   string
	State  string
	Zip    string
}{
	{"BLUE RIDGE DR", "MARTINEZ", "CA", "94553"},
	{
		Street: "W 200 N",
		City:   "NORTH SALT LAKE",
		State:  "UT",
		Zip:    "84054",
	},
	{
		Street: "CALLE LOIZA",
		City:   "San Juan",
		State:  "PR",
		Zip:    "00909",
	},
	{
		Street: "CALLE LOÍZA",
		City:   "San Juan",
		State:  "PR",
		Zip:    "00909",
	},
	{
		Street: "CALLE LOIZA",
		City:   "San Juan",
		State:  "PR",
		Zip:    "00913",
	},
	{
		Street: "CALLE LOÍZA",
		City:   "San Juan",
		State:  "PR",
		Zip:    "00913",
	},
	{
		Street: "W 200 N",
		City:   "NORTH SALT LAKE",
		State:  "UT",
		Zip:    "84054",
	},
	{

		Street: "W 9000 S",
		City:   "West Jordan",
		State:  "UT",
		Zip:    "84088",
	},
	{
		Street: "W Fox Park Dr",
		City:   "West Jordan",
		State:  "UT",
		Zip:    "84088",
	},
	// {
	// 	Street: "Fox Park Dr",
	// 	City:   "West Jordan",
	// 	State:  "UT",
	// 	Zip:    "84088",
	// },
	{
		Street: "W 9200 S",
		City:   "West Jordan",
		State:  "UT",
		Zip:    "84088",
	},
	{
		Street: "9200 S",
		City:   "West Jordan",
		State:  "UT",
		Zip:    "84088",
	},
	{
		Street: "Pleasant Hill Rd",
		City:   "Pleasant Hill",
		State:  "CA",
		Zip:    "94523",
	},
	{
		Street: "pleasant hill rd",
		City:   "pleasant hill",
		State:  "ca",
		Zip:    "94523",
	},
	{
		Street: "The Alameda",
		City:   "San Jose",
		State:  "CA",
		Zip:    "95126",
	},
	{
		Street: "The Alameda",
		City:   "San José",
		State:  "CA",
		Zip:    "95126",
	},
	{
		Street: "Chalan Tun Herman Pan",
		// TODO: Addresses and roads in the Northern Mariana Islands might deserve a
		// closer look. In particular, if someone puts "Saipan" as the city it
		// is not going to work - but the
		// [USPS Zip Locale Detail.xls](https://postalpro.usps.com/ZIP_Locale_Detail)
		// shows the "Physical City" field for 96950 as "Saipan" (though the "Locale"
		// lists a few different names.)
		City:  "DanDan",
		State: "MP",
		Zip:   "96950",
	},
}

type ZipAndCity struct {
	Zip  string
	City string
}

func TestCheckZipAndCity(t *testing.T) {
	testcases := slices.Collect(func(yield func(ZipAndCity) bool) {
		for _, td := range testData {
			if !yield(ZipAndCity{
				Zip:  td.Zip,
				City: td.City,
			}) {
				return
			}
		}
	})

	for _, tc := range testcases {
		found, err := zipcity.CheckZipAndCity(tc.Zip, tc.City)
		if err != nil {
			t.Fatalf("Error checking zip: '%s' and city: '%s': %s", tc.Zip, tc.City, err)
		}

		if !found {
			t.Fatalf("Expected zip: '%s' and city: '%s' to be found", tc.Zip, tc.City)
		}
	}
}

type ZipAndStreet struct {
	Zip    string
	Street string
}

func TestCheckZipAndStreet(t *testing.T) {
	testcases := slices.Collect(func(yield func(ZipAndStreet) bool) {
		for _, td := range testData {
			if !yield(ZipAndStreet{
				Zip:    td.Zip,
				Street: td.Street,
			}) {
				return
			}
		}
	})

	for _, tc := range testcases {
		found, err := zipcity.CheckZipAndStreet(tc.Zip, tc.Street)
		if err != nil {
			t.Fatalf("Error checking zip: '%s' and street: '%s': %s", tc.Zip, tc.Street, err)
		}

		if !found {
			t.Fatalf("Expected zip: '%s' and street: '%s' to be found", tc.Zip, tc.Street)
		}
	}
}

type CityStateAndStreet struct {
	City   string
	State  string
	Street string
}

func TestCheckCityStateAndStreet(t *testing.T) {
	testcases := slices.Collect(func(yield func(CityStateAndStreet) bool) {
		for _, td := range testData {
			if !yield(CityStateAndStreet{
				City:   td.City,
				State:  td.State,
				Street: td.Street,
			}) {
				return
			}
		}
	})

	for _, tc := range testcases {
		found, err := zipcity.CheckCityStateAndStreet(tc.City, tc.State, tc.Street)
		if err != nil {
			t.Fatalf("Error checking city: '%s' state: '%s' and street: '%s': %s", tc.City, tc.State, tc.Street, err)
		}

		if !found {
			t.Fatalf("Expected city: '%s' state: '%s' and street: '%s' to be found", tc.City, tc.State, tc.Street)
		}
	}
}

func TestMatchStreet(t *testing.T) {
	// A Bloom filter can only promise the keys it holds, so, like the
	// Check tests above, these assert presence and never absence: a
	// regeneration is free to answer yes to a variant that is not there.
	cases := []struct {
		Street string
		City   string
		State  string
		Zip    string
		Want   zipcity.Match
	}{
		{"W Fox Park Dr", "West Jordan", "UT", "84088", zipcity.Match{Exact: true}},
		// TIGER attaches the W that a caller who knows the street as Fox
		// Park Dr does not have (#4).
		{"Fox Park Dr", "West Jordan", "UT", "84088", zipcity.Match{Variants: []string{"W FOX PARK DR"}}},
		// A street already carrying a directional on one side is only tried
		// with one on the other.
		{"9200 S", "West Jordan", "UT", "84088", zipcity.Match{Exact: true}},
		{"M St", "Washington", "DC", "20002", zipcity.Match{Variants: []string{"M ST NE"}}},
	}

	for _, tc := range cases {
		got, err := zipcity.MatchZipAndStreet(tc.Zip, tc.Street)
		if err != nil {
			t.Fatal(err)
		}
		if !holds(got, tc.Want) {
			t.Fatalf("Expected zip: '%s' and street: '%s' to match %+v, got %+v", tc.Zip, tc.Street, tc.Want, got)
		}
	}
	city, err := zipcity.MatchCityStateAndStreet("Washington", "DC", "M St")
	if err != nil {
		t.Fatal(err)
	}
	if !holds(city, zipcity.Match{Variants: []string{"M ST NE", "M ST SE", "M ST NW", "M ST SW"}}) {
		t.Fatalf("Expected every quadrant of M ST in Washington DC, got %+v", city)
	}
}

// holds reports whether got carries everything want promises.
func holds(got, want zipcity.Match) bool {
	if want.Exact && !got.Exact {
		return false
	}
	for _, v := range want.Variants {
		if !slices.Contains(got.Variants, v) {
			return false
		}
	}
	return true
}

type StateAndCity struct {
	State string
	City  string
}

func TestCitiesKnownFor(t *testing.T) {
	testcases := []struct {
		Zip  string
		Want []StateAndCity
	}{
		{"09001", []StateAndCity{{"AE", "APO"}}},
		// Straddles a state line: each name under its own state, states in order
		{"42223", []StateAndCity{{"KY", "FORT CAMPBELL"}, {"KY", "FORT CAMPBELL NORTH"}, {"TN", "CLARKSVILLE"}}},
		{"00000", nil},
		{"2017", nil},
	}

	for _, tc := range testcases {
		var got []StateAndCity
		for state, city := range zipcity.CitiesKnownFor(tc.Zip) {
			got = append(got, StateAndCity{state, city})
		}
		if !slices.Equal(got, tc.Want) {
			t.Errorf("CitiesKnownFor(%q) = %v, want %v", tc.Zip, got, tc.Want)
		}
	}
}

// Every name read out for a code is one the zip-city filter was built from,
// so the two cannot disagree about what we have seen.
func TestCitiesKnownForAgreeWithCheckZipAndCity(t *testing.T) {
	for _, td := range testData {
		n := 0
		for _, city := range zipcity.CitiesKnownFor(td.Zip) {
			n++
			found, err := zipcity.CheckZipAndCity(td.Zip, city)
			if err != nil || !found {
				t.Errorf("CitiesKnownFor(%q) yielded %q, CheckZipAndCity = %v, %v", td.Zip, city, found, err)
			}
		}
		if n == 0 {
			t.Errorf("CitiesKnownFor(%q) yielded nothing, %q is in the filter", td.Zip, td.City)
		}
	}
}

func TestZipsKnownFor(t *testing.T) {
	got := slices.Collect(zipcity.ZipsKnownFor("UT", "West Jordan"))
	if !slices.Contains(got, "84088") {
		t.Errorf("ZipsKnownFor(UT, West Jordan) = %v, want it to contain 84088", got)
	}

	// A city nothing was seen for yields nothing.
	if got := slices.Collect(zipcity.ZipsKnownFor("UT", "Nowhereville")); got != nil {
		t.Errorf("ZipsKnownFor(UT, Nowhereville) = %v, want nothing", got)
	}

	// West Palm Beach and Palm Beach are both known in Florida, and neither
	// ZIP Code list is empty, but they are different places: no ZIP Code
	// carries both names.
	westPalmBeach := slices.Collect(zipcity.ZipsKnownFor("FL", "West Palm Beach"))
	palmBeach := slices.Collect(zipcity.ZipsKnownFor("FL", "Palm Beach"))
	if len(westPalmBeach) == 0 || len(palmBeach) == 0 {
		t.Fatalf("ZipsKnownFor(FL, West Palm Beach) = %v, ZipsKnownFor(FL, Palm Beach) = %v, want both non-empty", westPalmBeach, palmBeach)
	}
	for _, zip := range palmBeach {
		if slices.Contains(westPalmBeach, zip) {
			t.Errorf("%q is in both ZipsKnownFor(FL, West Palm Beach) and ZipsKnownFor(FL, Palm Beach)", zip)
		}
	}
}

// ZipsKnownFor is CitiesKnownFor read the other way: every code it yields
// for a city and state must itself yield that city under that state.
func TestZipsKnownForRoundTripsWithCitiesKnownFor(t *testing.T) {
	for zip := range zipcity.ZipsKnownFor("UT", "West Jordan") {
		found := false
		for state, city := range zipcity.CitiesKnownFor(zip) {
			if state == "UT" && city == "WEST JORDAN" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ZipsKnownFor(UT, West Jordan) yielded %q, but CitiesKnownFor(%q) does not carry WEST JORDAN under UT", zip, zip)
		}
	}
}
