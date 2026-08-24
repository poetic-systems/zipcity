//go:build ignore
// +build ignore

// generator.go - Run via 'go run internal/bloomgenerator/bloomgenerator.go' to generate filters.go

package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"text/template"
	"time"

	bloom "github.com/bits-and-blooms/bloom/v3"
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
	zipStreetData := map[string][]ZipStreetTuple{}
	zipCityData := []ZipCityTuple{}
	cityStreetData := []CityStreetTuple{}

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

	for _, pre := range prefixes {
		err := ustigerline.ReadAddressRanges(
			pre,
			func(info *ustigerline.AddressRange) error {
				// out, _ := json.MarshalIndent(info, "", "  ")
				// fmt.Printf("Address Range: %s\n", out)
				statefips := pre[len(pre)-5 : len(pre)-3]
				stateInfo := stateMap[statefips]

				if info.City != nil {
					zipCityData = append(zipCityData, ZipCityTuple{
						Zip:  info.Zip,
						City: info.City.Name,
					})
				}
				if len(info.Zip) > 4 && info.Street != nil {
					// NOTE: We store and load bloom filters on a per zip code prefix basis.
					zipscope := info.Zip[0:2]
					scoped, exists := zipStreetData[zipscope]
					if !exists {
						scoped = make([]ZipStreetTuple, 0)
					}
					scoped = append(scoped, ZipStreetTuple{
						Zip:    info.Zip,
						Street: info.Street.Name,
					})
					zipStreetData[zipscope] = scoped
				}
				if info.City != nil && info.Street != nil {
					cityStreetData = append(cityStreetData, CityStreetTuple{
						City:   info.City.Name,
						State:  stateInfo.USPS,
						Street: info.Street.Name,
					})
				}

				return nil
			},
		)
		if err != nil {
			fmt.Printf("Error from ustigerline.ReadAddressRanges(): %s\n", err)
			// panic(fmt.Errorf("Error from ustigerline.ReadAddressRanges(): %w", err))
		}
	}
	numZip2City := uint(len(zipCityData))
	numZip2Sreet := uint(len(zipStreetData))
	numCity2Street := uint(len(cityStreetData))
	fmt.Printf("Counts - zip-city: %d zip-street: %d city-street: %d\n", numZip2City, numZip2Sreet, numCity2Street)

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

		for _, record := range streets {
			// Generate the lookup key
			key := fmt.Sprintf("%s:%s", record.Zip, record.Street)
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

	for _, record := range zipCityData {
		// Generate the lookup key
		key := fmt.Sprintf("%s:%s", record.Zip, record.City)
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

	for _, record := range cityStreetData {
		// Generate the lookup key
		key := fmt.Sprintf("%s:%s:%s", record.City, record.State, record.Street)
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
	bloom "github.com/bits-and-blooms/bloom/v3"
)

type CompiledFilter string

const (
  ZipCity     CompiledFilter = "zip-city"
  CityStreet  CompiledFilter = "city-street"
{{- range $varName, $zsidentifier := .Files }}
	{{ $varName }} CompiledFilter = "{{- $zsidentifier -}}"
{{ end }}
)

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
		return nil, fmt.Errorf("Unsupported compiled filter name: %s", name)
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

	t := template.Must(template.New("filter").Parse(tmpl))

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
