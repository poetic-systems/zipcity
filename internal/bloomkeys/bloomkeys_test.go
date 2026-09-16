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
		got := bloomkeys.KeyZipStreet(tc.Zip, tc.Street)
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
		// Punctuation dropped (Pub 28 §222), the hyphen as a space
		{"27101", "Winston-Salem", "27101:WINSTON SALEM"},
		{"64063", "LEE'S SUMMIT", "64063:LEES SUMMIT"},
		// Abbreviated words spelled out (Pub 28 §223), wherever a word follows
		{"05478", "St. Albans", "05478:SAINT ALBANS"},
		{"39520", "BAY ST. LOUIS", "39520:BAY SAINT LOUIS"},
		{"49783", "Sault Ste. Marie", "49783:SAULT SAINTE MARIE"},
		{"10550", "MT VERNON", "10550:MOUNT VERNON"},
		{"76102", "Ft Worth", "76102:FORT WORTH"},
		// Already spelled out: unchanged, so both forms build one key
		{"05478", "SAINT ALBANS", "05478:SAINT ALBANS"},
		// A trailing ST is left alone
		{"00000", "HIGH ST", "00000:HIGH ST"},
	}

	for _, tc := range cases {
		got := bloomkeys.KeyZipCity(tc.Zip, tc.City)
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
		// The city is spelled out; the street is not touched
		{"St. Paul", "MN", "ST PAUL ST", "SAINT PAUL:MN:ST PAUL ST"},
	}

	for _, tc := range cases {
		got := bloomkeys.KeyCityStateStreet(tc.City, tc.State, tc.Street)
		if got != tc.Want {
			t.Fatalf("Wanted: '%s' Got: '%s", tc.Want, got)
		}
	}
}
