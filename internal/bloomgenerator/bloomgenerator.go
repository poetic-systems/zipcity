//go:build ignore
// +build ignore

// generator.go - Run via 'go run internal/bloomgenerator/bloomgenerator.go' to generate filters.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/areazip"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
	"github.com/poetic-systems/zipcity/internal/usgeonames"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
	"github.com/poetic-systems/zipcity/internal/zipcities"
)

type ZipStreetTuple struct {
	Zip    string
	Street string
}

type ZipCityTuple struct {
	Zip  string
	City string
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

	// NOTE: we are writing the zip-city and city-street maps directly to
	// binary files that are about 6MB and loaded via go:embed. We break
	// the zip-street relation up in to 100 different binary files to keep
	// them small and isolate changes. These are also loaded via go:embed.
	// The previous strategy of writing the bytes directly into the
	// generated template file used ~16 bits per bit of data, resulting
	// in a 100MB generated source file. The cumulative size of the
	// compiled filter directory is now about 26 MB.

	now := time.Now()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filterDir := path.Join(
		cwd,
		"generated",
		"compiled_filter",
		"bloom_filters",
	)

	// Data Ingestion Loop
	// loop through the parsed US Census Bureau TIGER files here
	zipStreetData := map[string]map[string]ZipStreetTuple{}
	zipCityData := map[string]ZipCityTuple{}
	cityStreetData := map[string]map[string]CityStreetTuple{}
	streetOnlyData := map[string]*ustigerline.StreetSide{}

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
		fmt.Printf("Warning: no GeoNames postal codes for %s: %s\n", country, err)
	}
	placesByZip, err := usgeonames.PlacesByPostalCode(geonamespaths)
	if err != nil {
		panic(err)
	}
	stateZips, countyZips, err := usgeonames.PostalCodesByArea(geonamespaths)
	if err != nil {
		panic(err)
	}

	// Every ZIP Code GeoNames knows contributes its city directly. TIGER can
	// only offer a pair where it has both a place and an address range, so
	// this is the whole of the zip-city relation and TIGER adds to it rather
	// than the other way round.
	for zip, places := range placesByZip {
		for _, place := range places {
			key, err := bloomkeys.KeyZipCity(zip, place.PlaceName)
			if err != nil {
				panic(err)
			}
			_, found := zipCityData[key]
			if !found {
				zipCityData[key] = ZipCityTuple{
					Zip:  zip,
					City: place.PlaceName,
				}
			}
		}
	}
	numGeonamesZip2City := uint(len(zipCityData))

	numZip2City := numGeonamesZip2City
	numZip2Sreet := uint(0)
	numCity2Street := uint(0)
	numStreetOnly := uint(0)
	for _, pre := range prefixes {
		allSides, err := ustigerline.ReadStreetSides(pre)
		if err != nil {
			fmt.Printf("Error from ustigerline.ReadStreetSides(): %s\n", err)
			continue
		}

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
			fmt.Printf("Note: address ranges in %s name a ZIP Code other than %s, so it is not lent\n", pre, countyzip)
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
			// A side outside any incorporated place still has a city on an
			// envelope: the post office that delivers its ZIP Code. Those
			// cities are the ones the city-street relation is missing, so a
			// side with no place of its own borrows every city GeoNames
			// names for its ZIP Codes. It borrows all of them rather than
			// one, because nothing here can tell which post office serves
			// which end of the street.
			postalcities := []string{}
			if len(cty) > 0 {
				postalcities = append(postalcities, cty)
			} else {
				for _, zip := range zips {
					for _, place := range placesByZip[zip] {
						postalcities = append(postalcities, place.PlaceName)
					}
				}
			}
			street := ""
			alts := ""
			if side.Street != nil {
				street = side.Street.Name
				altbytes, err := json.Marshal(side.Street.Alt)
				if err != nil {
					alts = string(altbytes)
				}
			}
			// if (len(cty) > 0 && nonalpha.MatchString(cty)) ||
			// 	(len(street) > 0 && nonalphanum.MatchString(street)) ||
			// 	(len(side.Zip) > 0 && nonalphanum.MatchString(side.Zip)) {
			// 	fmt.Printf("Non-alphabetical characters found in name associated with side. Zip: %s City: %s Street: %s\n", side.Zip, cty, street)
			// }
			if len(street) > 0 && (strings.Contains(street, "Ó ") ||
				strings.Contains(alts, "Ó ")) {
				// fmt.Printf("found Zip: %s City: %s Street: %s or: %s\n", side.Zip, cty, street, alts)
			}
			// Addresses at each end of a street may be served by different
			// ZIP Codes, so a side carries every ZIP Code its address ranges
			// name rather than one.
			for _, zip := range zips {
				if len(cty) > 0 && len(zip) > 4 {
					key, err := bloomkeys.KeyZipCity(zip, cty)
					if err != nil {
						panic(err)
					}
					_, found := zipCityData[key]
					if !found {
						zipCityData[key] = ZipCityTuple{
							Zip:  zip,
							City: cty,
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
					// make sure we include the primary name as well as the alternative names
					streetnames := append(side.Street.Alt, street)
					for _, stname := range streetnames {
						key, err := bloomkeys.KeyZipStreet(zip, stname)
						if err != nil {
							panic(err)
						}
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
				// make sure we include the primary name as well as the alternative names
				streetnames := append(side.Street.Alt, street)
				for _, cityname := range postalcities {
					for _, stname := range streetnames {
						stateCityStreetData, ok := cityStreetData[stateInfo.USPS]
						if !ok {
							stateCityStreetData = make(map[string]CityStreetTuple)
							cityStreetData[stateInfo.USPS] = stateCityStreetData
						}
						key, err := bloomkeys.KeyCityStateStreet(cityname, stateInfo.USPS, stname)
						if err != nil {
							panic(err)
						}
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
				streetOnlyData[street] = side
				numStreetOnly += 1
			}
		}
	}
	fmt.Printf("Counts - zip-city: %d (%d from GeoNames) zip-street: %d city-street: %d street-only: %d\n",
		numZip2City, numGeonamesZip2City, numZip2Sreet, numCity2Street, numStreetOnly)

	// Initialize Bloom Filters
	// Estimates for US: ~30M unique combinations. FPR: 0.1% (0.001)
	// With several address range files absent (mostly for islands):
	//  Counts - zip-city: 4396668 zip-street: 20604757 city-street: 4396668

	// NOTE: We don't attempt to compress the bloom filter bytes because a
	// bloom filter should have a relatively random distribution of bits set
	// if its hashing algorithm is working properly.

	zipstreetfiles := make(map[string]string, 0)
	for zipscope, streets := range zipStreetData {
		numThisZip2Street := uint(len(streets))
		// Add ~1/8 of overhead to the count for the base capacity
		nZS := numThisZip2Street + (numThisZip2Street >> 3)
		streetFilter := bloom.NewWithEstimates(nZS, 0.005)

		for key := range streets {
			streetFilter.Add([]byte(key))
		}

		// Serialize the scoped Zip to Street Bloom Filter
		zsVarName := fmt.Sprintf("ZipStreet%s", zipscope)
		zsidentifier := fmt.Sprintf("%s-%s", "zip-street", zipscope)
		zipstreetfiles[zsVarName] = zsidentifier

		err := serialize(
			path.Join(
				filterDir,
				bloomfilename.Filename(zsidentifier),
			),
			streetFilter,
		)
		if err != nil {
			panic(err)
		}
	}

	// The same relation the zip-city filter is built from, written out so a
	// caller holding a ZIP Code and no city has something to read rather than
	// only something to ask. Written before the filter is built, off
	// zipCityData itself, so the table and the filter cannot disagree about
	// what we have seen. See poetic-systems/zipcity#17.
	names := make(zipcities.Table, len(zipCityData))
	for _, pair := range zipCityData {
		names.Add(pair.Zip, pair.City)
	}
	namesize, err := zipcities.Measure(names)
	if err != nil {
		panic(err)
	}
	fmt.Println(namesize)
	err = writeZipCityNames(
		path.Join(cwd, "generated", "compiled_filter", "zip-city-names.tsv"),
		names,
	)
	if err != nil {
		panic(err)
	}

	// Add ~1/8 of overhead to the count for the base capacity
	nZC := numZip2City + (numZip2City >> 3)
	cityFilter := bloom.NewWithEstimates(nZC, 0.005)

	for key := range zipCityData {
		cityFilter.Add([]byte(key))
	}

	// Serialize the Zip to City Bloom Filter

	err = serialize(
		path.Join(
			filterDir,
			bloomfilename.Filename("zip-city"),
		),
		cityFilter,
	)
	if err != nil {
		panic(err)
	}

	citystreetfiles := make(map[string]string, 0)

	for uspsstate, stateCityStreetData := range cityStreetData {
		numThisCity2Street := uint(len(stateCityStreetData))
		// Add ~1/8 of overhead to the count for the base capacity
		nCS := numThisCity2Street + (numThisCity2Street >> 3)
		cityStreetFilter := bloom.NewWithEstimates(nCS, 0.005)

		for key := range stateCityStreetData {
			cityStreetFilter.Add([]byte(key))
		}

		// Serialize the City to Street Bloom Filter
		csVarName := fmt.Sprintf("CityStreet%s", uspsstate)
		csidentifier := fmt.Sprintf("%s-%s", "city-street", uspsstate)
		citystreetfiles[csVarName] = csidentifier

		err = serialize(
			path.Join(
				filterDir,
				bloomfilename.Filename(csidentifier),
			),
			cityStreetFilter,
		)
		if err != nil {
			panic(err)
		}
	}

	// Generate the Go source code containing the embedded asset
	tmpl := `// DO NOT EDIT! Code generated at {{ .Now }} by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"regexp"
	"strings"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
)

var zip5pattern = regexp.MustCompile({{ tick }}^\d{5}${{ tick }})

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

//go:embed bloom_filters/*.bin
var bloomDir embed.FS

// Keep a map of the raw, zero-allocation byte arrays
var bloom_filters = make(map[CompiledFilter]*bloom.BloomFilter)

// Initialize the bloom filter map immediately
func init() {
	for key, name := range allCompiledFilters {

		keypath := path.Join("bloom_filters", bloomfilename.Filename(key))

		// ReadFile allocates each byte slice once during boot
		data, err := bloomDir.ReadFile(keypath)
		if err != nil {
			panic(err)
		}

		filter := &bloom.BloomFilter{}
		reader := bytes.NewReader(data)
		_, err = filter.ReadFrom(reader)
		if err != nil {
			panic(fmt.Errorf("Failed to read %s bloom filter: %w", name, err))
		}

		bloom_filters[name] = filter
	}
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

// LoadFilter restores the compiled filter in memory
func LoadFilter(name CompiledFilter) (*bloom.BloomFilter, error) {
	filter, ok := bloom_filters[name]
	if !ok {
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	return filter, nil
}

type CompiledFilter string

const (
	Unrecognized CompiledFilter = ""
	ZipCity      CompiledFilter = "zip-city"
{{- range $varName, $csidentifier := .CSFiles }}
	{{ $varName }}   CompiledFilter = "{{- $csidentifier -}}"
{{- end }}
{{- range $varName, $zsidentifier := .ZSFiles }}
	{{ $varName }}   CompiledFilter = "{{- $zsidentifier -}}"
{{- end }}
)

var allCompiledFilters = map[string]CompiledFilter{
	"zip-city":       ZipCity,
{{- range $varName, $csidentifier := .CSFiles }}
	"{{- $csidentifier -}}": {{ $varName }},
{{- end }}
{{- range $varName, $zsidentifier := .ZSFiles }}
	"{{- $zsidentifier -}}": {{ $varName }},
{{- end }}
}

`
	templateFuncMap := template.FuncMap{
		"tick": func() string { return "`" },
	}
	t := template.Must(template.New("filter").Funcs(templateFuncMap).Parse(tmpl))

	err = os.MkdirAll("./generated/compiled_filter/bloom_filters", 0755)
	if err != nil {
		panic(err)
	}
	// Format what the template rendered rather than trusting its whitespace,
	// so the generated package is gofmt-clean however the template is written.
	var rendered bytes.Buffer
	err = t.Execute(&rendered, map[string]interface{}{
		"ZSFiles": zipstreetfiles,
		"CSFiles": citystreetfiles,
		"Absent":  absentRows(absent),
		"Now":     now.UTC().Format(time.RFC3339),
	})
	if err != nil {
		panic(err)
	}
	formatted, err := format.Source(rendered.Bytes())
	if err != nil {
		panic(err)
	}
	outName := "./generated/compiled_filter/compiled_filters.go"
	if err := os.WriteFile(outName, formatted, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("Successfully generated %s\n", outName)
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
// compiled filters. Nothing embeds it yet: what it costs decides whether it
// ships, and that is read off the file a full generation writes.
func writeZipCityNames(filename string, names zipcities.Table) error {
	err := os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return zipcities.Encode(file, names)
}

func serialize(filename string, data *bloom.BloomFilter) error {
	filedir := path.Dir(filename)
	err := os.MkdirAll(filedir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = data.WriteTo(file)
	if err != nil {
		return fmt.Errorf("Failed to write bloom filter to disk: %w", err)
	}
	return nil
}
