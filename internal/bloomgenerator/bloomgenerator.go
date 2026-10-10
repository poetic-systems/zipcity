//go:build ignore
// +build ignore

// bloomgenerator.go - Run via 'go run internal/bloomgenerator/bloomgenerator.go' to generate
// compiled_filters.go and the related binary files it embeds.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/format"
	"io"
	"log"
	"maps"
	"os"
	"path"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"text/template"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/areazip"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
	"github.com/poetic-systems/zipcity/internal/usgeonames"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
	"github.com/poetic-systems/zipcity/internal/usziplocale"
	"github.com/poetic-systems/zipcity/internal/zipcities"
)

// zipStreetFalsePositiveRate and cityStreetFalsePositiveRate are the false
// positive rates the zip-street and city-street filters are each built
// with. They are written into the generated compiled_filter package beside
// the filters (as ZipStreetFalsePositiveRate and CityStreetFalsePositiveRate)
// so a caller weighing a Match can read the rate a filter was actually
// generated with instead of a copy that can drift. They are separate
// constants, rather than one shared by every filter, so that one filter
// type's rate can move without an API change to the rest — see
// poetic-systems/zipcity#53.
const (
	zipStreetFalsePositiveRate  = 0.00075
	cityStreetFalsePositiveRate = 0.00075
)

type ZipStreetTuple struct {
	Zip    string
	Street string
}

type ZipCityTuple struct {
	Zip   string
	City  string
	State string
}

type CityStreetTuple struct {
	City   string
	State  string
	Street string
}

