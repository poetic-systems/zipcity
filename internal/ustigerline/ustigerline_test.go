package ustigerline_test

import (
	"fmt"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
	"github.com/twpayne/go-geom"
)

func TestReadFeaturesAndEdges(t *testing.T) {
	fileprefix := "../../data/us_census_tiger/tl_2025_01001"

	fmt.Printf("Reading edges and feature names for %s\n", fileprefix)
	err := ustigerline.ReadFeaturesAndEdges(
		fileprefix,
		func(id string, attributes map[string]any, aliases []string, geometry geom.T) error {
			fmt.Printf("\nID: %s Name: %s\n\tAliases: %s\n", id, attributes["FULLNAME"], aliases)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadZip(): %v", err)
	}
}
