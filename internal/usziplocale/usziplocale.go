// Package usziplocale reads the Postal Service's ZIP Locale Detail table.
//
// Every row is a postal facility — a post office, a station, a branch, a
// contract unit — and the ZIP Code it delivers, so the table is a facility
// directory rather than a ZIP Code to city table. That is why reading the
// LOCALE NAME as "the city" for a ZIP Code is wrong about one time in seven:
// the name is the office that delivers the mail, which is often the next town
// over (59732 is delivered from DILLON and GeoNames calls it GLEN). It is
// still a name the Postal Service writes on that ZIP Code's last line, which
// is why we key it, and it is recorded with its own source so a caller can
// tell it apart from the names GeoNames and TIGER offered. See
// poetic-systems/zipcity#48.
//
// Where the table is the better source outright is military mail. It names
// 571 ZIP Codes of class M, each with its own APO, FPO or DPO, where GeoNames
// has 302 fewer, never writes DPO at all, and calls 83 of them APO. What it
// does not give is the state: the row carries the gateway that accepts the
// mail (09xxx through Jersey City, 340xx Miami, 96xxx San Francisco), not the
// AA/AE/AP the address is written with. That still comes from the service
// area — usgeonames.MilitaryState — as poetic-systems/zipcity#24 established.
//
// The table is published as an .xlsx workbook, which is a zip archive of XML,
// so it is read here with archive/zip and encoding/xml rather than by taking
// on a spreadsheet library for one file a month.
package usziplocale

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// landingURL is the PostalPro page that publishes the table. The workbook
// itself lives under a path naming the month it was cut
// (/mnt/glusterfs/2026-09/ZIP_Locale_Detail.xlsx), so the link is read off
// the page rather than written here — the same way ustigerline takes the
// TIGER file list off the Census Bureau's own index instead of keeping a copy
// of it.
const landingURL = "https://postalpro.usps.com/ZIP_Locale_Detail"

// workbookLink finds the workbook's href on the landing page. The page is
// Drupal-rendered HTML and this is the only .xlsx on it.
var workbookLink = regexp.MustCompile(`(?i)href="([^"]*ZIP_Locale_Detail\.xlsx)"`)

var storagedir = filepath.Join(strings.Split("./data/usps_zip_locale_detail/", "/")...)

// Locale is one row of the table, reduced to the columns we use.
//
// Zip is the ZIP Code the facility delivers, Class the Postal Service's class
// for that code — M military, P PO box only, U unique, blank for an ordinary
// one. Name is the facility's own name and Type what kind of facility it is:
// P a post office, S a station, B a branch, 4 and 5 contract units. Physical
// City and State are where the facility itself sits, which for military mail
// is the gateway rather than anywhere the addressee is.
type Locale struct {
	Zip           string
	Class         string
	Name          string
	Type          string
	PhysicalCity  string
	PhysicalState string
}

// Military reports whether the row is a military ZIP Code: class M, named
// APO, FPO or DPO. Both are checked because the designation is the name we
// key, and a class-M row we cannot name is one we would rather report than
// guess at.
func (l *Locale) Military() bool {
	return l.Class == "M" && slices.Contains(designations, l.Name)
}

// designations are the three overseas military post office designations, the
// city line of a military address.
var designations = []string{"APO", "FPO", "DPO"}

// LocaleFunc receives each row in workbook order. Returning an error stops
// the read and is returned to the caller.
type LocaleFunc func(locale *Locale) error

