// DO NOT EDIT! Code generated at 2026-09-16T21:26:03Z by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	"embed"
	"fmt"
	"maps"
	"path"
	"regexp"
	"strings"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/poetic-systems/zipcity/internal/bloomfilename"
)

var zip5pattern = regexp.MustCompile(`^\d{5}$`)

// AbsentSources names, per TIGER area code, the source file types the Census
// Bureau published nothing of at generation time. Read off the Census Bureau's
// own index each generation rather than from a list kept here.
//
// A file the index does list must load or generation stops, so these are the
// only gaps in what the filters were built from. An area named here is one the
// filters know less about than the rest; absence from the filters is weaker
// evidence there than elsewhere. See poetic-systems/zipcity#2.
var AbsentSources = map[string][]string{
	"60010": {"addr"},
	"60020": {"addr"},
	"60030": {"addr"},
	"60040": {"addr"},
	"60050": {"addr"},
	"69085": {"addr"},
	"69100": {"addr"},
	"69120": {"addr"},
}

//go:embed bloom_filters/*.bin
var bloomDir embed.FS

// Keep a map of the raw, zero-allocation byte arrays
var bloom_filters = maps.Collect(func(yield func(CompiledFilter, *bloom.BloomFilter) bool) {
	for key, name := range allCompiledFilters {

		keypath := path.Join("bloom_filters", bloomfilename.Filename(key))

		// ReadFile allocates each byte slice once during boot
		data, err := bloomDir.ReadFile(keypath)
		if err != nil {
			panic(err)
		}

		filter := &bloom.BloomFilter{}
		reader := bytes.NewReader(data)
		_, err = filter.ReadFrom(reader)
		if err != nil {
			panic(fmt.Errorf("Failed to read %s bloom filter: %w", name, err))
		}

		if !yield(name, filter) {
			return
		}
	}
})

func toCompiledFilter(in string) (CompiledFilter, error) {
	cf, ok := allCompiledFilters[in]
	if ok {
		return cf, nil
	}
	return Unrecognized, fmt.Errorf("Unrecognized filter name")
}

func CityStreetFilterForState(state string) (CompiledFilter, error) {
	if len(state) != 2 {
		return Unrecognized, fmt.Errorf("USPS state abbreviation required")
	}

	filterid := fmt.Sprintf("city-street-%s", strings.ToUpper(state))
	cf, err := toCompiledFilter(filterid)
	if err != nil {
		return Unrecognized, fmt.Errorf("Supported USPS state abbreviation required")
	}
	return cf, nil
}

func ZipStreetFilterForZip(zip string) (CompiledFilter, error) {
	if !zip5pattern.MatchString(zip) {
		return Unrecognized, fmt.Errorf("5-digit zip code required")
	}
	zip2 := zip[0:2]
	filterid := fmt.Sprintf("zip-street-%s", zip2)
	cf, err := toCompiledFilter(filterid)
	if err != nil {
		return Unrecognized, fmt.Errorf("known 5-digit zip code required")
	}
	return cf, nil
}

