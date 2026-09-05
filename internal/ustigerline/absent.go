package ustigerline

import (
	"slices"
	"strconv"
	"strings"
)

// AbsentSources names, for each area, the required TIGER file types the Census
// Bureau's own index does not list for it.
//
// Absent is not the same as failed. A file the index lists must be present and
// readable or generation stops, so a filter that exists was built from a
// complete read of everything published. What a caller cannot see from the
// filters is the other case: an area for which nothing was published, so there
// was never anything to fail at. That is what this records.
//
// In the 2025 release only ADDR is ever absent, and only for eight county
// equivalents — the five in American Samoa and Rota, Tinian and the Northern
// Islands. Those areas carry a ZIP Code inferred from GeoNames rather than one
// an address range gave us. Nothing here is derived from a hard-coded list;
// it is read off the index each generation, so a change at the Census Bureau
// shows up as a change here. See poetic-systems/zipcity#2 and #7.
//
// Keys are area codes as TIGER names them — a five digit county-equivalent
// FIPS code, a two digit state code, or "us" — and values are file types,
// lowercase and sorted.
type AbsentSources map[string][]string

// absentSources reports which of the areas seen in the index are missing which
// file types.
//
// areasByType maps a file type to the areas the index lists a file for. An
// area is compared only against types published at its own granularity: a
// county equivalent is not missing the national state file, and the boundary
// between the two is the shape of the area code rather than a list of which
// type is which.
func absentSources(areasByType map[string]map[string]bool) AbsentSources {
	areasByShape := make(map[string]map[string]bool)
	for _, areas := range areasByType {
		for area := range areas {
			g := shapeOf(area)
			seen, ok := areasByShape[g]
			if !ok {
				seen = make(map[string]bool)
				areasByShape[g] = seen
			}
			seen[area] = true
		}
	}

	absent := make(AbsentSources)
	for filetype, areas := range areasByType {
		for peer := range areasByShape[shapeOfAny(areas)] {
			if !areas[peer] {
				absent[peer] = append(absent[peer], filetype)
			}
		}
	}

	for area, filetypes := range absent {
		slices.Sort(filetypes)
		absent[area] = slices.Compact(filetypes)
	}

	return absent
}

// shapeOf reports the kind of area an area code names. Codes of the same kind
// are the peers an absence is measured against: "5" for a county-equivalent
// FIPS code, "2" for a state code, and the code itself for a literal like "us"
// that names the whole country and so has no peers to be absent among.
func shapeOf(area string) string {
	for _, r := range area {
		if r < '0' || r > '9' {
			return area
		}
	}

	return strconv.Itoa(len(area))
}

// shapeOfAny reports the shape of any one of a file type's areas: every file
// of a type is published at the same granularity.
func shapeOfAny(areas map[string]bool) string {
	for area := range areas {
		return shapeOf(area)
	}

	return ""
}

// areaOf reports the area code a TIGER file name names: the last underscored
// segment of the name with the file type stripped off, so tl_2025_02100_addr
// is area 02100 and tl_2025_us_state is area us.
func areaOf(basename, suffix string) string {
	name := strings.TrimSuffix(basename, suffix)
	parts := strings.Split(name, "_")

	return parts[len(parts)-1]
}
