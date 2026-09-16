package zipcities

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	table := make(Table)
	table.Add("20170", "VA", "Herndon")
	table.Add("20170", "va", "HERNDON")
	table.Add("20170", "VA", "reston")
	table.Add("00926", "PR", "SAN JUAN")
	table.Add("00917", "PR", "CAÑO MARTIN PEÑA")
	table.Add("00917", "PR", "Cano Martin Pena")
	table.Add("05478", "VT", "St. Albans")
	table.Add("05478", "VT", "SAINT ALBANS")

	// Case, diacritics and Pub 28 spelling are the filter's, and a name
	// recorded twice is one name.
	want := Table{
		"20170": {"VA": {"HERNDON", "RESTON"}},
		"00926": {"PR": {"SAN JUAN"}},
		"00917": {"PR": {"CANO MARTIN PENA"}},
		"05478": {"VT": {"SAINT ALBANS"}},
	}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}
}

// A ZIP Code that straddles a state line, or that GeoNames files under no
// state, keeps each name under the state its source gave it.
func TestAddKeepsStatesApart(t *testing.T) {
	table := make(Table)
	table.Add("42223", "KY", "FORT CAMPBELL")
	table.Add("42223", "TN", "FORT CAMPBELL")
	table.Add("96860", "HI", "JBPHH")
	table.Add("96860", "", "FPO AA")

	want := Table{
		"42223": {"KY": {"FORT CAMPBELL"}, "TN": {"FORT CAMPBELL"}},
		"96860": {"HI": {"JBPHH"}, "": {"FPO AA"}},
	}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}
}

// GeoNames leaves a place name empty often enough that an empty string would
// otherwise become a city we claim to have seen.
func TestAddIgnoresEmpty(t *testing.T) {
	table := make(Table)
	table.Add("20170", "VA", "")
	table.Add("", "VA", "HERNDON")
	if len(table) != 0 {
		t.Errorf("Add() recorded %v, want nothing", table)
	}
}

func TestEncode(t *testing.T) {
	table := Table{
		"20170": {"VA": {"RESTON", "HERNDON"}},
		"00926": {"PR": {"SAN JUAN"}},
		"42223": {"TN": {"FORT CAMPBELL"}, "KY": {"FORT CAMPBELL"}},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}

	want := "00926\tPR\tSAN JUAN\n20170\tVA\tHERNDON\tRESTON\n42223\tKY\tFORT CAMPBELL\n42223\tTN\tFORT CAMPBELL\n"
	if got := buf.String(); got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

// The file is committed, so a generation that changed nothing must not show up
// as a diff.
func TestEncodeIsStable(t *testing.T) {
	table := make(Table)
	for _, zip := range []string{"20170", "00926", "99827", "01001"} {
		for _, state := range []string{"VA", "PR", "AK"} {
			table.Add(zip, state, "FIRST")
			table.Add(zip, state, "SECOND")
			table.Add(zip, state, "THIRD")
		}
	}

	var first bytes.Buffer
	if err := Encode(&first, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	for range 20 {
		var again bytes.Buffer
		if err := Encode(&again, table); err != nil {
			t.Fatalf("Encode() = %v", err)
		}
		if again.String() != first.String() {
			t.Fatalf("Encode() wrote %q on a later call, %q on the first", again.String(), first.String())
		}
	}
}

func TestDecode(t *testing.T) {
	table := Table{
		"20170": {"VA": {"HERNDON", "RESTON"}},
		"00926": {"PR": {"SAN JUAN"}},
		"42223": {"KY": {"FORT CAMPBELL"}, "TN": {"FORT CAMPBELL"}},
		"96860": {"HI": {"JBPHH"}, "": {"FPO AA"}},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	got, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode() = %v", err)
	}
	if !reflect.DeepEqual(got, table) {
		t.Errorf("Decode(Encode(%v)) = %v", table, got)
	}
}

func TestDecodeRejectsALineWithNoName(t *testing.T) {
	for _, in := range []string{"20170\n", "20170\tVA\n", "\tVA\tHERNDON\n", "20170\tVA\tHERNDON\n99999\tAK\n"} {
		if _, err := Decode(strings.NewReader(in)); err == nil {
			t.Errorf("Decode(%q) = nil error, want one", in)
		}
	}
}
