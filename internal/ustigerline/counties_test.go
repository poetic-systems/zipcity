package ustigerline_test

import (
	"slices"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

// The fixture county, asked for three times over, comes out in the order
// asked for and says the same thing however many are read at once.
func TestReadCountiesYieldsInOrderWhateverTheWorkerCount(t *testing.T) {
	defer ustigerline.UseFixtures()()

	prefixes := []string{ustigerline.FixturePrefix, ustigerline.FixturePrefix, ustigerline.FixturePrefix}
	sides, err := ustigerline.ReadStreetSides(ustigerline.FixturePrefix)
	if err != nil {
		t.Fatal(err)
	}
	for _, workers := range []int{1, 2, 8} {
		got := []string{}
		for county := range ustigerline.ReadCounties(prefixes, workers) {
			if county.Err != nil {
				t.Fatal(county.Err)
			}
			if len(county.Sides) != len(sides) {
				t.Errorf("workers=%d: %d sides, want %d", workers, len(county.Sides), len(sides))
			}
			got = append(got, county.Prefix)
		}
		if !slices.Equal(got, prefixes) {
			t.Errorf("workers=%d yielded %v, want %v", workers, got, prefixes)
		}
	}
}

// A county that cannot be read is reported in its turn rather than stopping
// the ones after it.
func TestReadCountiesReportsAnUnreadableCountyInItsTurn(t *testing.T) {
	defer ustigerline.UseFixtures()()

	prefixes := []string{"tl_2025_00000", ustigerline.FixturePrefix}
	counties := slices.Collect(ustigerline.ReadCounties(prefixes, 2))
	if len(counties) != 2 || counties[0].Err == nil || counties[1].Err != nil {
		t.Fatalf("got %+v", counties)
	}
}
