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
	{
		Street: "W 9200 S",
		City:   "West Jordan",
		State:  "UT",
		Zip:    "84088",
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

// Puerto Rico street names put the type at the front, and TIGER spells that
// type its own way — CLL for CALLE. Project US@ page 26 forbids abbreviating a
// street name, so a conforming caller sends the form on the left and the data
// holds the form on the right. Before the leading type was aliased, every one
// of these returned a definitive false for a street that is in the data.
//
// These are street names only, with no premise number attached. None of them
// identifies a residence.
func TestSpecFormPuertoRicoStreetNamesAreFound(t *testing.T) {
	for _, tc := range []struct {
		city, state, street string
	}{
		{"SAN JUAN", "PR", "CALLE LOIZA"},
		{"SAN JUAN", "PR", "CALLE ASHFORD"},
		{"SAN JUAN", "PR", "CALLE DE LA FORTALEZA"},
		{"SAN JUAN", "PR", "CALLE SAN FRANCISCO"},
		{"SAN JUAN", "PR", "AVENIDA ASHFORD"},
	} {
		found, err := zipcity.CheckCityStateAndStreet(tc.city, tc.state, tc.street)
		if err != nil {
			t.Fatalf("Error checking city: '%s' state: '%s' and street: '%s': %s", tc.city, tc.state, tc.street, err)
		}
		if !found {
			t.Errorf("Expected city: '%s' state: '%s' and street: '%s' to be found", tc.city, tc.state, tc.street)
		}
	}
}

// The spelling TIGER actually holds has to keep working. Aliasing adds
// spellings; it must not replace the one that was already right.
func TestTigerFormPuertoRicoStreetNamesStillWork(t *testing.T) {
	for _, street := range []string{"CLL LOIZA", "CLL ASHFORD", "CLL SAN FRANCISCO", "AVE ASHFORD"} {
		found, err := zipcity.CheckCityStateAndStreet("SAN JUAN", "PR", street)
		if err != nil {
			t.Fatalf("Error checking street '%s': %s", street, err)
		}
		if !found {
			t.Errorf("Expected street '%s' to still be found", street)
		}
	}
}

// An invented street name must not be found. Aliasing multiplies the number of
// filter probes, so this is the guard that the extra probes have not turned the
// filter into something that answers true for everything.
func TestAliasingDoesNotFindInventedStreets(t *testing.T) {
	for _, street := range []string{
		"CALLE ZZQXVW NONEXISTENT",
		"CLL ZZQXVW NONEXISTENT",
		"AVENIDA ZZQXVW NONEXISTENT",
	} {
		found, err := zipcity.CheckCityStateAndStreet("SAN JUAN", "PR", street)
		if err != nil {
			t.Fatalf("Error checking street '%s': %s", street, err)
		}
		if found {
			t.Errorf("Expected invented street '%s' not to be found", street)
		}
	}
}
