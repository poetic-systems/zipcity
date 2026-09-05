// Package areazip decides which ZIP Code a county equivalent may contribute to
// the sides of its streets that have none of their own.
//
// TIGER gives a side a ZIP Code by way of its address ranges, and for some
// areas it publishes no address ranges at all: there is no ADDR file for
// American Samoa, nor for Rota, Tinian or the Northern Islands, and the file
// published for Saipan describes 5 of that island's 6,458 sides. Where
// GeoNames names exactly one ZIP Code for the whole area, every address in it
// carries that ZIP Code, and saying so is the only thing that puts those
// streets into the filters at all. See poetic-systems/zipcity#7.
//
// The rules live here rather than in the generator so they can be tested
// without running a generation over the whole TIGER release.
package areazip

import "github.com/poetic-systems/zipcity/internal/ustigerline"

// Sole reports the one ZIP Code GeoNames names for a county equivalent, or the
// one it names for the whole state or territory where the county has none of
// its own. An area with more than one reports the empty string.
//
// More than one ZIP Code is nothing rather than a choice: knowing that a
// street is somewhere in Guam does not say which of Guam's twenty one ZIP
// Codes serves it, and a wrong ZIP Code in the filters is worse than a missing
// one.
func Sole(stateZips, countyZips map[string][]string, usps, countyfips string) string {
	for _, zips := range [][]string{countyZips[usps+countyfips], stateZips[usps]} {
		if len(zips) == 1 {
			return zips[0]
		}
	}

	return ""
}

// Contradicted reports whether any address range in the area names a ZIP Code
// other than the one given, in which case GeoNames is telling us less than the
// whole truth about the area and none of it may be contributed.
//
// Any single dissenting side is enough. That is deliberately the strict
// reading: GeoNames names one ZIP Code for most Puerto Rico municipios and the
// ranges show many, and 119 of the 141 areas GeoNames gives a single ZIP Code
// are refused this way. It also means one bad row in a published file can cost
// an area the whole contribution — see poetic-systems/zipcity#16, where a
// single Haines Borough range carries a Colorado ZIP Code. An absent key costs
// a caller a "maybe not" on an address that exists; a wrong key costs them a
// "maybe" on one that does not.
//
// A side that names no ZIP Code says nothing and does not contradict.
func Contradicted(sides map[string]*ustigerline.StreetSide, areazip string) bool {
	for _, side := range sides {
		for _, zip := range side.Zips {
			if zip != areazip {
				return true
			}
		}
	}

	return false
}
