package ustigerline

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"slices"
	"strings"
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
	counties := idx.countyPrefixes()
	if len(counties) < 3000 {
		t.Errorf("the index listed %d county equivalents, want the ~3,200 the release has", len(counties))
	}
	// The prefix is read off the file names rather than written down, so a
	// release that renamed its files would otherwise send the readers looking
	// for archives under a name nothing is stored as.
	if want := "tl_2025_01001"; counties[0] != want {
		t.Errorf("first county prefix = %q, want %q", counties[0], want)
	}
	t.Logf("file types read: %v", slices.Sorted(maps.Keys(idx.areasByType)))
}

// TestReadTigerfileIndexesRejectsAnIndexThatListsNothing is the failure
// poetic-systems/zipcity#26 reported: four of the six indexes came back with
// no file names, absentSources had nothing to compare and reported no
// absences, and the generator went on to panic on a release with no counties
// in it.
//
// Both bodies are answers the Census Bureau really serves under HTTP 200 —
// its rejection page, and a directory listing with the archives missing.
func TestReadTigerfileIndexesRejectsAnIndexThatListsNothing(t *testing.T) {
	bodies := map[string]string{
		"rejection page": "The requested URL was rejected. Please consult with your administrator.",
		"listing with no archives": `<html><body><table>
			<tr><td><a href="?C=N;O=D">Name</a></td></tr>
			<tr><td><a href="/geo/tiger/TIGER2025/">Parent Directory</a></td></tr>
			</table></body></html>`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/FEATNAMES") {
					fmt.Fprint(w, body)
					return
				}
				fmt.Fprint(w, `<a href="tl_2025_01001`+strings.ToLower(path.Base(r.URL.Path))+`.zip">`)
			}))
			defer server.Close()

			source, err := url.Parse(server.URL + "/")
			if err != nil {
				t.Fatalf("parsing the test server URL: %v", err)
			}
			required := []RequiredTigerfiles{
				{Source: source, Path: featureFTPPath, Suffix: "_featnames"},
				{Source: source, Path: edgeFTPPath, Suffix: "_edges"},
			}

			_, err = readTigerfileIndexes(required)
			if err == nil {
				t.Fatalf("readTigerfileIndexes accepted an index serving a %s", name)
			}
		})
	}
}

// TestReadTigerfileIndexesRejectsAnErrorStatus keeps a refused request from
// reading as an empty directory: the message should name what the Census
// Bureau answered rather than what we made of the body.
func TestReadTigerfileIndexesRejectsAnErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "go away", http.StatusForbidden)
	}))
	defer server.Close()

	source, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parsing the test server URL: %v", err)
	}

	_, err = readTigerfileIndexes([]RequiredTigerfiles{
		{Source: source, Path: featureFTPPath, Suffix: "_featnames"},
	})
	if err == nil {
		t.Fatal("readTigerfileIndexes accepted a 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %v, want it to name the status", err)
	}
}
