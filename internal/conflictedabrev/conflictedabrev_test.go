package conflictedabrev_test

import (
	"slices"
	"testing"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/conflictedabrev"
)

func filterOf(names ...string) *bloom.BloomFilter {
	f := bloom.NewWithEstimates(1000, 0.00001)
	for _, n := range names {
		for k := range conflictedabrev.Keys(n) {
			f.AddString(k)
		}
	}
	return f
}

func TestKeys(t *testing.T) {
	cases := []struct {
		Name string
		Want []string
	}{
		{"MAIN AVE", nil},
		{"MOUNT ZION", []string{"conflicted-next-word:MOUNT:ZION"}},
		// The last word has no next word, the first no previous one.
		{"ROCKY MT", []string{"conflicted-previous-word:ROCKY:MT"}},
		{"BAY ST LOUIS", []string{
			"conflicted-previous-word:BAY:ST",
			"conflicted-next-word:ST:LOUIS",
		}},
		// A lone word yields nothing, and extra spaces are not words.
		{"ST", nil},
		{"MT  ZION", []string{"conflicted-next-word:MT:ZION"}},
	}
	for _, tc := range cases {
		got := slices.Collect(conflictedabrev.Keys(tc.Name))
		if !slices.Equal(got, tc.Want) {
			t.Errorf("Keys(%q) = %v, want %v", tc.Name, got, tc.Want)
		}
	}
}

func TestReadings(t *testing.T) {
	f := filterOf("MOUNT ZION", "ROCKY MOUNT", "MT TREASURE", "BAY ST LOUIS", "MAIN ST", "CT HOUSE")
	cases := []struct {
		Word, Prev, Next string
		Want             []string
	}{
		// MT beside ZION is MOUNT: that is the spelling the names hold.
		{"MT", "", "ZION", []string{"MOUNT"}},
		{"MOUNT", "", "ZION", []string{"MOUNT"}},
		{"MT", "", "TREASURE", []string{"MT"}},
		{"mount", "ROCKY", "", []string{"MOUNT"}},
		{"ST", "BAY", "LOUIS", []string{"ST"}},
		{"ST", "MAIN", "", []string{"ST"}},
		{"CT", "", "HOUSE", []string{"CT"}},
		// Asking with a spelling reads the same conflict as asking with its abbreviation.
		{"COURT", "", "HOUSE", []string{"CT"}},
		// Not conflicted, or never seen beside these neighbors.
		{"AVE", "MAIN", "", nil},
		{"ST", "ELM", "", nil},
		{"ST", "", "", nil},
	}
	for _, tc := range cases {
		got := conflictedabrev.Readings(f, tc.Word, tc.Prev, tc.Next)
		if !slices.Equal(got, tc.Want) {
			t.Errorf("Readings(%q, %q, %q) = %v, want %v", tc.Word, tc.Prev, tc.Next, got, tc.Want)
		}
	}
}

func TestReadingsMoreThanOne(t *testing.T) {
	f := filterOf("MOUNT ZION", "MT ZION")
	got := conflictedabrev.Readings(f, "MT", "", "ZION")
	want := []string{"MT", "MOUNT"}
	if !slices.Equal(got, want) {
		t.Errorf("Readings = %v, want %v", got, want)
	}
}
