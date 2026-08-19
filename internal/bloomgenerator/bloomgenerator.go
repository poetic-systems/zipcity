//go:build ignore
// +build ignore

// generator.go - Run via 'go run internal/bloomgenerator/bloomgenerator.go' to generate filters.go

package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
	"github.com/twpayne/go-geom"
)

// FIXME: this code is copy pasta from a Google search / AI Mode session, with minor
// changes. It needs a thorough rework to ensure that it works as expected.

type ZipStreetTuple struct {
	Zip    string
	Street string
}

func main() {
	// TODO: See if we actually need all of the streets of if we
	// can get away with just encoding the challenging ones: the ones
	// that start with directionals, those that are numeric or
	// alphabetic and might need to be spelled out or not, etc.
	// NOTE: the current generated template file is 100MB without any
	// data in the bloom filter - just zeros. We also have not accounted
	// for the city, region, zip lookup. We'll need to do better than that.

	now := time.Now()

	// Initialize Bloom Filters
	// Estimates for US: ~30M unique combinations. FPR: 0.1% (0.001)
	// streetFilter := bloom.NewWithEstimates(30000000, 0.001)
	streetFilter := bloom.NewWithEstimates(50000, 0.001)

	// Data Ingestion Loop
	// loop through the parsed US Census Bureau TIGER files here
	zipStreetData := []ZipStreetTuple{}

	// Cache the census data locally if we don't already have it
	prefixes, err := ustigerline.DownloadAllRequiredTigerfiles()
	if err != nil {
		panic(err)
	}

	for i, pre := range prefixes {
		err := ustigerline.ReadFeaturesAndEdges(
			pre,
			func(info *ustigerline.StreetInfo, attributes map[string]any, geometry geom.T) error {
				out, _ := json.MarshalIndent(attributes, "", "  ")
				fmt.Printf("\nName: %s\nAliases: %s\nAttributes: %s", info.Name, info.Alt, out)

				// FIXME: this zip code data won't work. It isn't set for a large
				// number of streets. We might end up doing a per state lookup instead.
				zips := make([]string, 0)
				zl := attributes["ZIPL"].(string)
				if len(zl) > 0 {
					zips = append(zips, zl)
				}

				zr := attributes["ZIPR"].(string)
				if len(zr) > 0 {
					zips = append(zips, zr)
				}

				for _, a := range info.Alt {
					for _, z := range zips {
						zipStreetData = append(zipStreetData, ZipStreetTuple{
							Zip:    z,
							Street: a,
						})
					}
				}

				return nil
			},
		)
		if err != nil {
			panic(fmt.Errorf("Error from ustigerline.ReadFeaturesAndEdges(): %w", err))
		}
		if i > 0 {
			os.Exit(0)
		}
	}

	for _, record := range zipStreetData {
		// Generate the lookup key
		key := fmt.Sprintf("%s:%s", record.Zip, record.Street)
		streetFilter.Add([]byte(key))
	}

	// Serialize the Bloom Filter into raw bytes
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(streetFilter); err != nil {
		panic(err)
	}

	// Should we further compress the buffer bytes? It probably won't work
	// because a bloom filter should have a relatively random distribution
	// of bits set if its hashing algorithm is working properly.

	// Generate the Go File containing the embedded asset
	tmpl := `// DO NOT EDIT! Code generated at {{ .Now }} by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	"encoding/gob"
	bloom "github.com/bits-and-blooms/bloom/v3"
)

// LoadStreetFilter restores the compiled filter in memory
func LoadStreetFilter() (*bloom.BloomFilter, error) {
	var filter bloom.BloomFilter
	buf := bytes.NewBuffer(RawStreetFilterBytes)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

// RawStreetFilterBytes holds the pre-compiled Bloom filter
var RawStreetFilterBytes = []byte({
	{{range .Bytes}}{{.}},{{end}}
})

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

	// Pass the byte slice to the template
	err = t.Execute(outFile, map[string]interface{}{
		"Bytes": buf.Bytes(),
		"Now":   now.UTC().Format(time.RFC3339),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully generated %s\n", outFile.Name())
}