var nonalpha = regexp.MustCompile(`[^a-zA-Z ]+`)
var nonalphanum = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func main() {
	// TODO: See if we actually need all of the streets of if we
	// can get away with just encoding the challenging ones: the ones
	// that start with directionals, those that are numeric or
	// alphabetic and might need to be spelled out or not, etc.

	// NOTE: we are writing the city-street map directly to binary files,
	// bucketed by state or territory into about 56 files totaling about
	// 12MB, and loaded via go:embed. We break the zip-street relation up
	// into 100 different binary files, totaling about 11MB, to keep them
	// small and isolate changes — a regeneration only touches the buckets
	// whose data actually changed. These are also loaded via go:embed. The
	// zip-city relation is instead written as zip-city-names.tsv (about
	// 1.2MB) and read back exactly, rather than built into a filter of its
	// own — see poetic-systems/zipcity#55. The cumulative size of the
	// compiled filter directory is now about 23MB.

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	embeddedDir := path.Join(
		cwd,
		"generated",
		"embedded_filter",
	)
	filterDir := path.Join(
		embeddedDir,
		"data",
	)

	// Data Ingestion Loop
	// loop through the parsed US Census Bureau TIGER files here
	zipStreetData := map[string]map[string]ZipStreetTuple{}
	zipCityData := map[string]ZipCityTuple{}
	cityStreetData := map[string]map[string]CityStreetTuple{}

	// Cache the census data locally if we don't already have it
	prefixes, absent, err := ustigerline.DownloadAllRequiredTigerfiles()
	if err != nil {
		panic(err)
	}

	// remove the county-level fips code from the first prefix to get the base prefix
	// for requesting state fips codes and abbreviations
	baseprefix := prefixes[0][0 : len(prefixes[0])-6]
	stateMap, err := ustigerline.ReadStates(baseprefix)
	if err != nil {
		panic(err)
	}

	// TIGER gives a side its city from the PLACEFP of the face beside it,
	// which is blank outside an incorporated place. The Postal Service
	// addresses those to the city of the delivering post office instead, so
	// the city exists but is not in TIGER. GeoNames is where we get it.
	// See poetic-systems/zipcity#5.
	geonamespaths, geonamesabsent, err := usgeonames.DownloadUSPostalCodes()
	if err != nil {
		panic(err)
	}
	for country, err := range geonamesabsent {
		log.Printf("Warning: no GeoNames postal codes for %s: %s", country, err)
	}
	placesByZip, err := usgeonames.PlacesByPostalCode(geonamespaths)
	if err != nil {
		panic(err)
	}
	stateZips, countyZips, err := usgeonames.PostalCodesByArea(geonamespaths)
	if err != nil {
		panic(err)
	}

	// The Postal Service's own ZIP Locale Detail table names the unit that
	// delivers each ZIP Code. That is not the same claim as GeoNames' postal
	// city or TIGER's place, so it goes into the ZIP Code to city relation
	// under its own source rather than mixed in with them, and it does not go
	// into the street filters at all: nothing here saw a street beside it.
	// Military mail is the one place the table is the better source outright,
	// naming 571 codes individually where GeoNames has far fewer and never
	// writes DPO. See poetic-systems/zipcity#48.
	zldpath, err := usziplocale.Download()
	if err != nil {
		panic(err)
	}
	militaryzips, err := usziplocale.MilitaryZips(zldpath)
	if err != nil {
		panic(err)
	}
	postoffices, err := usziplocale.PostOfficesByZip(zldpath)
	if err != nil {
		panic(err)
	}

	// Every ZIP Code GeoNames knows contributes its city directly. TIGER can
	// only offer a pair where it has both a place and an address range, so
	// this is the whole of the zip-city relation and TIGER adds to it rather
	// than the other way round.
	for zip, places := range placesByZip {
		for _, place := range places {
			key := bloomkeys.KeyZipCity(zip, place.PlaceName)
			_, found := zipCityData[key]
			if !found {
				zipCityData[key] = ZipCityTuple{
					Zip:   zip,
					City:  place.PlaceName,
					State: place.StateUSPS,
				}
			}
		}
	}
	numGeonamesZip2City := uint(len(zipCityData))

	numZip2City := numGeonamesZip2City
	numZip2Sreet := uint(0)
	numCity2Street := uint(0)
	numStreetOnly := uint(0)
	// additivezips are the ZIP Codes that gain a city-street key from the
	// additive postal city rule below — the ones the either/or reading used to
	// lose. Counted rather than asserted because the change is only worth what
	// it actually reaches. See poetic-systems/zipcity#59.
	additivezips := map[string]struct{}{}
	// Reading a county is most of a generation and each one is independent,
	// so they are read several at a time and merged here in order. Half the
	// CPUs is a working guess: the largest county (Los Angeles) reads in
	// about 2.5 GB, the rest in far less. See #46.
	for county := range ustigerline.ReadCounties(prefixes, max(1, runtime.NumCPU()/2)) {
		pre, allSides, err := county.Prefix, county.Sides, county.Err
		if err != nil {
			log.Printf("Error from ustigerline.ReadStreetSides(): %s", err)
			continue
		}
		log.Printf("Read %s: %d street sides", pre, len(allSides))

		statefips := pre[len(pre)-5 : len(pre)-3]
		countyfips := pre[len(pre)-3:]
		stateInfo := stateMap[statefips]

		// Where GeoNames names exactly one ZIP Code for a county equivalent —
		// or for a territory that has no counties — every address in it carries
		// that ZIP Code, whether or not TIGER describes an address range for the
		// side, and its own ranges get a veto over that. Both rules, and the
		// evidence for them, are in internal/areazip.
		countyzip := areazip.Sole(stateZips, countyZips, stateInfo.USPS, countyfips)
		if len(countyzip) > 0 && areazip.Contradicted(allSides, countyzip) {
			log.Printf("Note: address ranges in %s name a ZIP Code other than %s, so it is not lent", pre, countyzip)
			countyzip = ""
		}

		for _, side := range allSides {
			// A side with no ZIP Code of its own borrows the one its county
			// equivalent has, where the county has one to lend.
			zips := side.Zips
			if len(zips) == 0 && len(countyzip) > 0 {
				zips = []string{countyzip}
			}
			cty := ""
			if side.City != nil {
				cty = side.City.Name
			}
			// A side inside an incorporated place still has a second city on
			// an envelope: the post office that delivers its ZIP Code, which
			// is often not the place it sits in. So the names are additive —
			// the place TIGER gives the side, and every postal city GeoNames
			// names for the side's ZIP Codes — rather than the place where
			// there is one and the postal cities only otherwise.
			//
			// Both are names a caller legitimately writes on that street, and
			// taking only the place lost the postal city for every side in a
			// place whose delivering office has a different name: an address
			// in DanDan, Saipan is delivered from SAIPAN MP, and the
			// city-street key SAIPAN|MP|CHALAN TUN HERMAN PAN was missing
			// because TIGER had a place name to offer. All of them are taken
			// rather than one, because nothing here can tell which post
			// office serves which end of the street. Duplicates are left to
			// the key map. See poetic-systems/zipcity#59.
			postalcities := []string{}
			if len(cty) > 0 {
				postalcities = append(postalcities, cty)
			}
			for _, zip := range zips {
				for _, place := range placesByZip[zip] {
					postalcities = append(postalcities, place.PlaceName)
					// The ZIP Codes where the additive rule is the whole
					// difference: TIGER had a place, so the postal city used
					// to be dropped, and it is a different name.
					if len(cty) > 0 && bloomkeys.City(place.PlaceName) != bloomkeys.City(cty) {
						additivezips[zip] = struct{}{}
					}
				}
			}
			street := ""
			streetnames := []string{}
			if side.Street != nil {
				street = side.Street.Name
				// Only the Pub 28 renderings are keyed. Name is the edges
				// FULLNAME, TIGER's own abbreviated spelling (I- 5, UNION
				// PACIFIC RR, CLL LOIZA), which no caller in spec form sends.
				// See #1.
				streetnames = side.Street.Alt
			}

			// Addresses at each end of a street may be served by different
			// ZIP Codes, so a side carries every ZIP Code its address ranges
			// name rather than one.
			for _, zip := range zips {
				if len(cty) > 0 && len(zip) > 4 {
					key := bloomkeys.KeyZipCity(zip, cty)
					_, found := zipCityData[key]
					if !found {
						zipCityData[key] = ZipCityTuple{
							Zip:   zip,
							City:  cty,
							State: stateInfo.USPS,
						}
						numZip2City += 1
					}
				}
				if len(zip) > 4 && len(street) > 0 {
					// NOTE: We store and load bloom filters on a per zip code prefix basis.
					zipscope := zip[0:2]
					scoped, exists := zipStreetData[zipscope]
					if !exists {
						scoped = make(map[string]ZipStreetTuple, 0)
					}
					for _, stname := range streetnames {
						key := bloomkeys.KeyZipStreet(zip, stname)
						_, found := scoped[key]
						if !found {
							scoped[key] = ZipStreetTuple{
								Zip:    zip,
								Street: stname,
							}
							zipStreetData[zipscope] = scoped
							numZip2Sreet += 1
						}
					}
				}
			}
			if len(postalcities) > 0 && len(street) > 0 {
				for _, cityname := range postalcities {
					for _, stname := range streetnames {
						stateCityStreetData, ok := cityStreetData[stateInfo.USPS]
						if !ok {
							stateCityStreetData = make(map[string]CityStreetTuple)
							cityStreetData[stateInfo.USPS] = stateCityStreetData
						}
						key := bloomkeys.KeyCityStateStreet(cityname, stateInfo.USPS, stname)
						_, found := stateCityStreetData[key]
						if !found {
							stateCityStreetData[key] = CityStreetTuple{
								City:   cityname,
								State:  stateInfo.USPS,
								Street: stname,
							}
							numCity2Street += 1
						}
					}
				}
			}
			if len(street) > 0 && len(cty) == 0 && len(zips) == 0 {
				numStreetOnly += 1
			}
		}
	}
	log.Printf("Counts - zip-city: %d (%d from GeoNames) zip-street: %d city-street: %d street-only: %d additive-postal-city zips: %d",
		numZip2City, numGeonamesZip2City, numZip2Sreet, numCity2Street, numStreetOnly, len(additivezips))

	// Initialize Bloom Filters
	// Estimates for US: ~30M unique combinations. FPR: 0.075% (0.00075)
	// With several address range files absent (mostly for islands):
	//  Counts - zip-city: 4396668 zip-street: 20604757 city-street: 4396668

	// NOTE: We don't attempt to compress the bloom filter bytes because a
	// bloom filter should have a relatively random distribution of bits set
	// if its hashing algorithm is working properly.

	// fileHashes carries one SHA-256 per compiled filter file, keyed by the
	// same Go identifier compiled_filter.go const it loads under
	// (CityStreetAK, ZipStreet00, ...), so a caller can verify just the one
	// file it actually reads rather than a hash that only resolves once
	// every filter has been read — see poetic-systems/zipcity#73 and the
	// per-filter lazy decode in internal/bloomdata.
	fileHashes := map[string]string{}

	uspsstatekeys := slices.Collect(maps.Keys(cityStreetData))
	slices.Sort(uspsstatekeys)

	citystreetfiles := make(map[string]string, 0)

	for _, uspsstate := range uspsstatekeys {
		stateCityStreetData := cityStreetData[uspsstate]
		numThisCity2Street := uint(len(stateCityStreetData))
		cityStreetFilter := bloom.NewWithEstimates(numThisCity2Street, cityStreetFalsePositiveRate)

		for key := range stateCityStreetData {
			cityStreetFilter.Add([]byte(key))
		}

		// Serialize the City to Street Bloom Filter
		csVarName := fmt.Sprintf("CityStreet%s", uspsstate)
		csidentifier := fmt.Sprintf("%s-%s", "city-street", uspsstate)
		citystreetfiles[csVarName] = csidentifier

		csHash, err := serialize(
			path.Join(
				filterDir,
				bloomfilename.Filename(csidentifier),
			),
			cityStreetFilter,
		)
		if err != nil {
			panic(err)
		}
		fileHashes[csVarName] = csHash
	}

	zipscopekeys := slices.Collect(maps.Keys(zipStreetData))
	slices.Sort(zipscopekeys)
	zipstreetfiles := make(map[string]string, 0)
	for _, zipscope := range zipscopekeys {
		streets := zipStreetData[zipscope]
		numThisZip2Street := uint(len(streets))
		streetFilter := bloom.NewWithEstimates(numThisZip2Street, zipStreetFalsePositiveRate)

		for key := range streets {
			streetFilter.Add([]byte(key))
		}

		// Serialize the scoped Zip to Street Bloom Filter
		zsVarName := fmt.Sprintf("ZipStreet%s", zipscope)
		zsidentifier := fmt.Sprintf("%s-%s", "zip-street", zipscope)
		zipstreetfiles[zsVarName] = zsidentifier

		zsHash, err := serialize(
			path.Join(
				filterDir,
				bloomfilename.Filename(zsidentifier),
			),
			streetFilter,
		)
		if err != nil {
			panic(err)
		}
		fileHashes[zsVarName] = zsHash
	}

	// The ZIP Code to city name relation, off zipCityData itself, written
	// out so a caller holding a ZIP Code and no city has something to read.
	// CheckZipAndCity reads it back too, for an exact answer over the same
	// data — see poetic-systems/zipcity#55. The state is the one the source
	// placed the name in: GeoNames' admin code for its rows, the county's
	// state for TIGER's. See poetic-systems/zipcity#17.
	names := make(zipcities.Table, len(zipCityData))
	for _, pair := range zipCityData {
		names.Add(pair.Zip, pair.State, pair.City, zipcities.Seen)
	}

	// The Postal Service's delivery unit for each ZIP Code, under its own
	// source so a caller can tell the recommendation from the sightings.
	//
	// A military code takes its state from the service area rather than from
	// the row, because the row names the gateway that accepts the mail
	// (JERSEY CITY NJ, MIAMI FL, SAN FRANCISCO CA) and an address to an
	// overseas post office is written AA, AE or AP — the point of
	// poetic-systems/zipcity#24. A code the service areas do not cover is
	// reported rather than filed under no state at all. Everywhere else the
	// row's own PHYSICAL STATE is the state the delivering office stands in,
	// which is the state that name belongs to.
	numMilitaryZip2City := 0
	uspszips := map[string]struct{}{}
	for _, zip := range slices.Sorted(maps.Keys(militaryzips)) {
		unit := militaryzips[zip]
		state := usgeonames.MilitaryState(zip)
		if state == "" {
			log.Printf("Warning: military ZIP Code %s (%s) is in no known service area", zip, unit.Name)
			continue
		}
		names.Add(zip, state, unit.Name, zipcities.USPS)
		uspszips[zip] = struct{}{}
		numMilitaryZip2City += 1
	}
	numUspsZip2City := 0
	differing := []string{}
	for _, zip := range slices.Sorted(maps.Keys(postoffices)) {
		if _, military := militaryzips[zip]; military {
			continue
		}
		for _, office := range postoffices[zip] {
			names.Add(zip, office.PhysicalState, office.Name, zipcities.USPS)
			uspszips[zip] = struct{}{}
			numUspsZip2City += 1
			// PHYSICAL CITY is the town the office building stands in, which
			// is not always what it is called: it is measured here and left
			// out of the table until there is a reason to believe it is a
			// last line somebody writes. See poetic-systems/zipcity#48.
			if office.PhysicalCity != "" && bloomkeys.City(office.PhysicalCity) != bloomkeys.City(office.Name) {
				differing = append(differing, fmt.Sprintf("%s %s/%s", zip, office.Name, office.PhysicalCity))
			}
		}
	}
	log.Printf("Counts - usps zip-city: %d names over %d ZIP Codes, %d of them military; PHYSICAL CITY differs from LOCALE NAME on %d rows, first 20: %s",
		numUspsZip2City+numMilitaryZip2City, len(uspszips), numMilitaryZip2City,
		len(differing), strings.Join(differing[:min(20, len(differing))], ", "))

	zipCityNamesHash, err := writeZipCityNames(
		path.Join(filterDir, "zip-city-names.tsv"),
		names,
	)
	if err != nil {
		panic(err)
	}

	// Generate the Go source code containing the filter identifiers
	tmpl_compiled := `// Code generated by internal/bloomgenerator/bloomgenerator.go. DO NOT EDIT.
package compiled_filter

import (
	"fmt"
	"iter"
	"maps"
	"regexp"
	"strings"
)

// DataFileSHA256 carries one SHA-256 per compiled filter file, so a caller
// can verify just the one file it actually reads rather than a hash that
// only resolves once every filter has been read. See
// poetic-systems/zipcity#73 and the per-filter lazy decode in
// internal/bloomdata.
var DataFileSHA256 = map[CompiledFilter]string{
{{- range $name, $hash := .DataFileHashes }}
	{{ $name }}: "{{ $hash }}",
{{- end }}
}

const ZipCityNamesSHA256 = "{{- .ZipCityNamesSHA256 -}}"

var zip5pattern = regexp.MustCompile({{ tick }}^\d{5}${{ tick }})

// ZipStreetFalsePositiveRate and CityStreetFalsePositiveRate are the false
// positive rates the zip-street and city-street filters in this package
// were each built with (bloom.NewWithEstimates in
// internal/bloomgenerator/bloomgenerator.go). They are generated here,
// beside the filters, so neither can drift from what its filter was
// actually built with.
const (
	ZipStreetFalsePositiveRate  = {{ .ZipStreetFalsePositiveRate }}
	CityStreetFalsePositiveRate = {{ .CityStreetFalsePositiveRate }}
)

// AbsentSources names, per TIGER area code, the source file types the Census
// Bureau published nothing of at generation time. Read off the Census Bureau's
// own index each generation rather than from a list kept here.
//
// A file the index does list must load or generation stops, so these are the
// only gaps in what the filters were built from. An area named here is one the
// filters know less about than the rest; absence from the filters is weaker
// evidence there than elsewhere. See poetic-systems/zipcity#2.
var AbsentSources = map[string][]string{
{{- range .Absent }}
	"{{ .Area }}": { {{- range $i, $t := .Types }}{{ if $i }}, {{ end }}"{{ $t }}"{{ end -}} },
{{- end }}
}

func toCompiledFilter(in string) (CompiledFilter, error) {
	cf, ok := allCompiledFilters[in]
	if ok {
		return cf, nil
	}
	return Unrecognized, fmt.Errorf("Unrecognized filter name")
}

func CityStreetFilterForState(state string) (CompiledFilter, error) {
	if len(state) != 2 {
		return Unrecognized, fmt.Errorf("USPS state abbreviation required")
	}

	filterid := fmt.Sprintf("city-street-%s", strings.ToUpper(state))
	cf, err := toCompiledFilter(filterid)
	if err != nil {
		return Unrecognized, fmt.Errorf("Supported USPS state abbreviation required")
	}
	return cf, nil
}

func ZipStreetFilterForZip(zip string) (CompiledFilter, error) {
	if !zip5pattern.MatchString(zip) {
		return Unrecognized, fmt.Errorf("5-digit zip code required")
	}
	zip2 := zip[0:2]
	filterid := fmt.Sprintf("zip-street-%s", zip2)
	cf, err := toCompiledFilter(filterid)
	if err != nil {
		return Unrecognized, fmt.Errorf("known 5-digit zip code required")
	}
	return cf, nil
}

type CompiledFilter string

const (
	Unrecognized CompiledFilter = ""
{{- range $varName, $csidentifier := .CSFiles }}
	{{ $varName }}   CompiledFilter = "{{- $csidentifier -}}"
{{- end }}
{{- range $varName, $zsidentifier := .ZSFiles }}
	{{ $varName }}   CompiledFilter = "{{- $zsidentifier -}}"
{{- end }}
)

var allCompiledFilters = map[string]CompiledFilter{
{{- range $varName, $csidentifier := .CSFiles }}
	"{{- $csidentifier -}}": {{ $varName }},
{{- end }}
{{- range $varName, $zsidentifier := .ZSFiles }}
	"{{- $zsidentifier -}}": {{ $varName }},
{{- end }}
}

func All() iter.Seq[CompiledFilter] {
	return maps.Values(allCompiledFilters)
}

`

	templateFuncMap := template.FuncMap{
		"tick": func() string { return "`" },
	}

	tcf := template.Must(template.New("compiled_filter").Funcs(templateFuncMap).Parse(tmpl_compiled))

	err = os.MkdirAll("./generated/compiled_filter", 0755)
	if err != nil {
		panic(err)
	}

	// Format what the template rendered rather than trusting its whitespace,
	// so the generated package is gofmt-clean however the template is written.
	var rendered_compiled_filters bytes.Buffer
	err = tcf.Execute(&rendered_compiled_filters, map[string]interface{}{
		"ZSFiles":                     zipstreetfiles,
		"CSFiles":                     citystreetfiles,
		"Absent":                      absentRows(absent),
		"ZipStreetFalsePositiveRate":  zipStreetFalsePositiveRate,
		"CityStreetFalsePositiveRate": cityStreetFalsePositiveRate,
		"DataFileHashes":              fileHashes,
		"ZipCityNamesSHA256":          zipCityNamesHash,
	})
	if err != nil {
		panic(err)
	}
	formatted_compiled, err := format.Source(rendered_compiled_filters.Bytes())
	if err != nil {
		panic(err)
	}
	compiledOutName := "./generated/compiled_filter/compiled_filter.go"
	if err := os.WriteFile(compiledOutName, formatted_compiled, 0644); err != nil {
		panic(err)
	}
	log.Printf("Successfully generated %s", compiledOutName)

	// Generate the Go source code containing the embedded asset
	tmpl_embedded := `// Code generated by internal/bloomgenerator/bloomgenerator.go. DO NOT EDIT.
package embedded_filter

import (
	"embed"
	"io/fs"
)

//go:embed data/*
var bloomDir embed.FS

func PrepareFS() (fs.FS, error) {
	return bloomDir, nil
}
`
	tembd := template.Must(template.New("filter").Funcs(templateFuncMap).Parse(tmpl_embedded))

	err = os.MkdirAll(filterDir, 0755)
	if err != nil {
		panic(err)
	}

	// Format what the template rendered rather than trusting its whitespace,
	// so the generated package is gofmt-clean however the template is written.
	var rendered_embedded_filters bytes.Buffer
	err = tembd.Execute(&rendered_embedded_filters, map[string]interface{}{
		// Not sure we need any context here. We only generate this file because it
		// is in the generated directory with the files it embeds
	})
	if err != nil {
		panic(err)
	}
	formatted_embedded, err := format.Source(rendered_embedded_filters.Bytes())
	if err != nil {
		panic(err)
	}
	embeddedOutName := path.Join(
		embeddedDir,
		"embedded_filter.go",
	)
	if err := os.WriteFile(embeddedOutName, formatted_embedded, 0644); err != nil {
		panic(err)
	}
	log.Printf("Successfully generated %s", embeddedOutName)
}

// absentRows puts the absent-source report in a fixed order, so that two
// generations over the same index write the same bytes.
func absentRows(absent ustigerline.AbsentSources) []struct {
	Area  string
	Types []string
} {
	rows := make([]struct {
		Area  string
		Types []string
	}, 0, len(absent))
	for _, area := range slices.Sorted(maps.Keys(absent)) {
		rows = append(rows, struct {
			Area  string
			Types []string
		}{Area: area, Types: absent[area]})
	}

	return rows
}

// writeZipCityNames writes the ZIP Code to city name table beside the
// compiled filters, where it is committed as the readable form of the
// zip-city relation, and embedded so CitiesKnownFor and CheckZipAndCity can
// both read it back.
func writeZipCityNames(filename string, names zipcities.Table) (string, error) {
	err := os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	writer := io.MultiWriter(file, h)

	if err := zipcities.Encode(writer, names); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func serialize(filename string, data *bloom.BloomFilter) (string, error) {
	filedir := path.Dir(filename)
	err := os.MkdirAll(filedir, 0755)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	writer := io.MultiWriter(file, h)

	_, err = data.WriteTo(writer)
	if err != nil {
		return "", fmt.Errorf("Failed to write bloom filter to disk: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