// Download caches the workbook locally and returns its path.
//
// The local name keeps the month the workbook was published under, taken from
// the URL, so a cache holding several months is unambiguous and the file a
// generation read can be named. A workbook already on disk is never fetched
// again; a short or corrupt download removes the file rather than leaving
// something that fails later and further away, which is how usgeonames and
// ustigerline both behave.
func Download() (string, error) {
	fileurl, err := workbookURL()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(storagedir, 0755)
	if err != nil {
		return "", fmt.Errorf("Error creating ZIP Locale Detail storage directory %s: %w", storagedir, err)
	}

	localpath := filepath.Join(storagedir, localName(fileurl))
	out, err := os.OpenFile(localpath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return localpath, nil
		}
		return "", err
	}
	defer out.Close()

	log.Printf("Downloading %s to %s", fileurl, localpath)
	resp, err := get(fileurl)
	if err != nil {
		os.Remove(localpath)
		return "", err
	}
	defer resp.Body.Close()

	expectedSize := resp.ContentLength
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(localpath)
		return "", err
	}
	if expectedSize != -1 && bytesWritten != expectedSize {
		os.Remove(localpath)
		return "", fmt.Errorf("Incomplete download! Got %d of %d bytes", bytesWritten, expectedSize)
	}
	out.Close() // this will get called twice!

	reader, err := zip.OpenReader(localpath)
	if err != nil {
		os.Remove(localpath)
		return "", fmt.Errorf("Corrupt or incomplete workbook: %w", err)
	}
	defer reader.Close()

	log.Printf("Downloaded all %d bytes of %s", bytesWritten, localpath)

	return localpath, nil
}

// localName is the workbook's published path reduced to its month and file
// name: 2026-09_ZIP_Locale_Detail.xlsx.
func localName(fileurl string) string {
	parsed, err := url.Parse(fileurl)
	if err != nil {
		return "ZIP_Locale_Detail.xlsx"
	}

	dir, file := path.Split(parsed.Path)
	month := path.Base(path.Clean(dir))
	if month == "." || month == "/" || month == "" {
		return file
	}

	return month + "_" + file
}

// workbookURL reads the landing page and resolves the workbook's link against
// it. A page that no longer carries the link is an error rather than a guess
// at where the file moved to.
func workbookURL() (string, error) {
	resp, err := get(landingURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	page, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading %s: %w", landingURL, err)
	}

	href := workbookLink.FindSubmatch(page)
	if href == nil {
		return "", fmt.Errorf("%s no longer links a ZIP_Locale_Detail.xlsx", landingURL)
	}

	link, err := url.Parse(string(href[1]))
	if err != nil {
		return "", fmt.Errorf("%s links an unparseable workbook %q: %w", landingURL, href[1], err)
	}
	base, err := url.Parse(landingURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(link).String(), nil
}

// get fetches a URL and fails on any status but 200.
//
// PostalPro is served through a CDN that refuses Go's default user agent, so
// one is set. Nothing else about the request matters.
func get(fileurl string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fileurl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "zipcity/1 (+https://github.com/poetic-systems/zipcity)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s returned status %d", fileurl, resp.StatusCode)
	}

	return resp, nil
}

// ReadLocales calls localeFn for every data row on every sheet of the
// workbook at archivepath, in workbook order.
//
// The table is published over three sheets — Detail, Unique and Other — which
// hold the same columns and differ only in which ZIP Codes they carry, so they
// are read as one sequence. A caller wanting only military rows reads them all
// and keeps the ones Military reports; which sheet a row came from is not a
// fact about the ZIP Code.
func ReadLocales(archivepath string, localeFn LocaleFunc) error {
	reader, err := zip.OpenReader(archivepath)
	if err != nil {
		return fmt.Errorf("Error opening ZIP Locale Detail workbook %s: %w", archivepath, err)
	}
	defer reader.Close()

	shared, err := readSharedStrings(&reader.Reader)
	if err != nil {
		return fmt.Errorf("Error reading %s: %w", archivepath, err)
	}
	sheets, err := readSheetPaths(&reader.Reader)
	if err != nil {
		return fmt.Errorf("Error reading %s: %w", archivepath, err)
	}

	for _, sheet := range sheets {
		err = readSheet(&reader.Reader, sheet, shared, localeFn)
		if err != nil {
			return fmt.Errorf("Error reading %s in %s: %w", sheet, archivepath, err)
		}
	}

	return nil
}

