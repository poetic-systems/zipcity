package ustigerline_test

import (
	"fmt"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
	"github.com/twpayne/go-geom"
)

func TestReadFeaturesAndEdges(t *testing.T) {
	fileprefix := "tl_2025_01001"

	fmt.Printf("Reading edges and feature names for %s\n", fileprefix)
	err := ustigerline.ReadFeaturesAndEdges(
		fileprefix,
		func(info *ustigerline.StreetInfo, attributes map[string]any, geometry geom.T) error {
			fmt.Printf("\nID: %s Name: %s\n\tAliases: %s\n", info.TLID, info.Name, info.Alt)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadFeaturesAndEdges(): %v", err)
	}
}
