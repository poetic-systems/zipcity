package ustigerline_test

import (
	"os"
	"path/filepath"
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

// Both sides of an edge are separate address ranges sharing one TLID, and each
// side carries its own ZIP Code. Collecting them under the TLID alone let the
// second side read overwrite the first, so an edge that should have yielded a
// ZIP Code for each side yielded one for whichever side was read last. See #5.
func TestBothSidesOfAnEdgeKeepTheirOwnZip(t *testing.T) {
	fileprefix := cachedAddrPrefix(t)

	zips := map[string]map[string]string{}
	err := ustigerline.ReadAddressRanges(
		fileprefix,
		func(info *ustigerline.AddressRange) error {
			if info.Side != "L" && info.Side != "R" {
				t.Errorf("address range for TLID with side %q; want L or R", info.Side)
				return nil
			}
			sides, ok := zips[info.TLID]
			if !ok {
				sides = map[string]string{}
				zips[info.TLID] = sides
			}
			if _, repeated := sides[info.Side]; repeated {
				t.Errorf("side %s of an edge was reported more than once", info.Side)
			}
			sides[info.Side] = info.Zip
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadAddressRanges(): %v", err)
	}

	twoSided := 0
	for _, sides := range zips {
		if sides["L"] != "" && sides["R"] != "" {
			twoSided++
		}
	}
	if twoSided == 0 {
		t.Fatalf(
			"no edge in %s reported a ZIP Code on both sides; %d edges were read",
			fileprefix, len(zips),
		)
	}
	t.Logf("%d of %d edges reported a ZIP Code on both sides", twoSided, len(zips))
}
