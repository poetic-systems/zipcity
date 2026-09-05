package ustigerline_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

func TestVerifyTigerfileZipAcceptsAPublishedArchive(t *testing.T) {
	path := filepath.Join("testdata", "us_census_tiger", "addr", ustigerline.FixturePrefix+"_addr.zip")

	err := ustigerline.VerifyTigerfileZip(path)
	if err != nil {
		t.Fatalf("Error from ustigerline.VerifyTigerfileZip(%s): %v", path, err)
	}
}

func TestVerifyTigerfileZipRejectsWhatTheCensusBureauServesInstead(t *testing.T) {
	cases := map[string][]byte{
		// A 200 answer carrying the Census Bureau's rejection page.
		"error page": []byte("The requested URL was rejected. Please consult with your administrator."),
		// A download cut short.
		"truncated archive": firstBytesOf(t, filepath.Join("testdata", "us_census_tiger", "addr", ustigerline.FixturePrefix+"_addr.zip")),
		// A well-formed archive with nothing in it.
		"empty archive": emptyZip(t),
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tl_2025_01001_addr.zip")
			err := os.WriteFile(path, body, 0666)
			if err != nil {
				t.Fatalf("Error writing %s: %v", path, err)
			}

			err = ustigerline.VerifyTigerfileZip(path)
			if err == nil {
				t.Fatalf("ustigerline.VerifyTigerfileZip(%s) accepted a %s", path, name)
			}
		})
	}
}

func TestVerifyTigerfileZipRejectsAnAbsentArchive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tl_2025_69085_addr.zip")

	err := ustigerline.VerifyTigerfileZip(path)
	if err == nil {
		t.Fatalf("ustigerline.VerifyTigerfileZip(%s) accepted a file that is not there", path)
	}
}

func firstBytesOf(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading %s: %v", path, err)
	}
	return contents[:len(contents)/2]
}

func emptyZip(t *testing.T) []byte {
	t.Helper()

	archive := &bytes.Buffer{}
	err := zip.NewWriter(archive).Close()
	if err != nil {
		t.Fatalf("Error writing an empty archive: %v", err)
	}
	return archive.Bytes()
}
