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
