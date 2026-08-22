package ustigerline

import (
	"archive/zip"
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
var storagedir = filepath.Join(strings.Split("./data/us_census_tiger/", "/")...)

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

type StreetFunc func(info *StreetInfo) error

type StreetInfo struct {
	TLID       string
	Name       string
	Alt        []string
	Attributes map[string]any
	Geo        geom.T
}

type CityFunc func(info *CityInfo) error
type CityInfo struct {
	PlaceFP    string
	TFID       []string
	Name       string
	Attributes map[string]any
	Geo        geom.T
}

type AddressRangeFunc func(info *AddressRange) error

type AddressRange struct {
	// FromHouseNum string  // From Address range itself in Addr file
	// ToHouseNum string    // From Address range itself in Addr file
	TLID   string
	Zip    string      // From Address range itself in Addr file
	Side   string      // "L" or "R"
	Street *StreetInfo // From Edge and Features via the TLID
	City   *CityInfo   // From PlaceFP associated with TFID{Side} in Faces to Place file via edges
}

func ReadAddressRanges(fileprefix string, addrFn AddressRangeFunc) error {
	addrDbfPath := filepath.Join(storagedir, "addr", fmt.Sprintf("%s_addr.zip", fileprefix))

	addressranges, err := shapefile.ReadZipFile(addrDbfPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(addrDbfPath)
		}
		return fmt.Errorf("unable to read %s: %w", addrDbfPath, err)
	}

	// Multiple Address Ranges map to the same TLID (edge feature)
	addrMap := make(map[string]*AddressRange)

	// addressranges has TLID to tie address ranges back to edges. Each address range
	// also has a side of the road and a zip code, which we can use to loosely associate
	// a zip code with a face / city by associating the zip from the address range to the
	// TFID in the edge record for the side of the road the address range is on. Loosely,
	// because we are not tying the zipcode to where on the road the address range occurs
	// (yet.)
	for ar := range addressranges.Records() {
		// TLID is the TIGER/Line ID. It is used to link the address range from the
		// addr.zip to the edge from edges.zip. It is type int.
		rawTLID, found := ar["TLID"]
		if !found {
			continue
		}
		tlid := fmt.Sprintf("%v", rawTLID)
		rawSide, _ := ar["SIDE"]
		arSide := fmt.Sprintf("%s", rawSide)
		rawZip, _ := ar["ZIP"]
		arZip := fmt.Sprintf("%v", rawZip)

		if tlid != "" && arSide != "" {
			// build up the list of tfids for this place
			arInfo, ok := addrMap[tlid]
			if !ok {
				arInfo = &AddressRange{
					TLID: tlid,
					Zip:  "",
				}
			}
			arInfo.Zip = arZip
			if arSide == "L" || arSide == "R" {
				arInfo.Side = arSide
			}
			addrMap[tlid] = arInfo
		}
	}

	addrTfidMap := make(map[string]*AddressRange)
	// get the street from edges and features
	// and set up lookup via faces
	ReadFeaturesAndEdges(fileprefix, func(
		stInfo *StreetInfo,
	) error {
		arInfo, ok := addrMap[stInfo.TLID]
		if ok {
			arInfo.Street = stInfo
			rawTFID := stInfo.Attributes["TFIDR"]
			if arInfo.Side == "L" {
				rawTFID = stInfo.Attributes["TFIDL"]
			}
			tfid := fmt.Sprintf("%v", rawTFID)
			addrTfidMap[tfid] = arInfo
		}
		return nil
	})

	// get the city for the left anf right faces
	ReadFacesAndPlaces(fileprefix, func(
		ctyInfo *CityInfo,
	) error {
		for _, tfid := range ctyInfo.TFID {
			arInfo, ok := addrTfidMap[tfid]
			if ok {
				arInfo.City = ctyInfo
			}
		}
		return nil
	})

	for _, arInfo := range addrMap {
		err := addrFn(arInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadFacesAndPlaces(fileprefix string, cityFn CityFunc) error {
	facesDbfPath := filepath.Join(storagedir, "faces", fmt.Sprintf("%s_faces.zip", fileprefix))
	// places files just use the state fipscode, not the county fips code
	stateprefix := fileprefix[0 : len(fileprefix)-3]
	placeDbfPath := filepath.Join(storagedir, "place", fmt.Sprintf("%s_place.zip", stateprefix))

	// Many TFIDs map to the same PLACEFP (City)
	placefpMap := make(map[string]*CityInfo)

	// fmt.Printf("Reading %s\n", facesDbfPath)
	faces, err := shapefile.ReadZipFile(facesDbfPath, nil)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", facesDbfPath, err)
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(facesDbfPath)
		}
		return err
	}

	for facefields := range faces.Records() {
		// TFID is the TIGER/Face ID. It is used to link the "face" from the
		// faces.zip to the place (generally a city) from place.zip. It is type int.
		rawTFID, found := facefields["TFID"]
		if !found {
			out, _ := json.MarshalIndent(facefields, "", "  ")
			fmt.Printf("No TFID found in %s\n", out)
			continue
		}
		tfid := fmt.Sprintf("%v", rawTFID)
		rawPlaceFP, _ := facefields["PLACEFP"]
		placefp := fmt.Sprintf("%v", rawPlaceFP)

		// fmt.Printf("TFID: '%s' PLACEFP: '%s'\n", tfid, placefp)

		if tfid != "" && placefp != "" {
			// build up the list of tfids for this place
			ctyInfo, ok := placefpMap[placefp]
			if !ok {
				ctyInfo = &CityInfo{
					PlaceFP: placefp,
					TFID:    make([]string, 0),
					Name:    "",
				}
			}
			ctyInfo.TFID = append(ctyInfo.TFID, tfid)
			placefpMap[placefp] = ctyInfo
		}
	}

	places, err := shapefile.ReadZipFile(placeDbfPath, nil)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", placeDbfPath, err)
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(placeDbfPath)
		}
		return err
	}

	for pl, geometry := range places.Records() {
		rawPlaceFP, found := pl["PLACEFP"]
		if !found {
			// out, _ := json.MarshalIndent(pl, "", "  ")
			// fmt.Printf("No PLACEFP found in %s\n", out)
			continue
		}
		placeFP := fmt.Sprintf("%v", rawPlaceFP)
		ctyInfo, found := placefpMap[placeFP]
		if !found {
			// fmt.Printf("No city info found for '%s'\n", placeFP)
			continue
		}
		ctyInfo.Name = fmt.Sprintf("%s", pl["NAME"])
		ctyInfo.Attributes = pl
		ctyInfo.Geo = geometry

		// out, _ := json.MarshalIndent(ctyInfo, "", "  ")
		// fmt.Printf("%s\n", out)

		err := cityFn(ctyInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFeaturesAndEdges(fileprefix string, shapeFn StreetFunc) error {
	featnamesDbfPath := filepath.Join(storagedir, "featnames", fmt.Sprintf("%s_featnames.zip", fileprefix))
	edgesShpPath := filepath.Join(storagedir, "edges", fmt.Sprintf("%s_edges.zip", fileprefix))

	featnameIndex := make(map[string]*StreetInfo)

	featnames, err := shapefile.ReadZipFile(featnamesDbfPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(featnamesDbfPath)
		}
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
		tlid := fmt.Sprintf("%v", rawTLID)
		fullname, _ := fields["FULLNAME"].(string)

		if tlid != "" && fullname != "" {
			// build up the list of alternative names for this feature
			stInfo, ok := featnameIndex[tlid]
			if !ok {
				stInfo = &StreetInfo{
					TLID: tlid,
					Name: "",
					Alt:  make([]string, 0),
				}
			}
			stInfo.Alt = append(stInfo.Alt, fullname)
			featnameIndex[tlid] = stInfo
			// out, _ := json.MarshalIndent(fields, "", "  ")
			// fmt.Printf("%s\n", out)
		}
	}

	edges, err := shapefile.ReadZipFile(edgesShpPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(edgesShpPath)
		}
		return err
	}

	for attributes, geometry := range edges.Records() {
		rawTLID, found := attributes["TLID"]
		if !found {
			continue
		}
		edgeTLID := fmt.Sprintf("%v", rawTLID)
		stInfo, found := featnameIndex[edgeTLID]
		if !found {
			continue
		}
		stInfo.Name = fmt.Sprintf("%s", attributes["FULLNAME"])
		stInfo.Attributes = attributes
		stInfo.Geo = geometry

		err := shapeFn(stInfo)
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

	errorURLs := make(map[*url.URL]error)
	for _, rtf := range allfiles {
		err := downloadTigerfileZip(rtf, dir)
		if err != nil {
			errorURLs[rtf] = err
		}
	}

	// make sure all expected files exist
	for _, v := range allfiles {
		urlErr, hasError := errorURLs[v]
		if hasError {
			fmt.Printf("%s error: %s\n", v, urlErr)
			continue
		}
		localpath, err := localTigerfilePath(v, dir)
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(localpath)
		if err != nil || errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s may not exist: %w", localpath, err)
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

func localTigerfilePath(fileurl *url.URL, dir *os.File) (string, error) {
	filename := path.Base(fileurl.Path)
	before, found := strings.CutSuffix(filename, ".zip")
	if !found {
		return "", fmt.Errorf("%s does not have a .zip extension", filename)
	}
	parts := strings.Split(before, "_")
	return filepath.Join(dir.Name(), parts[len(parts)-1], filename), nil
}

func downloadTigerfileZip(fileurl *url.URL, dir *os.File) error {
	localpath, err := localTigerfilePath(fileurl, dir)
	if err != nil {
		return err
	}
	subdir := path.Dir(localpath)
	err = os.MkdirAll(subdir, 0755)
	if err != nil {
		return fmt.Errorf("Error creating TIGER file type subdirectory %s: %w", localpath, err)
	}

	// open the local file for writing first and only request
	// the file via http if it doesn't exist.
	out, err := os.OpenFile(localpath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	defer out.Close()

	fmt.Printf("Downloading %s to %s\n", fileurl.String(), localpath)
	resp, err := http.Get(fileurl.String())
	if err != nil {
		return err
	}
	/*
		Sometimes a response comes back with a 200 status code and this message
		instead of the zip file:
		"The requested URL was rejected. Please consult with your administrator.

		Your support ID is: 13427891559851952768"
	*/
	fmt.Println("Status Code:", resp.StatusCode)
	expectedSize := resp.ContentLength
	if expectedSize == -1 {
		fmt.Println("Warning: Server did not provide Content-Length header")
	}
	defer resp.Body.Close()

	bytesWritten, err := io.Copy(out, resp.Body)
	if expectedSize != -1 && bytesWritten != expectedSize {
		// Delete the bad file
		os.Remove(localpath)
		return fmt.Errorf("Incomplete download! Got %d of %d bytes\n", bytesWritten, expectedSize)
	}

	fmt.Printf("Downloaded all %d bytes of %s\n", bytesWritten, localpath)
	out.Close() // this will get called twice!
	// <-time.After(200 * time.Millisecond)

	reader, err := zip.OpenReader(localpath)
	if err != nil {
		os.Remove(localpath)
		return fmt.Errorf("Corrupt or incomplete zip: %w", err)
	}
	defer reader.Close()

	return err
}
