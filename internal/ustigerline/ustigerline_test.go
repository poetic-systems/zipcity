package ustigerline_test

import (
	"fmt"
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

func TestReadFeaturesAndEdges(t *testing.T) {
	fileprefix := "tl_2025_01001"

	fmt.Printf("Reading edges and feature names for %s\n", fileprefix)
	err := ustigerline.ReadFeaturesAndEdges(
		fileprefix,
		func(info *ustigerline.StreetInfo) error {
			fmt.Printf("\nID: %s Name: %s\n\tAliases: %s\n", info.TLID, info.Name, info.Alt)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadFeaturesAndEdges(): %v", err)
	}
}

func TestReadFacesAndPlaces(t *testing.T) {
	fileprefix := "tl_2025_01001"

	fmt.Printf("Reading edges and feature names for %s\n", fileprefix)
	err := ustigerline.ReadFacesAndPlaces(
		fileprefix,
		func(info *ustigerline.CityInfo) error {
			fmt.Printf("\nPlaceID: %s Name: %s\n", info.PlaceFP, info.Name)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadFeaturesAndEdges(): %v", err)
	}
}

func TestReadAddressRanges(t *testing.T) {
	fileprefix := "tl_2025_01001"

	fmt.Printf("Reading edges and feature names for %s\n", fileprefix)
	err := ustigerline.ReadAddressRanges(
		fileprefix,
		func(info *ustigerline.AddressRange) error {
			fmt.Printf("\nID: %s Street: %s Side: %s\nZip: %s City: %s\n", info.TLID, info.Street.Name, info.Side, info.Zip, info.City.Name)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("Error from ustigerline.ReadFeaturesAndEdges(): %v", err)
	}
}