// LoadFilter restores the compiled filter in memory
func LoadFilter(name CompiledFilter) (*bloom.BloomFilter, error) {
	filter, ok := bloom_filters[name]
	if !ok {
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	return filter, nil
}

type CompiledFilter string

const (
	Unrecognized CompiledFilter = ""
	ZipCity      CompiledFilter = "zip-city"
	CityStreetAK CompiledFilter = "city-street-AK"
	CityStreetAL CompiledFilter = "city-street-AL"
	CityStreetAR CompiledFilter = "city-street-AR"
	CityStreetAS CompiledFilter = "city-street-AS"
	CityStreetAZ CompiledFilter = "city-street-AZ"
	CityStreetCA CompiledFilter = "city-street-CA"
	CityStreetCO CompiledFilter = "city-street-CO"
	CityStreetCT CompiledFilter = "city-street-CT"
	CityStreetDC CompiledFilter = "city-street-DC"
	CityStreetDE CompiledFilter = "city-street-DE"
	CityStreetFL CompiledFilter = "city-street-FL"
	CityStreetGA CompiledFilter = "city-street-GA"
	CityStreetGU CompiledFilter = "city-street-GU"
	CityStreetHI CompiledFilter = "city-street-HI"
	CityStreetIA CompiledFilter = "city-street-IA"
	CityStreetID CompiledFilter = "city-street-ID"
	CityStreetIL CompiledFilter = "city-street-IL"
	CityStreetIN CompiledFilter = "city-street-IN"
	CityStreetKS CompiledFilter = "city-street-KS"
	CityStreetKY CompiledFilter = "city-street-KY"
	CityStreetLA CompiledFilter = "city-street-LA"
	CityStreetMA CompiledFilter = "city-street-MA"
	CityStreetMD CompiledFilter = "city-street-MD"
	CityStreetME CompiledFilter = "city-street-ME"
	CityStreetMI CompiledFilter = "city-street-MI"
	CityStreetMN CompiledFilter = "city-street-MN"
	CityStreetMO CompiledFilter = "city-street-MO"
	CityStreetMP CompiledFilter = "city-street-MP"
	CityStreetMS CompiledFilter = "city-street-MS"
	CityStreetMT CompiledFilter = "city-street-MT"
	CityStreetNC CompiledFilter = "city-street-NC"
	CityStreetND CompiledFilter = "city-street-ND"
	CityStreetNE CompiledFilter = "city-street-NE"
	CityStreetNH CompiledFilter = "city-street-NH"
	CityStreetNJ CompiledFilter = "city-street-NJ"
	CityStreetNM CompiledFilter = "city-street-NM"
	CityStreetNV CompiledFilter = "city-street-NV"
	CityStreetNY CompiledFilter = "city-street-NY"
	CityStreetOH CompiledFilter = "city-street-OH"
	CityStreetOK CompiledFilter = "city-street-OK"
	CityStreetOR CompiledFilter = "city-street-OR"
	CityStreetPA CompiledFilter = "city-street-PA"
	CityStreetPR CompiledFilter = "city-street-PR"
	CityStreetRI CompiledFilter = "city-street-RI"
	CityStreetSC CompiledFilter = "city-street-SC"
	CityStreetSD CompiledFilter = "city-street-SD"
	CityStreetTN CompiledFilter = "city-street-TN"
	CityStreetTX CompiledFilter = "city-street-TX"
	CityStreetUT CompiledFilter = "city-street-UT"
	CityStreetVA CompiledFilter = "city-street-VA"
	CityStreetVI CompiledFilter = "city-street-VI"
	CityStreetVT CompiledFilter = "city-street-VT"
	CityStreetWA CompiledFilter = "city-street-WA"
	CityStreetWI CompiledFilter = "city-street-WI"
	CityStreetWV CompiledFilter = "city-street-WV"
	CityStreetWY CompiledFilter = "city-street-WY"
	ZipStreet00  CompiledFilter = "zip-street-00"
	ZipStreet01  CompiledFilter = "zip-street-01"
	ZipStreet02  CompiledFilter = "zip-street-02"
	ZipStreet03  CompiledFilter = "zip-street-03"
	ZipStreet04  CompiledFilter = "zip-street-04"
	ZipStreet05  CompiledFilter = "zip-street-05"
	ZipStreet06  CompiledFilter = "zip-street-06"
	ZipStreet07  CompiledFilter = "zip-street-07"
	ZipStreet08  CompiledFilter = "zip-street-08"
	ZipStreet09  CompiledFilter = "zip-street-09"
	ZipStreet10  CompiledFilter = "zip-street-10"
	ZipStreet11  CompiledFilter = "zip-street-11"
	ZipStreet12  CompiledFilter = "zip-street-12"
	ZipStreet13  CompiledFilter = "zip-street-13"
	ZipStreet14  CompiledFilter = "zip-street-14"
	ZipStreet15  CompiledFilter = "zip-street-15"
	ZipStreet16  CompiledFilter = "zip-street-16"
	ZipStreet17  CompiledFilter = "zip-street-17"
	ZipStreet18  CompiledFilter = "zip-street-18"
	ZipStreet19  CompiledFilter = "zip-street-19"
	ZipStreet20  CompiledFilter = "zip-street-20"
	ZipStreet21  CompiledFilter = "zip-street-21"
	ZipStreet22  CompiledFilter = "zip-street-22"
	ZipStreet23  CompiledFilter = "zip-street-23"
	ZipStreet24  CompiledFilter = "zip-street-24"
	ZipStreet25  CompiledFilter = "zip-street-25"
	ZipStreet26  CompiledFilter = "zip-street-26"
	ZipStreet27  CompiledFilter = "zip-street-27"
	ZipStreet28  CompiledFilter = "zip-street-28"
	ZipStreet29  CompiledFilter = "zip-street-29"
	ZipStreet30  CompiledFilter = "zip-street-30"
	ZipStreet31  CompiledFilter = "zip-street-31"
	ZipStreet32  CompiledFilter = "zip-street-32"
	ZipStreet33  CompiledFilter = "zip-street-33"
	ZipStreet34  CompiledFilter = "zip-street-34"
	ZipStreet35  CompiledFilter = "zip-street-35"
	ZipStreet36  CompiledFilter = "zip-street-36"
	ZipStreet37  CompiledFilter = "zip-street-37"
	ZipStreet38  CompiledFilter = "zip-street-38"
	ZipStreet39  CompiledFilter = "zip-street-39"
	ZipStreet40  CompiledFilter = "zip-street-40"
	ZipStreet41  CompiledFilter = "zip-street-41"
	ZipStreet42  CompiledFilter = "zip-street-42"
	ZipStreet43  CompiledFilter = "zip-street-43"
	ZipStreet44  CompiledFilter = "zip-street-44"
	ZipStreet45  CompiledFilter = "zip-street-45"
	ZipStreet46  CompiledFilter = "zip-street-46"
	ZipStreet47  CompiledFilter = "zip-street-47"
	ZipStreet48  CompiledFilter = "zip-street-48"
	ZipStreet49  CompiledFilter = "zip-street-49"
	ZipStreet50  CompiledFilter = "zip-street-50"
	ZipStreet51  CompiledFilter = "zip-street-51"
	ZipStreet52  CompiledFilter = "zip-street-52"
	ZipStreet53  CompiledFilter = "zip-street-53"
	ZipStreet54  CompiledFilter = "zip-street-54"
	ZipStreet55  CompiledFilter = "zip-street-55"
	ZipStreet56  CompiledFilter = "zip-street-56"
	ZipStreet57  CompiledFilter = "zip-street-57"
	ZipStreet58  CompiledFilter = "zip-street-58"
	ZipStreet59  CompiledFilter = "zip-street-59"
	ZipStreet60  CompiledFilter = "zip-street-60"
	ZipStreet61  CompiledFilter = "zip-street-61"
	ZipStreet62  CompiledFilter = "zip-street-62"
	ZipStreet63  CompiledFilter = "zip-street-63"
	ZipStreet64  CompiledFilter = "zip-street-64"
	ZipStreet65  CompiledFilter = "zip-street-65"
	ZipStreet66  CompiledFilter = "zip-street-66"
	ZipStreet67  CompiledFilter = "zip-street-67"
	ZipStreet68  CompiledFilter = "zip-street-68"
	ZipStreet69  CompiledFilter = "zip-street-69"
	ZipStreet70  CompiledFilter = "zip-street-70"
	ZipStreet71  CompiledFilter = "zip-street-71"
	ZipStreet72  CompiledFilter = "zip-street-72"
	ZipStreet73  CompiledFilter = "zip-street-73"
	ZipStreet74  CompiledFilter = "zip-street-74"
	ZipStreet75  CompiledFilter = "zip-street-75"
	ZipStreet76  CompiledFilter = "zip-street-76"
	ZipStreet77  CompiledFilter = "zip-street-77"
	ZipStreet78  CompiledFilter = "zip-street-78"
	ZipStreet79  CompiledFilter = "zip-street-79"
	ZipStreet80  CompiledFilter = "zip-street-80"
	ZipStreet81  CompiledFilter = "zip-street-81"
	ZipStreet82  CompiledFilter = "zip-street-82"
	ZipStreet83  CompiledFilter = "zip-street-83"
	ZipStreet84  CompiledFilter = "zip-street-84"
	ZipStreet85  CompiledFilter = "zip-street-85"
	ZipStreet86  CompiledFilter = "zip-street-86"
	ZipStreet87  CompiledFilter = "zip-street-87"
	ZipStreet88  CompiledFilter = "zip-street-88"
	ZipStreet89  CompiledFilter = "zip-street-89"
	ZipStreet90  CompiledFilter = "zip-street-90"
	ZipStreet91  CompiledFilter = "zip-street-91"
	ZipStreet92  CompiledFilter = "zip-street-92"
	ZipStreet93  CompiledFilter = "zip-street-93"
	ZipStreet94  CompiledFilter = "zip-street-94"
	ZipStreet95  CompiledFilter = "zip-street-95"
	ZipStreet96  CompiledFilter = "zip-street-96"
	ZipStreet97  CompiledFilter = "zip-street-97"
	ZipStreet98  CompiledFilter = "zip-street-98"
	ZipStreet99  CompiledFilter = "zip-street-99"
)

var allCompiledFilters = map[string]CompiledFilter{
	"zip-city":       ZipCity,
	"city-street-AK": CityStreetAK,
	"city-street-AL": CityStreetAL,
	"city-street-AR": CityStreetAR,
	"city-street-AS": CityStreetAS,
	"city-street-AZ": CityStreetAZ,
	"city-street-CA": CityStreetCA,
	"city-street-CO": CityStreetCO,
	"city-street-CT": CityStreetCT,
	"city-street-DC": CityStreetDC,
	"city-street-DE": CityStreetDE,
	"city-street-FL": CityStreetFL,
	"city-street-GA": CityStreetGA,
	"city-street-GU": CityStreetGU,
	"city-street-HI": CityStreetHI,
	"city-street-IA": CityStreetIA,
	"city-street-ID": CityStreetID,
	"city-street-IL": CityStreetIL,
	"city-street-IN": CityStreetIN,
	"city-street-KS": CityStreetKS,
	"city-street-KY": CityStreetKY,
	"city-street-LA": CityStreetLA,
	"city-street-MA": CityStreetMA,
	"city-street-MD": CityStreetMD,
	"city-street-ME": CityStreetME,
	"city-street-MI": CityStreetMI,
	"city-street-MN": CityStreetMN,
	"city-street-MO": CityStreetMO,
	"city-street-MP": CityStreetMP,
	"city-street-MS": CityStreetMS,
	"city-street-MT": CityStreetMT,
	"city-street-NC": CityStreetNC,
	"city-street-ND": CityStreetND,
	"city-street-NE": CityStreetNE,
	"city-street-NH": CityStreetNH,
	"city-street-NJ": CityStreetNJ,
	"city-street-NM": CityStreetNM,
	"city-street-NV": CityStreetNV,
	"city-street-NY": CityStreetNY,
	"city-street-OH": CityStreetOH,
	"city-street-OK": CityStreetOK,
	"city-street-OR": CityStreetOR,
	"city-street-PA": CityStreetPA,
	"city-street-PR": CityStreetPR,
	"city-street-RI": CityStreetRI,
	"city-street-SC": CityStreetSC,
	"city-street-SD": CityStreetSD,
	"city-street-TN": CityStreetTN,
	"city-street-TX": CityStreetTX,
	"city-street-UT": CityStreetUT,
	"city-street-VA": CityStreetVA,
	"city-street-VI": CityStreetVI,
	"city-street-VT": CityStreetVT,
	"city-street-WA": CityStreetWA,
	"city-street-WI": CityStreetWI,
	"city-street-WV": CityStreetWV,
	"city-street-WY": CityStreetWY,
	"zip-street-00":  ZipStreet00,
	"zip-street-01":  ZipStreet01,
	"zip-street-02":  ZipStreet02,
	"zip-street-03":  ZipStreet03,
	"zip-street-04":  ZipStreet04,
	"zip-street-05":  ZipStreet05,
	"zip-street-06":  ZipStreet06,
	"zip-street-07":  ZipStreet07,
	"zip-street-08":  ZipStreet08,
	"zip-street-09":  ZipStreet09,
	"zip-street-10":  ZipStreet10,
	"zip-street-11":  ZipStreet11,
	"zip-street-12":  ZipStreet12,
	"zip-street-13":  ZipStreet13,
	"zip-street-14":  ZipStreet14,
	"zip-street-15":  ZipStreet15,
	"zip-street-16":  ZipStreet16,
	"zip-street-17":  ZipStreet17,
	"zip-street-18":  ZipStreet18,
	"zip-street-19":  ZipStreet19,
	"zip-street-20":  ZipStreet20,
	"zip-street-21":  ZipStreet21,
	"zip-street-22":  ZipStreet22,
	"zip-street-23":  ZipStreet23,
	"zip-street-24":  ZipStreet24,
	"zip-street-25":  ZipStreet25,
	"zip-street-26":  ZipStreet26,
	"zip-street-27":  ZipStreet27,
	"zip-street-28":  ZipStreet28,
	"zip-street-29":  ZipStreet29,
	"zip-street-30":  ZipStreet30,
	"zip-street-31":  ZipStreet31,
	"zip-street-32":  ZipStreet32,
	"zip-street-33":  ZipStreet33,
	"zip-street-34":  ZipStreet34,
	"zip-street-35":  ZipStreet35,
	"zip-street-36":  ZipStreet36,
	"zip-street-37":  ZipStreet37,
	"zip-street-38":  ZipStreet38,
	"zip-street-39":  ZipStreet39,
	"zip-street-40":  ZipStreet40,
	"zip-street-41":  ZipStreet41,
	"zip-street-42":  ZipStreet42,
	"zip-street-43":  ZipStreet43,
	"zip-street-44":  ZipStreet44,
	"zip-street-45":  ZipStreet45,
	"zip-street-46":  ZipStreet46,
	"zip-street-47":  ZipStreet47,
	"zip-street-48":  ZipStreet48,
	"zip-street-49":  ZipStreet49,
	"zip-street-50":  ZipStreet50,
	"zip-street-51":  ZipStreet51,
	"zip-street-52":  ZipStreet52,
	"zip-street-53":  ZipStreet53,
	"zip-street-54":  ZipStreet54,
	"zip-street-55":  ZipStreet55,
	"zip-street-56":  ZipStreet56,
	"zip-street-57":  ZipStreet57,
	"zip-street-58":  ZipStreet58,
	"zip-street-59":  ZipStreet59,
	"zip-street-60":  ZipStreet60,
	"zip-street-61":  ZipStreet61,
	"zip-street-62":  ZipStreet62,
	"zip-street-63":  ZipStreet63,
	"zip-street-64":  ZipStreet64,
	"zip-street-65":  ZipStreet65,
	"zip-street-66":  ZipStreet66,
	"zip-street-67":  ZipStreet67,
	"zip-street-68":  ZipStreet68,
	"zip-street-69":  ZipStreet69,
	"zip-street-70":  ZipStreet70,
	"zip-street-71":  ZipStreet71,
	"zip-street-72":  ZipStreet72,
	"zip-street-73":  ZipStreet73,
	"zip-street-74":  ZipStreet74,
	"zip-street-75":  ZipStreet75,
	"zip-street-76":  ZipStreet76,
	"zip-street-77":  ZipStreet77,
	"zip-street-78":  ZipStreet78,
	"zip-street-79":  ZipStreet79,
	"zip-street-80":  ZipStreet80,
	"zip-street-81":  ZipStreet81,
	"zip-street-82":  ZipStreet82,
	"zip-street-83":  ZipStreet83,
	"zip-street-84":  ZipStreet84,
	"zip-street-85":  ZipStreet85,
	"zip-street-86":  ZipStreet86,
	"zip-street-87":  ZipStreet87,
	"zip-street-88":  ZipStreet88,
	"zip-street-89":  ZipStreet89,
	"zip-street-90":  ZipStreet90,
	"zip-street-91":  ZipStreet91,
	"zip-street-92":  ZipStreet92,
	"zip-street-93":  ZipStreet93,
	"zip-street-94":  ZipStreet94,
	"zip-street-95":  ZipStreet95,
	"zip-street-96":  ZipStreet96,
	"zip-street-97":  ZipStreet97,
	"zip-street-98":  ZipStreet98,
	"zip-street-99":  ZipStreet99,
}
