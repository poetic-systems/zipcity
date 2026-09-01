package ustigerline_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("unable to find the working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("unable to find the module root above the working directory")
		}
		dir = parent
	}
}

// cachedAddrPrefix returns the file prefix of any locally cached ADDR file.
// The TIGER data set is too large to check in, so this test works against
// whichever counties the developer happens to have downloaded rather than
// naming one that may not be present.
func cachedAddrPrefix(t *testing.T) string {
	t.Helper()

	// The reader resolves the TIGER cache relative to the working directory,
	// which for a test is the package directory rather than the module root.
	t.Chdir(moduleRoot(t))

	entries, err := os.ReadDir(filepath.Join("data", "us_census_tiger", "addr"))
	if err != nil {
		t.Skipf("no cached TIGER ADDR files to read: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, "_addr.zip") {
			return strings.TrimSuffix(name, "_addr.zip")
		}
	}
	t.Skip("no cached TIGER ADDR files to read")
	return ""
}

// Many address ranges share one TLID: both sides of an edge are described
// under it, and each side is described by a range per run of house numbers.
// Recording one ZIP Code per TLID let the second side overwrite the first, and
// recording one per side let the second range overwrite the first. See #5.
func TestEverySideKeepsEveryZipOfItsAddressRanges(t *testing.T) {
	fileprefix := cachedAddrPrefix(t)

	zips := map[string]map[string][]string{}
	err := ustigerline.ReadAddressRanges(
		fileprefix,
		func(info *ustigerline.AddressRange) error {
			if info.Side != "L" && info.Side != "R" {
				t.Errorf("address range for TLID with side %q; want L or R", info.Side)
				return nil
			}
			if info.Zip == "" {
				t.Errorf("address range reported with no ZIP Code")
				return nil
			}
			sides, ok := zips[info.TLID]
			if !ok {
				sides = map[string][]string{}
				zips[info.TLID] = sides
			}
			if slices.Contains(sides[info.Side], info.Zip) {
				t.Errorf("a ZIP Code was reported more than once for side %s of an edge", info.Side)
			}
			sides[info.Side] = append(sides[info.Side], info.Zip)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadAddressRanges(): %v", err)
	}

	twoSided, multiZip := 0, 0
	for _, sides := range zips {
		if len(sides["L"]) > 0 && len(sides["R"]) > 0 {
			twoSided++
		}
		for _, sideZips := range sides {
			if len(sideZips) > 1 {
				multiZip++
			}
		}
	}
	if twoSided == 0 {
		t.Errorf(
			"no edge in %s reported a ZIP Code on both sides; %d edges were read",
			fileprefix, len(zips),
		)
	}
	// A side is described by an address range per run of house numbers, and
	// the ranges at each end of a street may name different ZIP Codes. Keeping
	// one ZIP Code per side threw the rest away.
	if multiZip == 0 {
		t.Errorf(
			"no side of any edge in %s reported more than one ZIP Code; %d edges were read",
			fileprefix, len(zips),
		)
	}
	t.Logf(
		"%d of %d edges reported a ZIP Code on both sides; %d sides reported more than one ZIP Code",
		twoSided, len(zips), multiZip,
	)
}
