package bloomkeys_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/bloomkeys"
)

func TestKeyZipStreet(t *testing.T) {
	cases := []struct {
		Zip    string
		Street string
		Want   string
	}{
		{"94523", "Pleasant Hill Rd", "94523:PLEASANT HILL RD"},
		{"00662", "CALLEJÓN AMADOR", "00662:CALLEJON AMADOR"},
		// Zip: 00662 City: ISABELA Street: CALLEJÓN AMADOR
		{"00662", "CALLEJÓN MONSO MÉNDEZ", "00662:CALLEJON MONSO MENDEZ"},
		// Zip: 00917 City: CAÑO MARTIN PEÑA Street: CALLE E
		{"00917", "CALLE E", "00917:CALLE E"},
	}

	for _, tc := range cases {
		got, err := bloomkeys.KeyZipStreet(tc.Zip, tc.Street)
		if err != nil {
			t.Fatalf("Error preparing zip-street key: %s", err)
		}
		if got != tc.Want {
			t.Fatalf("Wanted: '%s' Got: '%s", tc.Want, got)
		}
	}
}

func TestKeyZipCity(t *testing.T) {
	cases := []struct {
		Zip  string
		City string
		Want string
	}{
		{"94523", "Pleasant Hill", "94523:PLEASANT HILL"},
		// Zip: 00662 City: ISABELA Street: CALLEJÓN AMADOR
		{"00662", "ISABELA", "00662:ISABELA"},
		// Zip: 00917 City: CAÑO MARTIN PEÑA Street: CALLE E
		{"00917", "CAÑO MARTIN PEÑA", "00917:CANO MARTIN PENA"},
	}

	for _, tc := range cases {
		got, err := bloomkeys.KeyZipCity(tc.Zip, tc.City)
		if err != nil {
			t.Fatalf("Error preparing zip-street key: %s", err)
		}
		if got != tc.Want {
			t.Fatalf("Wanted: '%s' Got: '%s", tc.Want, got)
		}
	}
}

func TestKeyCityStateStreet(t *testing.T) {
	cases := []struct {
		City   string
		State  string
		Street string
		Want   string
	}{
		{"Pleasant Hill", "CA", "Pleasant Hill Rd", "PLEASANT HILL:CA:PLEASANT HILL RD"},
		// Zip: 00662 City: ISABELA Street: CALLEJÓN AMADOR
		{"ISABELA", "PR", "CALLEJÓN AMADOR", "ISABELA:PR:CALLEJON AMADOR"},
		// Zip: 00917 City: CAÑO MARTIN PEÑA Street: CALLE E
		{"CAÑO MARTIN PEÑA", "PR", "CALLE E", "CANO MARTIN PENA:PR:CALLE E"},
	}

	for _, tc := range cases {
		got, err := bloomkeys.KeyCityStateStreet(tc.City, tc.State, tc.Street)
		if err != nil {
			t.Fatalf("Error preparing zip-street key: %s", err)
		}
		if got != tc.Want {
			t.Fatalf("Wanted: '%s' Got: '%s", tc.Want, got)
		}
	}
}
