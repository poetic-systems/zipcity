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
var stateFTPPath = "STATE/"
var storagedir = filepath.Join(strings.Split("./data/us_census_tiger/", "/")...)

var zipfiles = regexp.MustCompile(`tl_\d+_(\d+|us)_\w+\.zip`)

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

type StateInfo struct {
	Name    string
	USPS    string
	StateFP string
}

type AddressRangeFunc func(info *AddressRange) error

// AddressRange is one address range record as the file gives it, minus the
// house numbers we do not read. TIGER records a range per run of house numbers
// on each side of an edge, so many of them share a TLID and many repeat a side
// and its ZIP Code. They are reported one for one, in file order; deciding what
// a side's ZIP Codes are from them is the caller's business.
type AddressRange struct {
	// FromHouseNum string  // From Address range itself in Addr file
	// ToHouseNum string    // From Address range itself in Addr file
	TLID string
	Zip  string // From Address range itself in Addr file
	Side string // "L" or "R"
}

type StreetSide struct {
	// FromHouseNum string  // From Address range itself in Addr file
	// ToHouseNum string    // From Address range itself in Addr file
	TLID string
	// Zips holds every ZIP Code the address ranges give this side of the
	// edge. Addresses at each end of a street may be served by different ZIP
	// Codes, so a side has no single one.
	Zips   []string    // From Address ranges themselves in Addr file
	Side   string      // "L" or "R"
	Street *StreetInfo // From Edge and Features via the TLID
	City   *CityInfo   // From PlaceFP associated with TFID{Side} in Faces to Place file via edges
}

func ReadStates(fileprefix string) (map[string]*StateInfo, error) {
	statesDbfPath := filepath.Join(storagedir, "state", fmt.Sprintf("%s_us_state.zip", fileprefix))

	usstates, err := shapefile.ReadZipFile(statesDbfPath, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not a valid zip file") {
			os.Remove(statesDbfPath)
		}
		return nil, fmt.Errorf("unable to read %s: %w", statesDbfPath, err)
	}

	stateMap := make(map[string]*StateInfo)
	for usst := range usstates.Records() {
		stfips := asString(usst["STATEFP"])
		stname := asString(usst["NAME"])
		stusps := asString(usst["STUSPS"])

		if len(stfips) > 0 {
			// fmt.Printf("State: %s USPS: %s StateFP: %s\n", stname, stusps, stfips)
			stateMap[stfips] = &StateInfo{
				Name:    strings.ToUpper(stname),
				USPS:    stusps,
				StateFP: stfips,
			}
		}
	}

	return stateMap, nil
}

func tlidSideKey(tlid, side string) string {
	return fmt.Sprintf("%s-%s", tlid, side)
}