// MilitaryZips reads the military ZIP Codes out of the workbook, keyed by ZIP
// Code.
//
// Each of these codes has exactly one row, because a military ZIP Code is
// served by exactly one overseas post office, and the row's name is the APO,
// FPO or DPO the address is written with. A second row for a code we already
// have is an error rather than a silent choice between them.
func MilitaryZips(archivepath string) (map[string]*Locale, error) {
	military := map[string]*Locale{}
	err := ReadLocales(archivepath, func(locale *Locale) error {
		if locale.Class != "M" {
			return nil
		}
		if !locale.Military() {
			return fmt.Errorf("ZIP Code %s is class M but named %q, not APO, FPO or DPO", locale.Zip, locale.Name)
		}
		if held, found := military[locale.Zip]; found && held.Name != locale.Name {
			return fmt.Errorf("ZIP Code %s is named both %s and %s", locale.Zip, held.Name, locale.Name)
		}
		military[locale.Zip] = locale

		return nil
	})
	if err != nil {
		return nil, err
	}

	return military, nil
}

// PostOfficesByZip reads the post office rows — LOCALE TYPE P — out of the
// workbook, keyed by the ZIP Code they deliver.
//
// A ZIP Code may have several: about 3,000 do, where more than one office
// delivers into it. All of them are kept, because the Postal Service accepts
// the last line each of them writes and nothing here can say which is
// preferred — the same reason PlacesByPostalCode keeps every GeoNames name.
// About 8,800 ZIP Codes have no post office row at all, only stations and
// branches, and those yield nothing.
func PostOfficesByZip(archivepath string) (map[string][]*Locale, error) {
	offices := map[string][]*Locale{}
	err := ReadLocales(archivepath, func(locale *Locale) error {
		if locale.Type != "P" || locale.Zip == "" || locale.Name == "" {
			return nil
		}
		offices[locale.Zip] = append(offices[locale.Zip], locale)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return offices, nil
}

// The columns we read, by the header the workbook prints over them. Every one
// of these has to be found or we are not reading the file we think we are, the
// same check usgeonames makes on its column count.
const (
	colZip           = "DELIVERY ZIPCODE"
	colClass         = "ZIP CLASS CODE"
	colName          = "LOCALE NAME"
	colType          = "LOCALE TYPE"
	colPhysicalCity  = "PHYSICAL CITY"
	colPhysicalState = "PHYSICAL STATE"
)

var required = []string{colZip, colClass, colName, colType, colPhysicalCity, colPhysicalState}

// readSheet streams one sheet, reading its header first and then its rows.
//
// The header is not always one row: the Detail sheet prints DELIVERY ZIPCODE
// on a single line while the Other sheet splits it over three, DELIVERY above
// ZIPCODE. So header text is accumulated down each column until every column
// we need has been named, which reconstructs the one-line headings from the
// stacked ones and leaves the two sheets read by the same code.
func readSheet(reader *zip.Reader, sheetpath string, shared []string, localeFn LocaleFunc) error {
	file, err := reader.Open(sheetpath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	header := map[string]string{}
	columns := map[string]string{}
	for {
		row, err := nextRow(decoder)
		if err != nil {
			return err
		}
		if row == nil {
			if len(columns) == 0 {
				return fmt.Errorf("no header naming %s", strings.Join(required, ", "))
			}
			return nil
		}

		cells := row.values(shared)
		if len(columns) == 0 {
			for column, value := range cells {
				header[column] = strings.TrimSpace(header[column] + " " + value)
			}
			columns = headerColumns(header)
			continue
		}

		err = localeFn(&Locale{
			Zip:           cells[columns[colZip]],
			Class:         cells[columns[colClass]],
			Name:          cells[columns[colName]],
			Type:          cells[columns[colType]],
			PhysicalCity:  cells[columns[colPhysicalCity]],
			PhysicalState: cells[columns[colPhysicalState]],
		})
		if err != nil {
			return err
		}
	}
}

// headerColumns inverts the accumulated header text into the column each of
// the headings we need sits in, or nil until every one of them is there.
func headerColumns(header map[string]string) map[string]string {
	columns := map[string]string{}
	for column, heading := range header {
		if slices.Contains(required, heading) {
			columns[heading] = column
		}
	}
	if len(columns) < len(required) {
		return nil
	}

	return columns
}

// row is one <row> of a sheet, and cell one <c> within it.
type row struct {
	Cells []cell `xml:"c"`
}

type cell struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	Inline    text   `xml:"is"`
}

// text is a run of spreadsheet text, which is either a single <t> or a series
// of <r> runs each holding one.
type text struct {
	Plain []string `xml:"t"`
	Runs  []struct {
		Plain string `xml:"t"`
	} `xml:"r"`
}

func (t text) String() string {
	parts := append([]string{}, t.Plain...)
	for _, run := range t.Runs {
		parts = append(parts, run.Plain)
	}

	return strings.TrimSpace(strings.Join(parts, ""))
}

// values reads a row into its cell text by column letter, resolving shared
// strings. A cell the sheet leaves out is simply absent, so a caller indexing
// a column reads "" for it.
func (r *row) values(shared []string) map[string]string {
	values := make(map[string]string, len(r.Cells))
	for _, c := range r.Cells {
		values[column(c.Reference)] = c.text(shared)
	}

	return values
}

// text resolves one cell. Most of this workbook is shared strings, whose <v>
// is an index into the table rather than the text itself; an inline string
// carries its own, and anything else — a number, a date — is taken as
// written.
func (c *cell) text(shared []string) string {
	switch c.Type {
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(c.Value))
		if err != nil || index < 0 || index >= len(shared) {
			return ""
		}
		return shared[index]
	case "inlineStr":
		return c.Inline.String()
	}

	return strings.TrimSpace(c.Value)
}

// column is the column letters of a cell reference: B for B3, AA for AA1207.
func column(reference string) string {
	for i, r := range reference {
		if r >= '0' && r <= '9' {
			return reference[:i]
		}
	}

	return reference
}

// nextRow decodes the next <row> of a sheet, or nil at the end of it.
func nextRow(decoder *xml.Decoder) (*row, error) {
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "row" {
			continue
		}

		next := &row{}
		if err := decoder.DecodeElement(next, &start); err != nil {
			return nil, err
		}

		return next, nil
	}
}

