package ustigerline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-shapefile"
)

var ftpbase *url.URL
var featureFTPPath = "FEATNAMES/"
var edgeFTPPath = "EDGES/"
var addrFTPPath = "ADDR/"
var facesFTPPath = "FACES/"
var placeFTPPath = "PLACE/"
var storagedir = "../../data/us_census_tiger/"

var zipfiles = regexp.MustCompile(`tl_\d+_\d+_\w+\.zip`)

func init() {
	var err error
	ftpbase, err = url.Parse("https://www2.census.gov/geo/tiger/TIGER2025/")
	if err != nil {
		panic(err)
	}
}

type RequiredTigerfiles struct {
	Source *url.URL
	Path   string
	Suffix string
	Set    string
}

// The All Lines shapefile (edges.shp) contains the geometry and attributes of each
// topological primitive edge. Each edge has a unique TLID (TIGER/Line Identifier) value.
// ...
// The Feature Names relationship table (featnames.dbf) contains a record for each feature
// name-edge combination, and includes the feature name attributes. The edge to which a
// Feature Names relationship table record applies can be determined by linking to the All
// Lines shapefile on the TLID attribute. Multiple Feature Names relationship table records
// can link to the same edge, for example, a road edge could link to US Hwy 22 and
// Rathburn Road. The linear feature to which the feature name applies is identified by the
// LINEARID attribute. Multiple feature names may exist for the same edge (linear
// features are not included in the data set, but could be constructed using the All Lines
// shapefile and the relationship tables).
// See https://www2.census.gov/geo/pdfs/maps-data/data/tiger/tgrshp2008/rel_file_desc_2008.pdf
//
// 5.1.1 ZIP Codes and Address Ranges
// The address numbers used to create address ranges are house number-street name style addresses (or
// city-style addresses). A house number-street name style address minimally consists of a structure
// number, street name, and a 5-digit ZIP Code (e.g., 213 Main Street 90210). In the 2025 TIGER/Line
// Shapefiles, ZIP Codes are only associated to address ranges.
// The ZIP Code is an attribute of the address ranges. The Address Ranges relationship file has a 5-digit
// ZIP Code field containing a numeric code that may have leading zeroes. Both sides of a street typically
// have the same ZIP Code, but this is not always true. Different ZIP Codes may serve different sides of a
// street or cover addresses at each end of a street. Nearly all address ranges will have a ZIP Code, but
// there are a few instances where unknown ZIP Codes result in null/blank values in the ZIP Code field.

type StreetFunc func(info *StreetInfo, attributes map[string]any, geometry geom.T) error

type StreetInfo struct {
	TLID     string
	Name     string
	Alt      []string
	ZipCodes []string // may not be populated
}

type CityFunc func(info *CityInfo, attributes map[string]any, geometry geom.T) error
type CityInfo struct {
	PlaceFP  string
	TFID     string
	Name     string
	ZipCodes []string // may not be populated
}

type AddressRangeFunc func(info *AddressRange, attributes map[string]any, geometry geom.T) error

type AddressRange struct {
	TLID string
	// FromHouseNum string  // From Address range itself in Addr file
	// ToHouseNum string    // From Address range itself in Addr file
	// Side bool   // 0: left, 1: right
	StreetName     string   // From FeatName associated with Edge
	AltStreetNames []string // From FeatName associated with Edge
	Zip            string   // From Address range itself in Addr file
	PlaceFP        string   // From PlaceFP associated with TFID{Side} in Faces to Place file
}

// func ReadAddressRanges(fileprefix string, shapeFn StreetFunc) error {
// 	addrDbfPath := fmt.Sprintf("%s%s_addr.zip", storagedir, fileprefix)
// 	facesDbfPath := fmt.Sprintf("%s%s_faces.zip", storagedir, fileprefix)
// 	// places files just use the state fipscode, not the county fips code
// 	stateprefix := fileprefix[0 : len(fileprefix)-3]
// 	placeDbfPath := fmt.Sprintf("%s%s_place.zip", storagedir, stateprefix)

// }

