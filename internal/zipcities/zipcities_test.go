package zipcities

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	table := make(Table)
	table.Add("20170", "VA", "Herndon", Seen)
	table.Add("20170", "va", "HERNDON", Seen)
	table.Add("20170", "VA", "reston", Seen)
	table.Add("00926", "PR", "SAN JUAN", Seen)
	table.Add("00917", "PR", "CAÑO MARTIN PEÑA", Seen)
	table.Add("00917", "PR", "Cano Martin Pena", Seen)
	table.Add("05478", "VT", "St. Albans", Seen)
	table.Add("05478", "VT", "SAINT ALBANS", Seen)

	// Case, diacritics and Pub 28 spelling are the filter's, and a name
	// recorded twice by one source is one name.
	want := Table{
		"20170": {"VA": {{"HERNDON", Seen}, {"RESTON", Seen}}},
		"00926": {"PR": {{"SAN JUAN", Seen}}},
		"00917": {"PR": {{"CANO MARTIN PENA", Seen}}},
		"05478": {"VT": {{"SAINT ALBANS", Seen}}},
	}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}
}

// The same name from two sources is two records, because which sources offered
// a name is the thing Source exists to answer — but one name to a caller
// asking only what the names are.
func TestAddKeepsSourcesApart(t *testing.T) {
	table := make(Table)
	table.Add("59732", "MT", "Glen", Seen)
	table.Add("59732", "MT", "Dillon", USPS)
	table.Add("20170", "VA", "Herndon", Seen)
	table.Add("20170", "VA", "Herndon", USPS)

	want := Table{
		"59732": {"MT": {{"GLEN", Seen}, {"DILLON", USPS}}},
		"20170": {"VA": {{"HERNDON", Seen}, {"HERNDON", USPS}}},
	}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}

	if got := table.Names("20170", "VA"); !slices.Equal(got, []string{"HERNDON"}) {
		t.Errorf(`Names("20170", "VA") = %v, want [HERNDON]`, got)
	}
	if got := table.Names("59732", "MT"); !slices.Equal(got, []string{"DILLON", "GLEN"}) {
		t.Errorf(`Names("59732", "MT") = %v, want [DILLON GLEN]`, got)
	}
	if got := table.NamesFrom("59732", "MT", USPS); !slices.Equal(got, []string{"DILLON"}) {
		t.Errorf(`NamesFrom("59732", "MT", USPS) = %v, want [DILLON]`, got)
	}
	if got := table.NamesFrom("59732", "MT", Seen); !slices.Equal(got, []string{"GLEN"}) {
		t.Errorf(`NamesFrom("59732", "MT", Seen) = %v, want [GLEN]`, got)
	}
}

// A ZIP Code that straddles a state line, or that GeoNames files under no
// state, keeps each name under the state its source gave it.
func TestAddKeepsStatesApart(t *testing.T) {
	table := make(Table)
	table.Add("42223", "KY", "FORT CAMPBELL", Seen)
	table.Add("42223", "TN", "FORT CAMPBELL", Seen)
	table.Add("96860", "HI", "JBPHH", Seen)
	table.Add("96860", "", "FPO AA", Seen)

	want := Table{
		"42223": {"KY": {{"FORT CAMPBELL", Seen}}, "TN": {{"FORT CAMPBELL", Seen}}},
		"96860": {"HI": {{"JBPHH", Seen}}, "": {{"FPO AA", Seen}}},
	}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}
}

// GeoNames leaves a place name empty often enough that an empty string would
// otherwise become a city we claim to have seen.
func TestAddIgnoresEmpty(t *testing.T) {
	table := make(Table)
	table.Add("20170", "VA", "", Seen)
	table.Add("", "VA", "HERNDON", Seen)
	if len(table) != 0 {
		t.Errorf("Add() recorded %v, want nothing", table)
	}
}

func TestEncode(t *testing.T) {
	table := Table{
		"20170": {"VA": {{"RESTON", Seen}, {"HERNDON", Seen}, {"HERNDON", USPS}}},
		"00926": {"PR": {{"SAN JUAN", Seen}}},
		"42223": {"TN": {{"CLARKSVILLE", USPS}}, "KY": {{"FORT CAMPBELL", Seen}}},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}

	want := strings.Join([]string{
		"00926\tPR\tseen\tSAN JUAN",
		"20170\tVA\tseen\tHERNDON\tRESTON",
		"20170\tVA\tusps\tHERNDON",
		"42223\tKY\tseen\tFORT CAMPBELL",
		"42223\tTN\tusps\tCLARKSVILLE",
		"",
	}, "\n")
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
			table.Add(zip, state, "FIRST", Seen)
			table.Add(zip, state, "SECOND", Seen)
			table.Add(zip, state, "THIRD", USPS)
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
		"20170": {"VA": {{"HERNDON", Seen}, {"RESTON", Seen}, {"HERNDON", USPS}}},
		"00926": {"PR": {{"SAN JUAN", Seen}}},
		"42223": {"KY": {{"FORT CAMPBELL", Seen}}, "TN": {{"FORT CAMPBELL", Seen}}},
		"96860": {"HI": {{"JBPHH", Seen}}, "": {{"FPO AA", Seen}}},
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

func TestDecodeRejectsALineItCannotRead(t *testing.T) {
	for _, in := range []string{
		"20170\n",
		"20170\tVA\n",
		"20170\tVA\tseen\n",
		"\tVA\tseen\tHERNDON\n",
		"20170\tVA\tseen\tHERNDON\n99999\tAK\tseen\n",
		"20170\tVA\tHERNDON\tRESTON\n",
		"20170\tVA\tguessed\tHERNDON\n",
	} {
		if _, err := Decode(strings.NewReader(in)); err == nil {
			t.Errorf("Decode(%q) = nil error, want one", in)
		}
	}
}
