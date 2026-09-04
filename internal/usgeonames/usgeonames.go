// Package usgeonames reads the GeoNames postal code export for the United
// States.
//
// TIGER gives a street side its city from the PLACEFP of the face it borders,
// which is blank wherever the address is not inside an incorporated place. The
// Postal Service does not work that way: mail to an unincorporated area is
// addressed to the city of the post office that delivers it, so those addresses
// have a city even though no place contains them. That city is not in TIGER at
// all, and this is where we get it. See poetic-systems/zipcity#5.
//
// The export is one line per postal code and place name pair, so a ZIP Code
// with more than one acceptable city name appears more than once. Nothing here
// picks among them: a bloom filter is asked whether a pair is plausible, and
// every pair the Postal Service accepts is plausible.
package usgeonames

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// baseURL is where the GeoNames postal code export lives. The format is
// documented at https://download.geonames.org/export/zip/readme.txt and the
// data is licensed CC BY 4.0.
const baseURL = "https://download.geonames.org/export/zip/"

// PostalCountries is every GeoNames country file that carries a US ZIP Code.
//
// GeoNames splits its export by ISO country code, and it treats each territory
// as a country of its own, so US.zip is the fifty states and the District of
// Columbia and nothing else. The Postal Service does not draw that line —
// mail to San Juan is addressed with a ZIP Code like mail to Denver — so
// leaving the territories out would silently make every one of their ZIP Codes
// unknown to us. See poetic-systems/zipcity#7.
//
// The ISO code of each territory is also its USPS state abbreviation, which is
// what stateUSPS relies on.
var PostalCountries = []string{"US", "PR", "VI", "GU", "AS", "MP", "MH", "FM", "PW"}

var storagedir = filepath.Join(strings.Split("./data/geonames/", "/")...)

// fields is the column count the readme states. A row with any other number of
// columns is not the file we think we are reading, and we would rather say so
// than index into it.
const fields = 12

// Column positions, from the readme:
//
//	country code, postal code, place name, admin name1, admin code1,
//	admin name2, admin code2, admin name3, admin code3,
//	latitude, longitude, accuracy
const (
	colCountry    = 0
	colPostalCode = 1
	colPlaceName  = 2
	colStateName  = 3
	colStateUSPS  = 4
	colCountyName = 5
	colCountyCode = 6
)

// PostalPlace is one row of the export, reduced to the columns we use.
//
// StateUSPS is the two letter code KeyCityStateStreet wants. Where it comes
// from depends on the file, which is why stateUSPS exists. CountyCode is the
// county equivalent's FIPS code, where the row names one at all, and is why
// countyFIPS exists.
type PostalPlace struct {
	Country    string
	PostalCode string
	PlaceName  string
	StateName  string
	StateUSPS  string
	CountyName string
	CountyCode string
}

// stateUSPS reports the USPS state abbreviation for a row.
//
// In US.zip admin code1 is the state code, exactly what we want. In a
// territory file it is not: PR.zip records "00601 Adjuntas" with an admin
// code1 of "001", the municipio's FIPS code, because within GeoNames' model
// Puerto Rico is the country and its subdivisions are municipios. For those
// files the country code is the abbreviation the Postal Service uses, so that
// is what we take.
func stateUSPS(country, admincode1 string) string {
	if country == "US" {
		return admincode1
	}

	return country
}

// countyFIPS reports the county equivalent's FIPS code for a row, or "" where
// the row names none.
//
// Which column holds it depends on the file, because GeoNames models every
// territory as a country and its subdivisions as that country's states. US.zip
// and the files for Guam and the Virgin Islands put the state in admin code1
// and the county in admin code2 ("AK" and "013", "66" and "010"). Puerto Rico
// and the Northern Mariana Islands have only one level and put the municipio
// or municipality there instead ("001", "110"). American Samoa and Palau name
// neither.
//
// A county equivalent's FIPS code is three digits and none of the state level
// codes are, so the shape tells them apart and we do not have to keep a table
// of which file is shaped which way.
func countyFIPS(admincode1, admincode2 string) string {
	for _, code := range []string{admincode2, admincode1} {
		if isCountyFIPS(code) {
			return code
		}
	}

	return ""
}

func isCountyFIPS(code string) bool {
	if len(code) != 3 {
		return false
	}

	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return false
		}
	}

	return true
}

// PostalPlaceFunc receives each row in file order. Returning an error stops
// the read and is returned to the caller.
type PostalPlaceFunc func(place *PostalPlace) error

// DownloadUSPostalCodes caches every archive in PostalCountries locally and
// returns the paths of the ones it has.
//
// A country whose archive cannot be fetched is reported in absent rather than
// failing the whole download, and its ZIP Codes are simply not known to the
// filters — the same distinction poetic-systems/zipcity#2 asks for on the
// TIGER side. An absent country is a gap a caller can name; a failed download
// is a build that produces nothing.
func DownloadUSPostalCodes() (paths []string, absent map[string]error, err error) {
	err = os.MkdirAll(storagedir, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating GeoNames storage directory %s: %w", storagedir, err)
	}

	absent = map[string]error{}
	for _, country := range PostalCountries {
		localpath, err := downloadPostalCodeZip(country)
		if err != nil {
			absent[country] = err
			continue
		}

		paths = append(paths, localpath)
	}

	if len(paths) == 0 {
		return nil, absent, fmt.Errorf("no GeoNames postal code archive could be downloaded")
	}

	return paths, absent, nil
}