func ReadFeaturesAndEdges(fileprefix string, shapeFn StreetFunc) error {
	featnamesDbfPath := fmt.Sprintf("%s%s_featnames.zip", storagedir, fileprefix)
	edgesShpPath := fmt.Sprintf("%s%s_edges.zip", storagedir, fileprefix)

	featnameIndex := make(map[string]*StreetInfo)

	featnames, err := shapefile.ReadZipFile(featnamesDbfPath, nil)
	if err != nil {
		return err
	}

	for fields := range featnames.Records() {
		// TLID is the TIGER/Line ID. It is used to link the feature from the
		// featnames.zip to the edge from edges.zip. It is type int.
		// A featurename record should exist for every possible name of an edge.
		rawTLID, found := fields["TLID"]
		if !found {
			continue
		}
		tlid := fmt.Sprintf("%d", rawTLID)
		fullname, _ := fields["FULLNAME"].(string)

		if tlid != "" && fullname != "" {
			// build up the list of alternative names for this feature
			stInfo, ok := featnameIndex[tlid]
			if !ok {
				stInfo = &StreetInfo{
					TLID:     tlid,
					Name:     "",
					Alt:      make([]string, 0),
					ZipCodes: make([]string, 0),
				}
			}
			stInfo.Alt = append(stInfo.Alt, fullname)
			featnameIndex[tlid] = stInfo
			out, _ := json.MarshalIndent(fields, "", "  ")
			fmt.Printf("%s", out)
		}
	}

	// addr also has TLID to tie address ranges back to edges. Each address range has
	// a side of the road and a zip code, which we can use to loosely associate a zip
	// code with a face / city. Loosely, because we are not tying the zipcode to where
	// on the road the address range occurs (yet.)

	//
	edges, err := shapefile.ReadZipFile(edgesShpPath, nil)
	if err != nil {
		return err
	}

	for attributes, geometry := range edges.Records() {
		rawTLID, found := attributes["TLID"]
		if !found {
			continue
		}
		edgeLinearID := fmt.Sprintf("%d", rawTLID)
		stInfo, found := featnameIndex[edgeLinearID]
		if !found {
			continue
		}
		stInfo.Name = fmt.Sprintf("%s", attributes["FULLNAME"])

		err := shapeFn(stInfo, attributes, geometry)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadAllRequiredTigerfiles() ([]string, error) {
	return DownloadRequiredTigerfiles([]RequiredTigerfiles{
		{ftpbase, featureFTPPath, "_featnames", "county"},
		{ftpbase, edgeFTPPath, "_edges", "county"},
		{ftpbase, addrFTPPath, "_addr", "addr"},
		{ftpbase, facesFTPPath, "_faces", "county"},
		{ftpbase, placeFTPPath, "_place", "state"},
	})
}

func DownloadRequiredTigerfiles(required []RequiredTigerfiles) ([]string, error) {
	counts := make(map[string]map[string]int, 0)
	setSizes := make(map[string]int, 0)
	allfiles := make([]*url.URL, 0)
	for _, req := range required {
		source := req.Source
		setSizes[req.Set] += 1
		cnt, ok := counts[req.Set]
		if !ok {
			cnt = make(map[string]int, 0)
			counts[req.Set] = cnt
		}

		sourceindex := source.JoinPath(req.Path)
		sourcefiles, err := downloadFtpIndex(sourceindex)
		if err != nil {
			return nil, err
		}
		allfiles = append(allfiles, sourcefiles...)

		for _, v := range sourcefiles {
			b := path.Base(v.String())
			end := strings.LastIndex(b, "_")
			if end > 0 {
				b = b[0:end]
			}
			cnt[b] += 1
		}
	}

	mismatched := slices.Collect(func(yield func(string) bool) {
		for setkey, set := range counts {
			for url, c := range set {
				if c < setSizes[setkey] {
					if !yield(url) {
						return
					}
				}
			}
		}
	})
	if len(mismatched) > 0 {
		formatted, err := json.MarshalIndent(mismatched, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("Mismatched required TIGER file names; error generating report: %w", err)
		}
		return nil, fmt.Errorf("Mismatched required TIGER file names; report: %s, %v, %v, %d, %d", formatted, setSizes, counts, len(counts["county"]), len(mismatched))
	}

	err := os.MkdirAll(storagedir, 0755)
	if err != nil {
		return nil, err
	}

	dir, err := os.Open(storagedir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	for _, rtf := range allfiles {
		fmt.Printf("Downloading %s\n", rtf.String())
		err := downloadTigerfileZip(rtf, dir)
		if err != nil {
			return nil, err
		}
	}

	// make sure all expected files exist
	for _, v := range allfiles {
		filename := path.Base(v.String())
		filepath := filepath.Join(dir.Name(), filename)
		_, err := os.Stat(filepath)
		if err != nil || errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s may not exist: %w", filename, err)
		}
	}

	return slices.Collect(maps.Keys(counts["county"])), nil
}

func downloadFtpIndex(indexurl *url.URL) ([]*url.URL, error) {
	tigerfiles := make(map[string]*url.URL, 0)

	resp, err := http.Get(indexurl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body: %w", err)
	}
	bodyString := string(bodyBytes)

	tigerfilelike := zipfiles.FindAllString(bodyString, -1)
	for _, candidate := range tigerfilelike {
		_, found := tigerfiles[candidate]
		if !found {
			tfurl := indexurl.JoinPath(candidate)
			tigerfiles[candidate] = tfurl
		}
	}

	return slices.Collect(func(yield func(u *url.URL) bool) {
		for _, v := range tigerfiles {
			if !yield(v) {
				return
			}
		}
	}), nil
}

func downloadTigerfileZip(fileurl *url.URL, dir *os.File) error {
	filename := path.Base(fileurl.Path)
	filepath := filepath.Join(dir.Name(), filename)

	// open the local file for writing first and only request
	// the file via http if it doesn't exist.
	out, err := os.OpenFile(filepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	defer out.Close()

	resp, err := http.Get(fileurl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
