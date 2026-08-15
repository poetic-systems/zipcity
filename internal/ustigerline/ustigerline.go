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
var storagedir = "../../data/us_census_tiger/"

var zipfiles = regexp.MustCompile(`tl_\d+_\d+_\w+\.zip`)

func init() {
	var err error
	ftpbase, err = url.Parse("https://www2.census.gov/geo/tiger/TIGER2025/")
	if err != nil {
		panic(err)
	}
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

type ShapeFunc func(id string, attributes map[string]any, aliases []string, geometry geom.T) error

func ReadFeaturesAndEdges(fileprefix string, shapeFn ShapeFunc) error {
	featnamesDbfPath := fmt.Sprintf("%s%s_featnames.zip", storagedir, fileprefix)
	edgesShpPath := fmt.Sprintf("%s%s_edges.zip", storagedir, fileprefix)

	featnameIndex := make(map[string][]string)

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
			featnameIndex[tlid] = append(featnameIndex[tlid], fullname)
			out, _ := json.MarshalIndent(fields, "", "  ")
			fmt.Printf("%s", out)
		}
	}

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
		matchedNames, found := featnameIndex[edgeLinearID]
		if !found {
			continue
		}

		err := shapeFn(edgeLinearID, attributes, matchedNames, geometry)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadFeaturesAndEdges() ([]string, error) {
	return DownloadFeaturesAndEdgesFrom(ftpbase)
}

func DownloadFeaturesAndEdgesFrom(source *url.URL) ([]string, error) {
	featureindex := source.JoinPath(featureFTPPath)
	featurefiles, err := downloadFtpIndex(featureindex)
	if err != nil {
		return nil, err
	}

	edgeindex := source.JoinPath(edgeFTPPath)
	edgefiles, err := downloadFtpIndex(edgeindex)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, 0)
	for _, v := range featurefiles {
		b := path.Base(v.String())
		end := strings.LastIndex(b, "_")
		if end > 0 {
			b = b[0:end]
		}
		counts[b] += 1
	}
	for _, v := range edgefiles {
		b := path.Base(v.String())
		end := strings.LastIndex(b, "_")
		if end > 0 {
			b = b[0:end]
		}
		counts[b] += 1
	}
	mismatched := slices.Collect(func(yield func(string) bool) {
		for url, c := range counts {
			if c < 2 {
				if !yield(url) {
					return
				}
			}
		}
	})
	if len(mismatched) > 0 {
		formatted, err := json.MarshalIndent(mismatched, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("Mismatched feature names and edges file names; error generating report: %w", err)
		}
		return nil, fmt.Errorf("Mismatched feature names and edges file names; report: %s", formatted)
	}

	err = os.MkdirAll(storagedir, 0755)
	if err != nil {
		return nil, err
	}

	dir, err := os.Open(storagedir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	for _, ff := range featurefiles {
		err := downloadTigerfileZip(ff, dir)
		if err != nil {
			return nil, err
		}
	}

	for _, ef := range edgefiles {
		err := downloadTigerfileZip(ef, dir)
		if err != nil {
			return nil, err
		}
	}

	// make sure all expected files exist
	allfiles := append(featurefiles, edgefiles...)
	for _, v := range allfiles {
		filename := path.Base(v.String())
		filepath := filepath.Join(dir.Name(), filename)
		out, err := os.OpenFile(filepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
		out.Close()
		if err == nil || !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("%s may not exist: %w", filename, err)
		}
	}

	return slices.Collect(maps.Keys(counts)), nil
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
