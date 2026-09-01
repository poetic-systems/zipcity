//go:build ignore
// +build ignore

// generator.go - Run via 'go run internal/bloomgenerator/bloomgenerator.go' to generate filters.go

package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
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
	// compiled filter directory is now about 40 MB.

	now := time.Now()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Data Ingestion Loop
	// loop through the parsed US Census Bureau TIGER files here
	zipStreetData := map[string]map[string]ZipStreetTuple{}
	zipCityData := map[string]ZipCityTuple{}
	cityStreetData := map[string]CityStreetTuple{}
	streetOnlyData := map[string]*ustigerline.StreetSide{}

	// Cache the census data locally if we don't already have it
	prefixes, err := ustigerline.DownloadAllRequiredTigerfiles()
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

	numZip2City := uint(0)
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
		stateInfo := stateMap[statefips]

		for _, side := range allSides {
			cty := ""
			if side.City != nil {
				cty = side.City.Name
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
			for _, zip := range side.Zips {
				if len(cty) > 0 && len(zip) > 4 {
					key := bloomkeys.KeyZipCity(zip, cty)
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
			if len(cty) > 0 && len(street) > 0 {
				// make sure we include the primary name as well as the alternative names
				streetnames := append(side.Street.Alt, street)
				for _, stname := range streetnames {
					key := bloomkeys.KeyCityStateStreet(cty, stateInfo.USPS, stname)
					_, found := cityStreetData[key]
					if !found {
						cityStreetData[key] = CityStreetTuple{
							City:   cty,
							State:  stateInfo.USPS,
							Street: stname,
						}
						numCity2Street += 1
					}
				}
			}
			if len(street) > 0 && len(cty) == 0 && len(side.Zips) == 0 {
				streetOnlyData[street] = side
				numStreetOnly += 1
			}
		}
	}
	fmt.Printf("Counts - zip-city: %d zip-street: %d city-street: %d street-only: %d\n", numZip2City, numZip2Sreet, numCity2Street, numStreetOnly)

	// Initialize Bloom Filters
	// Estimates for US: ~30M unique combinations. FPR: 0.1% (0.001)
	// With several address range files absent (mostly for islands):
	//  Counts - zip-city: 4396668 zip-street: 20604757 city-street: 4396668

	// NOTE: We don't attempt to compress the bloom filter bytes because a
	// bloom filter should have a relatively random distribution of bits set
	// if its hashing algorithm is working properly.

	zipstreetfiles := make(map[string]string, 0)
	for zipscope, streets := range zipStreetData {
		numThisZip2Sreet := uint(len(streets))
		// Add ~1/8 of overhead to the count for the base capacity
		nZS := numThisZip2Sreet + (numThisZip2Sreet >> 3)
		streetFilter := bloom.NewWithEstimates(nZS, 0.01)

		for key := range streets {
			streetFilter.Add([]byte(key))
		}

		// Serialize the scoped Zip to Street Bloom Filter
		zsVarName := fmt.Sprintf("ZipStreet%s", zipscope)
		zsidentifier := fmt.Sprintf("%s-%s", "zip-street", zipscope)
		zipstreetfiles[zsVarName] = zsidentifier
		err = writeAsGob(
			path.Join(
				cwd,
				"generated",
				"compiled_filter",
				fmt.Sprintf("%s.bin", zsidentifier),
			),
			streetFilter,
		)
		if err != nil {
			panic(err)
		}
	}

	// Add ~1/8 of overhead to the count for the base capacity
	nZC := numZip2City + (numZip2City >> 3)
	cityFilter := bloom.NewWithEstimates(nZC, 0.01)

	for key := range zipCityData {
		cityFilter.Add([]byte(key))
	}

	// Serialize the Zip to City Bloom Filter
	err = writeAsGob(
		path.Join(
			cwd,
			"generated",
			"compiled_filter",
			fmt.Sprintf("%s.bin", "zip-city"),
		),
		cityFilter,
	)
	if err != nil {
		panic(err)
	}

	// Add ~1/8 of overhead to the count for the base capacity
	nCS := numCity2Street + (numCity2Street >> 3)
	cityStreetFilter := bloom.NewWithEstimates(nCS, 0.01)

	for key := range cityStreetData {
		cityStreetFilter.Add([]byte(key))
	}

	// Serialize the City to Street Bloom Filter
	err = writeAsGob(
		path.Join(
			cwd,
			"generated",
			"compiled_filter",
			fmt.Sprintf("%s.bin", "city-street"),
		),
		cityStreetFilter,
	)
	if err != nil {
		panic(err)
	}

	// Generate the Go source code containing the embedded asset
	tmpl := `// DO NOT EDIT! Code generated at {{ .Now }} by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"fmt"
	"regexp"
	bloom "github.com/bits-and-blooms/bloom/v3"
)

var zip5pattern = regexp.MustCompile({{ tick }}^\d{5}${{ tick }})

type CompiledFilter string

const (
	Unrecognized  CompiledFilter = ""
  ZipCity       CompiledFilter = "zip-city"
  CityStreet    CompiledFilter = "city-street"
{{- range $varName, $zsidentifier := .Files }}
	{{ $varName }}   CompiledFilter = "{{- $zsidentifier -}}"
{{- end }}
)

func ZipStreetFilterForZip(zip string) (CompiledFilter, error) {
	if !zip5pattern.MatchString(zip) {
		return Unrecognized, fmt.Errorf("5-digit zip code required")
	}
	zip2 := zip[0:2]
	filterid := fmt.Sprintf("zip-street-%s", zip2)
	switch filterid {
{{- range $varName, $zsidentifier := .Files }}
	case "{{- $zsidentifier -}}":
		return {{ $varName }}, nil
{{- end }}
	}
	return Unrecognized, fmt.Errorf("5-digit zip code required")
}

// LoadFilter restores the compiled filter in memory
func LoadFilter(name CompiledFilter) (*bloom.BloomFilter, error) {
	var filter bloom.BloomFilter
	var buf *bytes.Buffer
	switch name {
	case ZipCity:
		buf = bytes.NewBuffer(RawZipCityFilterBytes)
	case CityStreet:
		buf = bytes.NewBuffer(RawCityStreetFilterBytes)
	{{- range $varName, $zsidentifier := .Files }}
	case {{ $varName -}}:
		buf = bytes.NewBuffer(Raw{{- $varName -}}FilterBytes)
	{{- end }}
	default:
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

{{ range $varName, $zsidentifier := .Files }}
//go:embed {{ $zsidentifier -}}.bin
var Raw{{ $varName }}FilterBytes []byte    
{{ end }}

// RawZipCityFilterBytes holds the pre-compiled zip-city Bloom filter
//go:embed zip-city.bin
var RawZipCityFilterBytes []byte

// RawCityStreetFilterBytes holds the pre-compiled city-street Bloom filter
//go:embed city-street.bin
var RawCityStreetFilterBytes []byte

`
	templateFuncMap := template.FuncMap{
		"tick": func() string { return "`" },
	}
	t := template.Must(template.New("filter").Funcs(templateFuncMap).Parse(tmpl))

	err = os.MkdirAll("./generated/compiled_filter", 0755)
	if err != nil {
		panic(err)
	}
	outFile, err := os.Create("./generated/compiled_filter/compiled_filters.go")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	err = t.Execute(outFile, map[string]interface{}{
		"Files": zipstreetfiles,
		"Now":   now.UTC().Format(time.RFC3339),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully generated %s\n", outFile.Name())
}

func writeAsGob(filename string, data *bloom.BloomFilter) error {
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

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}
