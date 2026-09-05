package ustigerline

import (
	"maps"
	"reflect"
	"slices"
	"testing"
)

// TestAbsentSourcesAgainstTheLiveIndex reads the Census Bureau's own FTP
// indexes — six requests, no archives — and asserts that what it publishes
// today is still what the generated AbsentSources record was built from.
//
// The point is that a change at the Census Bureau shows up as a failure here
// rather than as quietly different filters. It reaches the network, so it is
// skipped under -short; CI does not depend on census.gov being up.
func TestAbsentSourcesAgainstTheLiveIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("reads the Census Bureau's FTP indexes")
	}

	// The 2025 release publishes every county file for every county
	// equivalent except ADDR, which it omits for the five county equivalents
	// of American Samoa and for Rota (69100), the Northern Islands (69085)
	// and Tinian (69120). Saipan (69110) does publish one, describing 5 of
	// that island's 6,458 sides. See poetic-systems/zipcity#7.
	want := AbsentSources{
		"60010": {"addr"},
		"60020": {"addr"},
		"60030": {"addr"},
		"60040": {"addr"},
		"60050": {"addr"},
		"69085": {"addr"},
		"69100": {"addr"},
		"69120": {"addr"},
	}

	idx, err := readTigerfileIndexes(allRequiredTigerfiles())
	if err != nil {
		t.Fatalf("reading the FTP indexes: %v", err)
	}

	got := absentSources(idx.areasByType)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the published files have changed.\n got %v\nwant %v", got, want)
	}

	// A guard on the reading rather than on the Census Bureau: if the index
	// parse ever came back empty, every area would look absent from
	// everything and the comparison above would still be doing its job on
	// nonsense.
	for filetype, areas := range idx.areasByType {
		if len(areas) == 0 {
			t.Errorf("%s: the index listed no files at all", filetype)
		}
	}
	if counties := idx.counts["county"]; len(counties) < 3000 {
		t.Errorf("the index listed %d county equivalents, want the ~3,200 the release has", len(counties))
	}
	t.Logf("file types read: %v", slices.Sorted(maps.Keys(idx.areasByType)))
}