// readSharedStrings reads the workbook's shared string table, which is where
// nearly every cell's text actually lives.
func readSharedStrings(reader *zip.Reader) ([]string, error) {
	file, err := reader.Open("xl/sharedStrings.xml")
	if err != nil {
		return nil, fmt.Errorf("no shared string table: %w", err)
	}
	defer file.Close()

	var table struct {
		Items []text `xml:"si"`
	}
	if err := xml.NewDecoder(file).Decode(&table); err != nil {
		return nil, err
	}

	shared := make([]string, len(table.Items))
	for i, item := range table.Items {
		shared[i] = item.String()
	}

	return shared, nil
}

// readSheetPaths reads the workbook's sheets in the order it lists them,
// following each sheet's relationship id to the part that holds it.
func readSheetPaths(reader *zip.Reader) ([]string, error) {
	targets, err := readRelationships(reader)
	if err != nil {
		return nil, err
	}

	file, err := reader.Open("xl/workbook.xml")
	if err != nil {
		return nil, fmt.Errorf("no workbook part: %w", err)
	}
	defer file.Close()

	var workbook struct {
		Sheets []struct {
			ID string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
		} `xml:"sheets>sheet"`
	}
	if err := xml.NewDecoder(file).Decode(&workbook); err != nil {
		return nil, err
	}

	paths := []string{}
	for _, sheet := range workbook.Sheets {
		target, found := targets[sheet.ID]
		if !found {
			return nil, fmt.Errorf("sheet %s names no part", sheet.ID)
		}
		paths = append(paths, path.Join("xl", target))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("workbook holds no sheets")
	}

	return paths, nil
}

// readRelationships reads the workbook's relationship ids to the parts they
// name, which is how a sheet's part is found.
func readRelationships(reader *zip.Reader) (map[string]string, error) {
	file, err := reader.Open("xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, fmt.Errorf("no workbook relationships: %w", err)
	}
	defer file.Close()

	var rels struct {
		Relationships []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.NewDecoder(file).Decode(&rels); err != nil {
		return nil, err
	}

	targets := map[string]string{}
	for _, rel := range rels.Relationships {
		targets[rel.ID] = rel.Target
	}

	return targets, nil
}
