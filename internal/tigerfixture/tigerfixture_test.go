package tigerfixture_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/poetic-systems/zipcity/internal/tigerfixture"
	"github.com/twpayne/go-shapefile"
)

// The place fixture is the smallest checked-in archive that carries all five
// members a subset has to deal with: .cpg, .dbf, .prj, .shp and .shx. It is
// the fixture this tool wrote, which is the point — a subset of a subset has
// to read back the same way.
var source = filepath.Join("..", "ustigerline", "testdata", "us_census_tiger", "place", "tl_2025_01_place.zip")

// A kept record has to come back saying exactly what it said in the file it
// came from. The lengths in the .shp and .shx headers are checked by the
// reader against the actual file size, so a read that succeeds at all is
// evidence they were patched correctly.
func TestASubsetKeepsWhatItSaysVerbatim(t *testing.T) {
	full, err := shapefile.ReadZipFile(source, nil)
	if err != nil {
		t.Fatalf("unable to read %s: %v", source, err)
	}
	if full.NumRecords() < 3 {
		t.Fatalf("%s holds %d records, too few to subset meaningfully", source, full.NumRecords())
	}

	// Every other record, so that the kept ones are not contiguous and the
	// record numbering cannot come out right by copying the source's.
	var want []int
	for i := range full.NumRecords() {
		if i%2 == 1 {
			want = append(want, i)
		}
	}

	at := 0
	dst := filepath.Join(t.TempDir(), "subset.zip")
	kept, err := tigerfixture.Subset(source, dst, func(map[string]any) bool {
		keep := at%2 == 1
		at++
		return keep
	})
	if err != nil {
		t.Fatalf("unable to subset %s: %v", source, err)
	}
	if kept != len(want) {
		t.Fatalf("kept %d records; want %d", kept, len(want))
	}

	subset, err := shapefile.ReadZipFile(dst, nil)
	if err != nil {
		t.Fatalf("unable to read the subset back: %v", err)
	}
	if subset.NumRecords() != len(want) {
		t.Fatalf("the subset holds %d records; want %d", subset.NumRecords(), len(want))
	}

	for n, i := range want {
		wantFields, wantGeom := full.Record(i)
		gotFields, gotGeom := subset.Record(n)
		if !reflect.DeepEqual(gotFields, wantFields) {
			t.Errorf("record %d of the subset says %v; want source record %d's %v", n, gotFields, i, wantFields)
		}
		if !reflect.DeepEqual(gotGeom, wantGeom) {
			t.Errorf("record %d of the subset has a different shape than source record %d", n, i)
		}
	}
}

// Dropping every record is the degenerate case, and it is the one most likely
// to leave a header describing records that are not there.
func TestAnEmptySubsetIsStillReadable(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "empty.zip")
	kept, err := tigerfixture.Subset(source, dst, func(map[string]any) bool { return false })
	if err != nil {
		t.Fatalf("unable to subset %s: %v", source, err)
	}
	if kept != 0 {
		t.Fatalf("kept %d records; want 0", kept)
	}

	subset, err := shapefile.ReadZipFile(dst, nil)
	if err != nil {
		t.Fatalf("unable to read the empty subset back: %v", err)
	}
	if subset.NumRecords() != 0 {
		t.Fatalf("the empty subset holds %d records; want 0", subset.NumRecords())
	}
}
