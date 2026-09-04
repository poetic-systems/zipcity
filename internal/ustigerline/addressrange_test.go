package ustigerline_test

import (
	"slices"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

// fixturePrefix points the readers at the checked-in county slice and names
// it. The full TIGER data set is too large to check in, so the fixtures are a
// connected slice of one county rather than the county itself; see
// internal/tigerfixture.
func fixturePrefix(t *testing.T) string {
	t.Helper()
	t.Cleanup(ustigerline.UseFixtures())
	return ustigerline.FixturePrefix
}

// ReadAddressRanges reports what the file says and decides nothing. TIGER
// records an address range per run of house numbers on each side of an edge, so
// many ranges repeat a side and its ZIP Code exactly, and a few name no ZIP Code
// at all. Collecting them into a map here instead of reporting them threw those
// away and left the caller unable to tell how a side was described. See #5.
func TestAddressRangesAreReportedOneForOne(t *testing.T) {
	fileprefix := fixturePrefix(t)

	records, repeated, noZip := 0, 0, 0
	seen := map[ustigerline.AddressRange]bool{}
	err := ustigerline.ReadAddressRanges(
		fileprefix,
		func(info *ustigerline.AddressRange) error {
			records++
			if info.Zip == "" {
				noZip++
			}
			if seen[*info] {
				repeated++
			}
			seen[*info] = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadAddressRanges(): %v", err)
	}

	if repeated == 0 {
		t.Errorf(
			"no address range in %s repeated one already reported, so ranges are being collapsed rather than reported; %d records were read",
			fileprefix, records,
		)
	}
	// Ranges naming no ZIP Code are rare enough that a given county may hold
	// none, so their count is reported rather than required. Whether they are
	// worth anything is ReadStreetSides' decision, not this reader's.
	t.Logf("%d records read, %d of them repeating one already reported and %d naming no ZIP Code", records, repeated, noZip)
}

// A side of an edge has no single ZIP Code: the ranges at each end of a street
// may name different ones. ReadStreetSides keeps every distinct ZIP Code its
// ranges name, and only ReadStreetSides decides that.
func TestEverySideKeepsEveryZipOfItsAddressRanges(t *testing.T) {
	fileprefix := fixturePrefix(t)

	sides, err := ustigerline.ReadStreetSides(fileprefix)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadStreetSides(): %v", err)
	}

	edges := map[string]int{}
	multiZip := 0
	for _, side := range sides {
		if side.Side != "L" && side.Side != "R" {
			t.Errorf("street side reported with side %q; want L or R", side.Side)
		}
		if len(side.Zips) == 0 {
			continue
		}
		edges[side.TLID]++
		if len(side.Zips) > 1 {
			multiZip++
		}
		for i, zip := range side.Zips {
			if zip == "" {
				t.Errorf("a side kept an empty ZIP Code")
			}
			if slices.Contains(side.Zips[:i], zip) {
				t.Errorf("a side kept the same ZIP Code twice")
			}
		}
	}

	twoSided := 0
	for _, withZips := range edges {
		if withZips == 2 {
			twoSided++
		}
	}
	if twoSided == 0 {
		t.Errorf("no edge in %s kept a ZIP Code on both sides; %d sides were read", fileprefix, len(sides))
	}
	if multiZip == 0 {
		t.Errorf("no side in %s kept more than one ZIP Code; %d sides were read", fileprefix, len(sides))
	}
	t.Logf(
		"%d of %d sides read; %d edges kept a ZIP Code on both sides, %d sides kept more than one",
		len(edges), len(sides), twoSided, multiZip,
	)
}
