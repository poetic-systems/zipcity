package ustigerline

import (
	"github.com/jonas-p/go-shp"
)

type ShapeFunc func(shape shp.Shape, row int, fields []shp.Field, reader *shp.ZipReader) error

func ReadZip(filename string, shapeFn ShapeFunc) error {
	// open a shapefile for reading
	shape, err := shp.OpenZip(filename)
	if err != nil {
		return err
	}
	defer shape.Close()

	// fields from the attribute table (DBF)
	fields := shape.Fields()

	// loop through all features in the shapefile
	for shape.Next() {
		n, p := shape.Shape()
		shapeFn(p, n, fields, shape)

		/*
			// print feature
			fmt.Println(reflect.TypeOf(p).Elem(), p.BBox())

			// print attributes
			for k, f := range fields {
				val := shape.ReadAttribute(n, k)
				fmt.Printf("\t%v: %v\n", f, val)
			}
			fmt.Println()
		*/
	}
	return nil
}
