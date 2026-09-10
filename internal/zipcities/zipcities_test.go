package zipcities

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	table := make(Table)
	table.Add("20170", "Herndon")
	table.Add("20170", "HERNDON")
	table.Add("20170", "reston")
	table.Add("00926", "SAN JUAN")

	// Case is the filter's case, and a name recorded twice is one name.
	want := Table{"20170": {"HERNDON", "RESTON"}, "00926": {"SAN JUAN"}}
	if !reflect.DeepEqual(table, want) {
		t.Errorf("Add() built %v, want %v", table, want)
	}
}

// GeoNames leaves a place name empty often enough that an empty string would
// otherwise become a city we claim to have seen.
func TestAddIgnoresEmpty(t *testing.T) {
	table := make(Table)
	table.Add("20170", "")
	table.Add("", "HERNDON")
	if len(table) != 0 {
		t.Errorf("Add() recorded %v, want nothing", table)
	}
}

func TestEncode(t *testing.T) {
	table := Table{"20170": {"RESTON", "HERNDON"}, "00926": {"SAN JUAN"}}

	var buf bytes.Buffer
	if err := Encode(&buf, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}

	want := "00926\tSAN JUAN\n20170\tHERNDON\tRESTON\n"
	if got := buf.String(); got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

// The file is committed, so a generation that changed nothing must not show up
// as a diff.
func TestEncodeIsStable(t *testing.T) {
	table := make(Table)
	for _, zip := range []string{"20170", "00926", "99827", "01001"} {
		table.Add(zip, "FIRST")
		table.Add(zip, "SECOND")
		table.Add(zip, "THIRD")
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
	table := Table{"20170": {"HERNDON", "RESTON"}, "00926": {"SAN JUAN"}}

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
	for _, in := range []string{"20170\n", "\tHERNDON\n", "20170\tHERNDON\n99999\n"} {
		if _, err := Decode(strings.NewReader(in)); err == nil {
			t.Errorf("Decode(%q) = nil error, want one", in)
		}
	}
}

func TestMeasure(t *testing.T) {
	table := Table{"20170": {"HERNDON", "RESTON"}, "00926": {"SAN JUAN"}}

	size, err := Measure(table)
	if err != nil {
		t.Fatalf("Measure() = %v", err)
	}
	if size.Zips != 2 || size.Names != 3 {
		t.Errorf("Measure() counted %d ZIP Codes and %d names, want 2 and 3", size.Zips, size.Names)
	}

	var buf bytes.Buffer
	if err := Encode(&buf, table); err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if size.Bytes != buf.Len() {
		t.Errorf("Measure() reported %d bytes, Encode wrote %d", size.Bytes, buf.Len())
	}
	if size.Gzipped == 0 {
		t.Error("Measure() reported 0 gzipped bytes")
	}
}
