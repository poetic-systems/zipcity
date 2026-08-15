package ustigerline

import (
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-shapefile"
)

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
	featnamesDbfPath := fmt.Sprintf("%s_featnames.zip", fileprefix)
	edgesShpPath := fmt.Sprintf("%s_edges.zip", fileprefix)

	featnameIndex := make(map[string][]string)
	featnameIndex2 := make(map[string][]string)

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
			matchedNames, found = featnameIndex2[edgeLinearID]
			if !found {
				continue
			}
		}

		err := shapeFn(edgeLinearID, attributes, matchedNames, geometry)
		if err != nil {
			return err
		}
	}
	return nil
}