// downloadPostalCodeZip fetches one country's archive if it is not already on
// disk and returns its path.
//
// This mirrors ustigerline's download: the local file is opened O_EXCL first,
// so an archive already on disk is never fetched again, and a short or corrupt
// download removes the file rather than leaving something that would fail
// later and further away.
func downloadPostalCodeZip(country string) (string, error) {
	localpath := filepath.Join(storagedir, country+".zip")
	out, err := os.OpenFile(localpath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return localpath, nil
		}
		return "", err
	}
	defer out.Close()

	fileurl := baseURL + country + ".zip"
	fmt.Printf("Downloading %s to %s\n", fileurl, localpath)
	resp, err := http.Get(fileurl)
	if err != nil {
		os.Remove(localpath)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(localpath)
		return "", fmt.Errorf("%s returned status %d", fileurl, resp.StatusCode)
	}

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
		return "", fmt.Errorf("Corrupt or incomplete zip: %w", err)
	}
	defer reader.Close()

	fmt.Printf("Downloaded all %d bytes of %s\n", bytesWritten, localpath)

	return localpath, nil
}

// militaryStates are the USPS state abbreviations for overseas military mail:
// Armed Forces Americas, Europe and Pacific.
var militaryStates = []string{"AA", "AE", "AP"}

// overseasMilitary splits a military post office row into its city and state.
//
// GeoNames leaves admin code1 blank for these and writes the pair into the
// place name instead, as "APO AE" or "FPO AA". On an address those are two
// components — APO is the city line and AE is the state — and a caller
// checking a city against a state would otherwise be handed "APO AE" as a
// city name that no address ever writes.
//
// The split only happens when the trailing word is one of the three military
// state codes. The file also holds a single "APO STA", which is not one of
// them and is left alone rather than guessed at.
func overseasMilitary(placename, state string) (string, string) {
	if len(state) > 0 {
		return placename, state
	}

	city, military, found := strings.Cut(placename, " ")
	if !found || !slices.Contains(militaryStates, military) {
		return placename, state
	}

	return city, military
}

// ReadUSPostalPlaces calls placeFn for every row in the archive at archivepath
// that belongs to the US postal system.
//
// Rows for any other country are skipped rather than reported. The per-country
// files should not contain any, but the same format is published for all
// countries at once as allCountries.zip, so a caller who points this at the
// wrong archive gets nothing rather than gets Canada.
func ReadUSPostalPlaces(archivepath string, placeFn PostalPlaceFunc) error {
	reader, err := zip.OpenReader(archivepath)
	if err != nil {
		return fmt.Errorf("Error opening GeoNames archive %s: %w", archivepath, err)
	}
	defer reader.Close()

	dataFile := strings.TrimSuffix(filepath.Base(archivepath), ".zip") + ".txt"
	file, err := reader.Open(dataFile)
	if err != nil {
		return fmt.Errorf("Error opening %s within %s: %w", dataFile, archivepath, err)
	}
	defer file.Close()

	rows := csv.NewReader(file)
	rows.Comma = '\t'
	// Place names carry quotation marks that are part of the name rather than
	// csv quoting, and the export has no quoting of its own.
	rows.LazyQuotes = true
	rows.FieldsPerRecord = fields

	for line := 1; ; line++ {
		row, err := rows.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error reading %s line %d: %w", dataFile, line, err)
		}

		country := row[colCountry]
		if !slices.Contains(PostalCountries, country) {
			continue
		}

		place, state := overseasMilitary(row[colPlaceName], stateUSPS(country, row[colStateUSPS]))

		err = placeFn(&PostalPlace{
			Country:    country,
			PostalCode: row[colPostalCode],
			PlaceName:  place,
			StateName:  row[colStateName],
			StateUSPS:  state,
			CountyName: row[colCountyName],
			CountyCode: countyFIPS(row[colStateUSPS], row[colCountyCode]),
		})
		if err != nil {
			return err
		}
	}
}

// PlacesByPostalCode reads the archives into a lookup from ZIP Code to every
// place the Postal Service accepts for it.
//
// A ZIP Code with one city yields a single entry; one serving several
// communities yields all of them. Callers wanting "the" city for a ZIP Code
// will not find it here, because the export does not say which is preferred
// and we would be inventing the answer.
func PlacesByPostalCode(archivepaths []string) (map[string][]*PostalPlace, error) {
	places := map[string][]*PostalPlace{}

	for _, archivepath := range archivepaths {
		err := ReadUSPostalPlaces(archivepath, func(place *PostalPlace) error {
			if len(place.PostalCode) == 0 || len(place.PlaceName) == 0 {
				return nil
			}

			places[place.PostalCode] = append(places[place.PostalCode], place)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return places, nil
}

// PostalCodesByArea reads the archives into the ZIP Codes GeoNames names for
// each area: bystate is keyed by USPS state abbreviation and bycounty by that
// abbreviation followed by the county equivalent's FIPS code, as "MP100".
//
// An area with exactly one ZIP Code tells a caller something a bloom filter
// cannot: every address in it carries that ZIP Code. See
// poetic-systems/zipcity#7, where that is the only thing left that says which
// ZIP Code an American Samoa street is in.
func PostalCodesByArea(archivepaths []string) (bystate, bycounty map[string][]string, err error) {
	bystate = map[string][]string{}
	bycounty = map[string][]string{}

	add := func(area map[string][]string, key, postalcode string) {
		if len(key) == 0 || slices.Contains(area[key], postalcode) {
			return
		}

		area[key] = append(area[key], postalcode)
	}

	for _, archivepath := range archivepaths {
		err := ReadUSPostalPlaces(archivepath, func(place *PostalPlace) error {
			if len(place.PostalCode) == 0 {
				return nil
			}

			add(bystate, place.StateUSPS, place.PostalCode)
			if len(place.CountyCode) > 0 {
				add(bycounty, place.StateUSPS+place.CountyCode, place.PostalCode)
			}

			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return bystate, bycounty, nil
}
