package usziplocale

import (
	"archive/zip"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The workbooks here are written by hand rather than checked in, so the tests
// say out loud what shape the reader expects: the header the Detail sheet
// prints on one line and the Other sheet stacks over three, the shared string
// table nearly every cell indexes into, and the inline strings a few use
// instead. Nothing here reaches the network; Download is not exercised.

func TestReadLocales(t *testing.T) {
	path := workbook(t, sheet{header: flatHeader, rows: detailRows, shared: true}, sheet{header: stackedHeader, rows: otherRows})

	var got []Locale
	err := ReadLocales(path, func(locale *Locale) error {
		got = append(got, *locale)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadLocales() = %v", err)
	}

	// Every data row of every sheet, in workbook order, whichever way the sheet
	// wrote its header or its text.
	want := []Locale{
		{"09001", "M", "APO", "P", "APO", "AE"},
		{"59732", "", "DILLON", "P", "DILLON", "MT"},
		{"20170", "", "HERNDON", "S", "HERNDON", "VA"},
		{"09002", "M", "FPO", "B", "FPO", "AE"},
		{"96950", "", "SAIPAN", "P", "SAIPAN", "MP"},
		{"96950", "", "CHALAN KANOA", "P", "SAIPAN", "MP"},
		{"09001", "M", "APO", "B", "APO", "AE"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadLocales() read %v, want %v", got, want)
	}
}

// A sheet whose header never names the columns we read is a different file
// than the one this package was written for, which is worth an error rather
// than a run of empty ZIP Codes.
func TestReadLocalesRejectsAWorkbookItCannotRead(t *testing.T) {
	path := workbook(t, sheet{header: [][]string{{"ZIP", "CITY"}}, rows: [][]string{{"20170", "HERNDON"}}, shared: true})

	err := ReadLocales(path, func(locale *Locale) error { return nil })
	if err == nil {
		t.Fatal("ReadLocales() = nil error, want one")
	}
	if !strings.Contains(err.Error(), colZip) {
		t.Errorf("ReadLocales() = %v, want it to name the column it could not find", err)
	}
}

func TestMilitaryZips(t *testing.T) {
	path := workbook(t, sheet{header: flatHeader, rows: detailRows, shared: true}, sheet{header: stackedHeader, rows: otherRows})

	military, err := MilitaryZips(path)
	if err != nil {
		t.Fatalf("MilitaryZips() = %v", err)
	}

	// Class M is the selector, not LOCALE TYPE: 09002 is a branch and belongs
	// here anyway. A second row for a code already held under the same name is
	// the same fact twice, not a conflict.
	if got := slices.Sorted(maps.Keys(military)); !slices.Equal(got, []string{"09001", "09002"}) {
		t.Errorf("MilitaryZips() keyed %v, want [09001 09002]", got)
	}
	if got := military["09002"].Name; got != "FPO" {
		t.Errorf("MilitaryZips()[09002].Name = %q, want FPO", got)
	}
	for zip, locale := range military {
		if !locale.Military() {
			t.Errorf("MilitaryZips() kept %s, which Military() calls false", zip)
		}
	}
}

// A class-M row named something other than APO, FPO or DPO is a row we cannot
// turn into a last line, and the two ways that can happen — no designation at
// all, or two designations for one code — are both reported rather than
// silently resolved.
func TestMilitaryZipsRejectsARowItCannotName(t *testing.T) {
	unnamed := workbook(t, sheet{header: flatHeader, shared: true, rows: [][]string{
		{"09001", "M", "JERSEY CITY", "P", "JERSEY CITY", "NJ"},
	}})
	if _, err := MilitaryZips(unnamed); err == nil {
		t.Error("MilitaryZips() on a class-M row named JERSEY CITY = nil error, want one")
	}

	conflicting := workbook(t, sheet{header: flatHeader, shared: true, rows: [][]string{
		{"09001", "M", "APO", "P", "APO", "AE"},
		{"09001", "M", "FPO", "B", "FPO", "AE"},
	}})
	if _, err := MilitaryZips(conflicting); err == nil {
		t.Error("MilitaryZips() on a code named both APO and FPO = nil error, want one")
	}
}

func TestPostOfficesByZip(t *testing.T) {
	path := workbook(t, sheet{header: flatHeader, rows: detailRows, shared: true}, sheet{header: stackedHeader, rows: otherRows})

	offices, err := PostOfficesByZip(path)
	if err != nil {
		t.Fatalf("PostOfficesByZip() = %v", err)
	}

	// Only LOCALE TYPE P: 20170's station and 09002's branch are not offices.
	// 96950 has two offices and keeps both, in the order the workbook lists
	// them.
	if got := slices.Sorted(maps.Keys(offices)); !slices.Equal(got, []string{"09001", "59732", "96950"}) {
		t.Errorf("PostOfficesByZip() keyed %v, want [09001 59732 96950]", got)
	}
	names := []string{}
	for _, office := range offices["96950"] {
		names = append(names, office.Name)
	}
	if !slices.Equal(names, []string{"SAIPAN", "CHALAN KANOA"}) {
		t.Errorf("PostOfficesByZip()[96950] named %v, want [SAIPAN CHALAN KANOA]", names)
	}
	// The office that delivers 59732 sits in Dillon, not in Glen, which is why
	// the generator counts these rather than trusting LOCALE NAME as the city.
	if got := offices["59732"][0].PhysicalCity; got != "DILLON" {
		t.Errorf("PostOfficesByZip()[59732].PhysicalCity = %q, want DILLON", got)
	}
}

func TestLocalName(t *testing.T) {
	for fileurl, want := range map[string]string{
		"https://postalpro.usps.com/mnt/glusterfs/2026-09/ZIP_Locale_Detail.xlsx": "2026-09_ZIP_Locale_Detail.xlsx",
		"https://postalpro.usps.com/ZIP_Locale_Detail.xlsx":                       "ZIP_Locale_Detail.xlsx",
	} {
		if got := localName(fileurl); got != want {
			t.Errorf("localName(%q) = %q, want %q", fileurl, got, want)
		}
	}
}

func TestColumn(t *testing.T) {
	for reference, want := range map[string]string{"B3": "B", "AA1207": "AA", "A": "A"} {
		if got := column(reference); got != want {
			t.Errorf("column(%q) = %q, want %q", reference, got, want)
		}
	}
}

// flatHeader is the Detail sheet's one-line heading, stackedHeader the Other
// sheet's, split over three rows with ZIP CLASS CODE the last to finish.
var (
	flatHeader = [][]string{
		{"DELIVERY ZIPCODE", "ZIP CLASS CODE", "LOCALE NAME", "LOCALE TYPE", "PHYSICAL CITY", "PHYSICAL STATE"},
	}
	stackedHeader = [][]string{
		{"DELIVERY", "ZIP", "LOCALE", "LOCALE", "PHYSICAL", "PHYSICAL"},
		{"ZIPCODE", "CLASS", "NAME", "TYPE", "CITY", "STATE"},
		{"", "CODE", "", "", "", ""},
	}

	detailRows = [][]string{
		{"09001", "M", "APO", "P", "APO", "AE"},
		{"59732", "", "DILLON", "P", "DILLON", "MT"},
		{"20170", "", "HERNDON", "S", "HERNDON", "VA"},
		{"09002", "M", "FPO", "B", "FPO", "AE"},
	}
	otherRows = [][]string{
		{"96950", "", "SAIPAN", "P", "SAIPAN", "MP"},
		{"96950", "", "CHALAN KANOA", "P", "SAIPAN", "MP"},
		{"09001", "M", "APO", "B", "APO", "AE"},
	}
)

// sheet is one worksheet to write: its header rows, its data rows, and whether
// its cells go through the shared string table or carry their text inline. The
// real workbook uses shared strings throughout; inline is here because the
// reader accepts both and one sheet exercising it is cheaper than trusting it.
type sheet struct {
	header [][]string
	rows   [][]string
	shared bool
}

// workbook writes a minimal .xlsx — the parts readSheetPaths, readSharedStrings
// and readSheet actually open — and returns its path.
func workbook(t *testing.T, sheets ...sheet) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ZIP_Locale_Detail.xlsx")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	shared := []string{}
	parts := map[string]string{}
	rels, list := []string{}, []string{}
	for i, s := range sheets {
		id := fmt.Sprintf("rId%d", i+1)
		target := fmt.Sprintf("worksheets/sheet%d.xml", i+1)
		rels = append(rels, fmt.Sprintf(`<Relationship Id=%q Target=%q/>`, id, target))
		list = append(list, fmt.Sprintf(`<sheet name="Sheet%d" sheetId="%d" r:id=%q/>`, i+1, i+1, id))
		parts["xl/"+target] = worksheet(s, &shared)
	}

	parts["xl/_rels/workbook.xml.rels"] = `<?xml version="1.0"?><Relationships>` + strings.Join(rels, "") + `</Relationships>`
	parts["xl/workbook.xml"] = `<?xml version="1.0"?><workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>` +
		strings.Join(list, "") + `</sheets></workbook>`
	parts["xl/sharedStrings.xml"] = sharedStrings(shared)

	for name, content := range parts {
		out, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := out.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	return path
}

// worksheet writes one sheet's XML, interning its text in shared when the
// sheet asked for shared strings. An empty cell is left out of the row
// entirely, the way a real sheet leaves one out.
func worksheet(s sheet, shared *[]string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><worksheet><sheetData>`)
	number := 0
	for _, rows := range [][][]string{s.header, s.rows} {
		for _, cells := range rows {
			number++
			b.WriteString(`<row>`)
			for i, value := range cells {
				if value == "" {
					continue
				}
				reference := fmt.Sprintf("%c%d", 'A'+i, number)
				if s.shared {
					index := slices.Index(*shared, value)
					if index < 0 {
						*shared = append(*shared, value)
						index = len(*shared) - 1
					}
					fmt.Fprintf(&b, `<c r=%q t="s"><v>%d</v></c>`, reference, index)
					continue
				}
				fmt.Fprintf(&b, `<c r=%q t="inlineStr"><is><t>%s</t></is></c>`, reference, value)
			}
			b.WriteString(`</row>`)
		}
	}
	b.WriteString(`</sheetData></worksheet>`)

	return b.String()
}

func sharedStrings(shared []string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><sst>`)
	for _, value := range shared {
		fmt.Fprintf(&b, `<si><t>%s</t></si>`, value)
	}
	b.WriteString(`</sst>`)

	return b.String()
}
