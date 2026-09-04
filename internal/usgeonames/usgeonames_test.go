package usgeonames

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Every row here is invented. The package reads a published format, not
// published values, so the fixture states the shapes the reader has to get
// right and names nothing real.
const usRows = "" +
	"US\t10001\tFairhaven\tExample State\tXA\tExample County\t001\t\t\t40.0\t-70.0\t4\n" +
	"US\t10001\tNorth Fairhaven\tExample State\tXA\tExample County\t001\t\t\t40.0\t-70.0\t4\n" +
	"US\t10002\tAPO AE\t\t\t\t\t\t\t40.0\t-70.0\t4\n" +
	"US\t10003\tAPO STA\t\t\t\t\t\t\t40.0\t-70.0\t4\n" +
	"CA\tX0X0X0\tSomewhere Else\tExample Province\tXP\t\t\t\t\t40.0\t-70.0\t4\n"

const prRows = "" +
	"PR\t00999\tVilla Ejemplo\tVilla Ejemplo\t042\t\t\t\t\t18.0\t-66.0\t4\n"

// American Samoa names no subdivision at all, which is what makes the whole
// territory the only area its ZIP Code can be attributed to.
const asRows = "" +
	"AS\t96999\tExample Village\tAs\t\t\t\t\t\t-14.0\t-170.0\t4\n"

// writeArchive builds a GeoNames-shaped archive: <code>.zip holding
// <code>.txt, which is how the reader finds the rows.
func writeArchive(t *testing.T, dir, code, rows string) string {
	t.Helper()

	path := filepath.Join(dir, code+".zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating %s: %s", path, err)
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	member, err := archive.Create(code + ".txt")
	if err != nil {
		t.Fatalf("creating %s.txt: %s", code, err)
	}
	_, err = member.Write([]byte(rows))
	if err != nil {
		t.Fatalf("writing %s.txt: %s", code, err)
	}
	err = archive.Close()
	if err != nil {
		t.Fatalf("closing %s: %s", path, err)
	}

	return path
}

func TestTheStateComesFromWhicheverColumnHoldsIt(t *testing.T) {
	dir := t.TempDir()
	paths := []string{
		writeArchive(t, dir, "US", usRows),
		writeArchive(t, dir, "PR", prRows),
	}

	places, err := PlacesByPostalCode(paths)
	if err != nil {
		t.Fatalf("PlacesByPostalCode: %s", err)
	}

	cases := []struct {
		name  string
		zip   string
		place string
		state string
	}{
		{"a state file takes its state from admin code1", "10001", "Fairhaven", "XA"},
		{"a territory takes its state from the country", "00999", "Villa Ejemplo", "PR"},
		{"a military post office splits city from state", "10002", "APO", "AE"},
		{"an unrecognized trailing word is left alone", "10003", "APO STA", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			found := places[c.zip]
			if len(found) == 0 {
				t.Fatalf("no place for %s", c.zip)
			}
			if found[0].PlaceName != c.place || found[0].StateUSPS != c.state {
				t.Errorf("read %q/%q, want %q/%q",
					found[0].PlaceName, found[0].StateUSPS, c.place, c.state)
			}
		})
	}
}

func TestAZipCodeKeepsEveryPlaceNamedForIt(t *testing.T) {
	dir := t.TempDir()

	places, err := PlacesByPostalCode([]string{writeArchive(t, dir, "US", usRows)})
	if err != nil {
		t.Fatalf("PlacesByPostalCode: %s", err)
	}

	found := places["10001"]
	if len(found) != 2 {
		t.Fatalf("kept %d places for a ZIP Code the file names twice, want 2", len(found))
	}
	if found[0].PlaceName != "Fairhaven" || found[1].PlaceName != "North Fairhaven" {
		t.Errorf("kept %q and %q", found[0].PlaceName, found[1].PlaceName)
	}
}

func TestARowFromAnotherPostalSystemIsSkipped(t *testing.T) {
	dir := t.TempDir()

	places, err := PlacesByPostalCode([]string{writeArchive(t, dir, "US", usRows)})
	if err != nil {
		t.Fatalf("PlacesByPostalCode: %s", err)
	}

	if _, found := places["X0X0X0"]; found {
		t.Error("a Canadian row was read from a file that should hold none")
	}
}

func TestAWrongColumnCountIsAnError(t *testing.T) {
	dir := t.TempDir()
	path := writeArchive(t, dir, "US", "US\t10001\tFairhaven\n")

	err := ReadUSPostalPlaces(path, func(*PostalPlace) error { return nil })
	if err == nil {
		t.Fatal("read a three column row as though it were the documented twelve")
	}
}

func TestTheCountyComesFromWhicheverColumnHoldsIt(t *testing.T) {
	dir := t.TempDir()
	paths := []string{
		writeArchive(t, dir, "US", usRows),
		writeArchive(t, dir, "PR", prRows),
		writeArchive(t, dir, "AS", asRows),
	}

	places, err := PlacesByPostalCode(paths)
	if err != nil {
		t.Fatalf("PlacesByPostalCode: %s", err)
	}

	// US.zip holds the state in admin code1 and the county in admin code2;
	// PR.zip has only the one level and holds the municipio in admin code1;
	// AS.zip names neither, and a military row names nothing at all.
	for _, want := range []struct {
		postalcode string
		countycode string
	}{
		{"10001", "001"},
		{"10002", ""},
		{"00999", "042"},
		{"96999", ""},
	} {
		found := places[want.postalcode]
		if len(found) == 0 {
			t.Fatalf("no place for %s", want.postalcode)
		}
		if found[0].CountyCode != want.countycode {
			t.Errorf("county code for %s: got %q, want %q", want.postalcode, found[0].CountyCode, want.countycode)
		}
	}
}

func TestAnAreaCollectsEveryZipCodeNamedWithinIt(t *testing.T) {
	dir := t.TempDir()
	paths := []string{
		writeArchive(t, dir, "US", usRows),
		writeArchive(t, dir, "PR", prRows),
		writeArchive(t, dir, "AS", asRows),
	}

	bystate, bycounty, err := PostalCodesByArea(paths)
	if err != nil {
		t.Fatalf("PostalCodesByArea: %s", err)
	}

	// A ZIP Code named twice, once per place name, is one ZIP Code.
	for area, want := range map[string][]string{"XA": {"10001"}, "AE": {"10002"}, "PR": {"00999"}, "AS": {"96999"}} {
		if !slices.Equal(bystate[area], want) {
			t.Errorf("ZIP Codes for state %s: got %v, want %v", area, bystate[area], want)
		}
	}

	for area, want := range map[string][]string{"XA001": {"10001"}, "PR042": {"00999"}} {
		if !slices.Equal(bycounty[area], want) {
			t.Errorf("ZIP Codes for county %s: got %v, want %v", area, bycounty[area], want)
		}
	}

	// American Samoa's ZIP Code belongs to the territory and to no county
	// within it, which is the whole of what we know about where it applies.
	if len(bycounty["AS"]) > 0 {
		t.Errorf("ZIP Codes for a territory with no counties: got %v, want none", bycounty["AS"])
	}
}