func ReadStreetSides(fileprefix string) (map[string]*StreetSide, error) {
	allSides := make(map[string]*StreetSide)

	tfidMap := make(map[string][]string)
	// get the street from edges and features
	// and set up lookup via faces
	err := ReadFeaturesAndEdges(fileprefix, func(
		stInfo *StreetInfo,
	) error {
		rTfid := asString(stInfo.Attributes["TFIDR"])
		if len(rTfid) > 0 {
			side := &StreetSide{
				TLID:   stInfo.TLID,
				Street: stInfo,
				Side:   "R",
			}
			k := tlidSideKey(stInfo.TLID, "R")
			allSides[k] = side
			tlidkeys, exists := tfidMap[rTfid]
			if !exists {
				tlidkeys = make([]string, 0)
			}
			tlidkeys = append(tlidkeys, k)
			tfidMap[rTfid] = tlidkeys
		}

		lTfid := asString(stInfo.Attributes["TFIDL"])
		if len(lTfid) > 0 {
			side := &StreetSide{
				TLID:   stInfo.TLID,
				Street: stInfo,
				Side:   "L",
			}
			k := tlidSideKey(stInfo.TLID, "L")
			allSides[k] = side
			tlidkeys, exists := tfidMap[lTfid]
			if !exists {
				tlidkeys = make([]string, 0)
			}
			tlidkeys = append(tlidkeys, k)
			tfidMap[lTfid] = tlidkeys
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get base edge data for sides: %w", err)
	}

	// get the city for the left and right faces
	err = ReadFacesAndPlaces(fileprefix, func(
		ctyInfo *CityInfo,
	) error {
		for _, tfid := range ctyInfo.TFID {
			tlidkeys := tfidMap[tfid]
			for _, sideTlidkey := range tlidkeys {
				side, ok := allSides[sideTlidkey]
				if ok {
					side.City = ctyInfo
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("unable to associate city data to side data: %s\n", err)
		// return nil, fmt.Errorf("unable to associate city data to side data: %w", err)
	}

	err = ReadAddressRanges(fileprefix, func(
		addrInfo *AddressRange,
	) error {
		// A range carrying no ZIP Code says nothing about the side it
		// describes, and one that names neither side of the road belongs to no
		// side we know of, so its key matches nothing here.
		if addrInfo.Zip == "" {
			return nil
		}
		side, ok := allSides[tlidSideKey(addrInfo.TLID, addrInfo.Side)]
		if !ok {
			return nil
		}
		// A side is described by an address range per run of house numbers,
		// and most of them repeat the side's ZIP Code, differing only in the
		// house numbers we do not keep. The ranges at each end of a street may
		// name different ZIP Codes, though, so a side has no single one and
		// every distinct one it names is kept, in the order the file gives
		// them.
		if slices.Contains(side.Zips, addrInfo.Zip) {
			return nil
		}
		side.Zips = append(side.Zips, addrInfo.Zip)
		return nil
	})
	if err != nil {
		fmt.Printf("unable to associate address range data to side data: %s\n", err)
		// return nil, fmt.Errorf("unable to associate address range data to side data: %w", err)
	}

	return allSides, nil
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
		tlid := asString(rawTLID)
		arSide := strings.ToUpper(asString(ar["SIDE"]))
		arZip := asString(ar["ZIP"])
		if arZip == "<nil>" {
			// The documentation above notes that a few address ranges carry no
			// ZIP Code at all. Report that as the absence it is.
			arZip = ""
		}

		err := addrFn(&AddressRange{TLID: tlid, Side: arSide, Zip: arZip})
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
		tfid := asString(rawTFID)
		rawPlaceFP, _ := facefields["PLACEFP"]
		placefp := asString(rawPlaceFP)

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
		placeFP := asString(rawPlaceFP)
		ctyInfo, found := placefpMap[placeFP]
		if !found {
			// fmt.Printf("No city info found for '%s'\n", placeFP)
			continue
		}
		ctyInfo.Name = strings.ToUpper(fmt.Sprintf("%s", pl["NAME"]))
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
		tlid := asString(rawTLID)
		rawFullname, found := fields["FULLNAME"]
		if !found {
			continue
		}
		fullname := strings.ToUpper(asString(rawFullname))

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
		edgeTLID := asString(rawTLID)
		stInfo, found := featnameIndex[edgeTLID]
		if !found {
			continue
		}
		rawFullname, found := attributes["FULLNAME"]
		if found {
			fullname := strings.ToUpper(asString(rawFullname))
			stInfo.Name = fullname
		}

		stInfo.Attributes = attributes
		stInfo.Geo = geometry

		err := shapeFn(stInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadAllRequiredTigerfiles() ([]string, AbsentSources, error) {
	return DownloadRequiredTigerfiles(allRequiredTigerfiles())
}

// allRequiredTigerfiles is every set of TIGER files a generation reads, and
// the granularity each is published at.
func allRequiredTigerfiles() []RequiredTigerfiles {
	return []RequiredTigerfiles{
		{ftpbase, featureFTPPath, "_featnames", "county"},
		{ftpbase, edgeFTPPath, "_edges", "county"},
		{ftpbase, addrFTPPath, "_addr", "addr"},
		{ftpbase, facesFTPPath, "_faces", "county"},
		{ftpbase, placeFTPPath, "_place", "state"},
		{ftpbase, stateFTPPath, "_state", "us"},
	}
}

// tigerfileIndex is what the Census Bureau's own FTP indexes say it publishes:
// every file listed, how many of each set's file types each area has, and
// which areas each file type covers.
type tigerfileIndex struct {
	files       []*url.URL
	counts      map[string]map[string]int
	setSizes    map[string]int
	areasByType map[string]map[string]bool
}

// readTigerfileIndexes reads the indexes and nothing else. Downloading the
// archives is the caller's business, which is what lets a test ask what the
// Census Bureau publishes today without fetching the whole release.
func readTigerfileIndexes(required []RequiredTigerfiles) (*tigerfileIndex, error) {
	idx := &tigerfileIndex{
		files:       make([]*url.URL, 0),
		counts:      make(map[string]map[string]int, 0),
		setSizes:    make(map[string]int, 0),
		areasByType: make(map[string]map[string]bool, len(required)),
	}
	for _, req := range required {
		idx.setSizes[req.Set] += 1
		cnt, ok := idx.counts[req.Set]
		if !ok {
			cnt = make(map[string]int, 0)
			idx.counts[req.Set] = cnt
		}

		filetype := strings.TrimPrefix(req.Suffix, "_")
		areas, ok := idx.areasByType[filetype]
		if !ok {
			areas = make(map[string]bool, 0)
			idx.areasByType[filetype] = areas
		}

		sourcefiles, err := downloadFtpIndex(req.Source.JoinPath(req.Path))
		if err != nil {
			return nil, err
		}
		idx.files = append(idx.files, sourcefiles...)

		for _, v := range sourcefiles {
			b := strings.TrimSuffix(path.Base(v.String()), ".zip")
			areas[areaOf(b, req.Suffix)] = true
			end := strings.LastIndex(b, "_")
			if end > 0 {
				b = b[0:end]
			}
			cnt[b] += 1
		}
	}

	return idx, nil
}

func DownloadRequiredTigerfiles(required []RequiredTigerfiles) ([]string, AbsentSources, error) {
	idx, err := readTigerfileIndexes(required)
	if err != nil {
		return nil, nil, err
	}
	counts, setSizes, allfiles := idx.counts, idx.setSizes, idx.files

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
			return nil, nil, fmt.Errorf("Mismatched required TIGER file names; error generating report: %w", err)
		}
		return nil, nil, fmt.Errorf("Mismatched required TIGER file names; report: %s, %v, %v, %d, %d", formatted, setSizes, counts, len(counts["county"]), len(mismatched))
	}

	err = os.MkdirAll(storagedir, 0755)
	if err != nil {
		return nil, nil, err
	}

	dir, err := os.Open(storagedir)
	if err != nil {
		return nil, nil, err
	}
	defer dir.Close()

	errorURLs := make(map[*url.URL]error)
	for _, rtf := range allfiles {
		err := downloadTigerfileZip(rtf, dir)
		if err != nil {
			errorURLs[rtf] = err
		}
	}

	// Every file the index listed must be here and must be readable. A file
	// the Census Bureau does not publish never reaches this list, so anything
	// missing or corrupt now is a failure to load data we have, not a gap in
	// what is published. Generating past it would bake that gap into the
	// filters, where an absent key is indistinguishable from a known-absent
	// one. See poetic-systems/zipcity#2.
	failures := make([]string, 0)
	for _, v := range allfiles {
		localpath, err := localTigerfilePath(v, dir)
		if err != nil {
			return nil, nil, err
		}
		if urlErr, hasError := errorURLs[v]; hasError {
			failures = append(failures, fmt.Sprintf("%s: %s", v, urlErr))
			continue
		}
		err = verifyTigerfileZip(localpath)
		if err != nil {
			// Drop it so the next run downloads it again, the same way a
			// download that arrives corrupt is dropped.
			os.Remove(localpath)
			failures = append(failures, fmt.Sprintf("%s: %s", localpath, err))
		}
	}
	if len(failures) > 0 {
		slices.Sort(failures)
		return nil, nil, fmt.Errorf(
			"%d of %d required TIGER files could not be loaded (removed any that were unreadable; re-run to download them again):\n  %s",
			len(failures), len(allfiles), strings.Join(failures, "\n  "),
		)
	}

	return slices.Collect(maps.Keys(counts["county"])), absentSources(idx.areasByType), nil
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

// verifyTigerfileZip reports whether a cached archive is one we can read. An
// empty archive is treated as unreadable: the Census Bureau sometimes answers a
// file request with a 200 and an error page, and a run that reads it as zero
// records looks exactly like a county with nothing in it.
func verifyTigerfileZip(localpath string) error {
	reader, err := zip.OpenReader(localpath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("archive holds no files")
	}
	return nil
}

func asString(input interface{}) string {
	s, ok := input.(string)
	if ok {
		// fmt.Printf("Formatting %v as string\n", input)
		return fmt.Sprintf("%s", s)
	}

	d, ok := input.(int)
	if ok {
		// fmt.Printf("Formatting %v as int\n", input)
		return fmt.Sprintf("%d", d)
	}

	// fmt.Printf("Formatting %v using Sprintf(v)\n", input)
	return fmt.Sprintf("%v", input)
}
