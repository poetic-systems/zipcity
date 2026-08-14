package ustigerline_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jonas-p/go-shp"
	"github.com/poetic-systems/zipcity/internal/ustigerline"
)

func TestReadZip(t *testing.T) {
	filename := "../../data/us_census_tiger/tl_2025_01001_edges.zip"
	fmt.Printf("Reading %s\n", filename)
	ustigerline.ReadZip(
		filename,
		func(shape shp.Shape, row int, fields []shp.Field, reader *shp.ZipReader) error {
			// print feature
			fmt.Println(reflect.TypeOf(shape).Elem(), shape.BBox())

			// print attributes
			for k, f := range fields {
				val := reader.Attribute(k)
				fmt.Printf("\t%v: %v\n", f, val)
			}
			fmt.Println()
			return nil
		},
	)
}
