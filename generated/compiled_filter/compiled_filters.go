// DO NOT EDIT! Code generated at 2026-09-13T03:26:49Z by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"regexp"

	bloom "github.com/bits-and-blooms/bloom/v3"
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

func CityStreetFilterForState(state string) (CompiledFilter, error) {
	if len(state) != 2 {
		return Unrecognized, fmt.Errorf("USPS state abbreviation required")
	}

	filterid := fmt.Sprintf("city-street-%s", state)
	switch filterid {
	case "city-street-AK":
		return CityStreetAK, nil
	case "city-street-AL":
		return CityStreetAL, nil
	case "city-street-AR":
		return CityStreetAR, nil
	case "city-street-AS":
		return CityStreetAS, nil
	case "city-street-AZ":
		return CityStreetAZ, nil
	case "city-street-CA":
		return CityStreetCA, nil
	case "city-street-CO":
		return CityStreetCO, nil
	case "city-street-CT":
		return CityStreetCT, nil
	case "city-street-DC":
		return CityStreetDC, nil
	case "city-street-DE":
		return CityStreetDE, nil
	case "city-street-FL":
		return CityStreetFL, nil
	case "city-street-GA":
		return CityStreetGA, nil
	case "city-street-GU":
		return CityStreetGU, nil
	case "city-street-HI":
		return CityStreetHI, nil
	case "city-street-IA":
		return CityStreetIA, nil
	case "city-street-ID":
		return CityStreetID, nil
	case "city-street-IL":
		return CityStreetIL, nil
	case "city-street-IN":
		return CityStreetIN, nil
	case "city-street-KS":
		return CityStreetKS, nil
	case "city-street-KY":
		return CityStreetKY, nil
	case "city-street-LA":
		return CityStreetLA, nil
	case "city-street-MA":
		return CityStreetMA, nil
	case "city-street-MD":
		return CityStreetMD, nil
	case "city-street-ME":
		return CityStreetME, nil
	case "city-street-MI":
		return CityStreetMI, nil
	case "city-street-MN":
		return CityStreetMN, nil
	case "city-street-MO":
		return CityStreetMO, nil
	case "city-street-MP":
		return CityStreetMP, nil
	case "city-street-MS":
		return CityStreetMS, nil
	case "city-street-MT":
		return CityStreetMT, nil
	case "city-street-NC":
		return CityStreetNC, nil
	case "city-street-ND":
		return CityStreetND, nil
	case "city-street-NE":
		return CityStreetNE, nil
	case "city-street-NH":
		return CityStreetNH, nil
	case "city-street-NJ":
		return CityStreetNJ, nil
	case "city-street-NM":
		return CityStreetNM, nil
	case "city-street-NV":
		return CityStreetNV, nil
	case "city-street-NY":
		return CityStreetNY, nil
	case "city-street-OH":
		return CityStreetOH, nil
	case "city-street-OK":
		return CityStreetOK, nil
	case "city-street-OR":
		return CityStreetOR, nil
	case "city-street-PA":
		return CityStreetPA, nil
	case "city-street-PR":
		return CityStreetPR, nil
	case "city-street-RI":
		return CityStreetRI, nil
	case "city-street-SC":
		return CityStreetSC, nil
	case "city-street-SD":
		return CityStreetSD, nil
	case "city-street-TN":
		return CityStreetTN, nil
	case "city-street-TX":
		return CityStreetTX, nil
	case "city-street-UT":
		return CityStreetUT, nil
	case "city-street-VA":
		return CityStreetVA, nil
	case "city-street-VI":
		return CityStreetVI, nil
	case "city-street-VT":
		return CityStreetVT, nil
	case "city-street-WA":
		return CityStreetWA, nil
	case "city-street-WI":
		return CityStreetWI, nil
	case "city-street-WV":
		return CityStreetWV, nil
	case "city-street-WY":
		return CityStreetWY, nil
	}
	return Unrecognized, fmt.Errorf("USPS state abbreviation required")
}

func ZipStreetFilterForZip(zip string) (CompiledFilter, error) {
	if !zip5pattern.MatchString(zip) {
		return Unrecognized, fmt.Errorf("5-digit zip code required")
	}
	zip2 := zip[0:2]
	filterid := fmt.Sprintf("zip-street-%s", zip2)
	switch filterid {
	case "zip-street-00":
		return ZipStreet00, nil
	case "zip-street-01":
		return ZipStreet01, nil
	case "zip-street-02":
		return ZipStreet02, nil
	case "zip-street-03":
		return ZipStreet03, nil
	case "zip-street-04":
		return ZipStreet04, nil
	case "zip-street-05":
		return ZipStreet05, nil
	case "zip-street-06":
		return ZipStreet06, nil
	case "zip-street-07":
		return ZipStreet07, nil
	case "zip-street-08":
		return ZipStreet08, nil
	case "zip-street-09":
		return ZipStreet09, nil
	case "zip-street-10":
		return ZipStreet10, nil
	case "zip-street-11":
		return ZipStreet11, nil
	case "zip-street-12":
		return ZipStreet12, nil
	case "zip-street-13":
		return ZipStreet13, nil
	case "zip-street-14":
		return ZipStreet14, nil
	case "zip-street-15":
		return ZipStreet15, nil
	case "zip-street-16":
		return ZipStreet16, nil
	case "zip-street-17":
		return ZipStreet17, nil
	case "zip-street-18":
		return ZipStreet18, nil
	case "zip-street-19":
		return ZipStreet19, nil
	case "zip-street-20":
		return ZipStreet20, nil
	case "zip-street-21":
		return ZipStreet21, nil
	case "zip-street-22":
		return ZipStreet22, nil
	case "zip-street-23":
		return ZipStreet23, nil
	case "zip-street-24":
		return ZipStreet24, nil
	case "zip-street-25":
		return ZipStreet25, nil
	case "zip-street-26":
		return ZipStreet26, nil
	case "zip-street-27":
		return ZipStreet27, nil
	case "zip-street-28":
		return ZipStreet28, nil
	case "zip-street-29":
		return ZipStreet29, nil
	case "zip-street-30":
		return ZipStreet30, nil
	case "zip-street-31":
		return ZipStreet31, nil
	case "zip-street-32":
		return ZipStreet32, nil
	case "zip-street-33":
		return ZipStreet33, nil
	case "zip-street-34":
		return ZipStreet34, nil
	case "zip-street-35":
		return ZipStreet35, nil
	case "zip-street-36":
		return ZipStreet36, nil
	case "zip-street-37":
		return ZipStreet37, nil
	case "zip-street-38":
		return ZipStreet38, nil
	case "zip-street-39":
		return ZipStreet39, nil
	case "zip-street-40":
		return ZipStreet40, nil
	case "zip-street-41":
		return ZipStreet41, nil
	case "zip-street-42":
		return ZipStreet42, nil
	case "zip-street-43":
		return ZipStreet43, nil
	case "zip-street-44":
		return ZipStreet44, nil
	case "zip-street-45":
		return ZipStreet45, nil
	case "zip-street-46":
		return ZipStreet46, nil
	case "zip-street-47":
		return ZipStreet47, nil
	case "zip-street-48":
		return ZipStreet48, nil
	case "zip-street-49":
		return ZipStreet49, nil
	case "zip-street-50":
		return ZipStreet50, nil
	case "zip-street-51":
		return ZipStreet51, nil
	case "zip-street-52":
		return ZipStreet52, nil
	case "zip-street-53":
		return ZipStreet53, nil
	case "zip-street-54":
		return ZipStreet54, nil
	case "zip-street-55":
		return ZipStreet55, nil
	case "zip-street-56":
		return ZipStreet56, nil
	case "zip-street-57":
		return ZipStreet57, nil
	case "zip-street-58":
		return ZipStreet58, nil
	case "zip-street-59":
		return ZipStreet59, nil
	case "zip-street-60":
		return ZipStreet60, nil
	case "zip-street-61":
		return ZipStreet61, nil
	case "zip-street-62":
		return ZipStreet62, nil
	case "zip-street-63":
		return ZipStreet63, nil
	case "zip-street-64":
		return ZipStreet64, nil
	case "zip-street-65":
		return ZipStreet65, nil
	case "zip-street-66":
		return ZipStreet66, nil
	case "zip-street-67":
		return ZipStreet67, nil
	case "zip-street-68":
		return ZipStreet68, nil
	case "zip-street-69":
		return ZipStreet69, nil
	case "zip-street-70":
		return ZipStreet70, nil
	case "zip-street-71":
		return ZipStreet71, nil
	case "zip-street-72":
		return ZipStreet72, nil
	case "zip-street-73":
		return ZipStreet73, nil
	case "zip-street-74":
		return ZipStreet74, nil
	case "zip-street-75":
		return ZipStreet75, nil
	case "zip-street-76":
		return ZipStreet76, nil
	case "zip-street-77":
		return ZipStreet77, nil
	case "zip-street-78":
		return ZipStreet78, nil
	case "zip-street-79":
		return ZipStreet79, nil
	case "zip-street-80":
		return ZipStreet80, nil
	case "zip-street-81":
		return ZipStreet81, nil
	case "zip-street-82":
		return ZipStreet82, nil
	case "zip-street-83":
		return ZipStreet83, nil
	case "zip-street-84":
		return ZipStreet84, nil
	case "zip-street-85":
		return ZipStreet85, nil
	case "zip-street-86":
		return ZipStreet86, nil
	case "zip-street-87":
		return ZipStreet87, nil
	case "zip-street-88":
		return ZipStreet88, nil
	case "zip-street-89":
		return ZipStreet89, nil
	case "zip-street-90":
		return ZipStreet90, nil
	case "zip-street-91":
		return ZipStreet91, nil
	case "zip-street-92":
		return ZipStreet92, nil
	case "zip-street-93":
		return ZipStreet93, nil
	case "zip-street-94":
		return ZipStreet94, nil
	case "zip-street-95":
		return ZipStreet95, nil
	case "zip-street-96":
		return ZipStreet96, nil
	case "zip-street-97":
		return ZipStreet97, nil
	case "zip-street-98":
		return ZipStreet98, nil
	case "zip-street-99":
		return ZipStreet99, nil
	}
	return Unrecognized, fmt.Errorf("5-digit zip code required")
}

// LoadFilter restores the compiled filter in memory
func LoadFilter(name CompiledFilter) (*bloom.BloomFilter, error) {
	filter := &bloom.BloomFilter{}
	var reader io.Reader
	switch name {
	case ZipCity:
		reader = bytes.NewReader(RawZipCityFilterBytes)
	case CityStreetAK:
		reader = bytes.NewReader(RawCityStreetAKFilterBytes)
	case CityStreetAL:
		reader = bytes.NewReader(RawCityStreetALFilterBytes)
	case CityStreetAR:
		reader = bytes.NewReader(RawCityStreetARFilterBytes)
	case CityStreetAS:
		reader = bytes.NewReader(RawCityStreetASFilterBytes)
	case CityStreetAZ:
		reader = bytes.NewReader(RawCityStreetAZFilterBytes)
	case CityStreetCA:
		reader = bytes.NewReader(RawCityStreetCAFilterBytes)
	case CityStreetCO:
		reader = bytes.NewReader(RawCityStreetCOFilterBytes)
	case CityStreetCT:
		reader = bytes.NewReader(RawCityStreetCTFilterBytes)
	case CityStreetDC:
		reader = bytes.NewReader(RawCityStreetDCFilterBytes)
	case CityStreetDE:
		reader = bytes.NewReader(RawCityStreetDEFilterBytes)
	case CityStreetFL:
		reader = bytes.NewReader(RawCityStreetFLFilterBytes)
	case CityStreetGA:
		reader = bytes.NewReader(RawCityStreetGAFilterBytes)
	case CityStreetGU:
		reader = bytes.NewReader(RawCityStreetGUFilterBytes)
	case CityStreetHI:
		reader = bytes.NewReader(RawCityStreetHIFilterBytes)
	case CityStreetIA:
		reader = bytes.NewReader(RawCityStreetIAFilterBytes)
	case CityStreetID:
		reader = bytes.NewReader(RawCityStreetIDFilterBytes)
	case CityStreetIL:
		reader = bytes.NewReader(RawCityStreetILFilterBytes)
	case CityStreetIN:
		reader = bytes.NewReader(RawCityStreetINFilterBytes)
	case CityStreetKS:
		reader = bytes.NewReader(RawCityStreetKSFilterBytes)
	case CityStreetKY:
		reader = bytes.NewReader(RawCityStreetKYFilterBytes)
	case CityStreetLA:
		reader = bytes.NewReader(RawCityStreetLAFilterBytes)
	case CityStreetMA:
		reader = bytes.NewReader(RawCityStreetMAFilterBytes)
	case CityStreetMD:
		reader = bytes.NewReader(RawCityStreetMDFilterBytes)
	case CityStreetME:
		reader = bytes.NewReader(RawCityStreetMEFilterBytes)
	case CityStreetMI:
		reader = bytes.NewReader(RawCityStreetMIFilterBytes)
	case CityStreetMN:
		reader = bytes.NewReader(RawCityStreetMNFilterBytes)
	case CityStreetMO:
		reader = bytes.NewReader(RawCityStreetMOFilterBytes)
	case CityStreetMP:
		reader = bytes.NewReader(RawCityStreetMPFilterBytes)
	case CityStreetMS:
		reader = bytes.NewReader(RawCityStreetMSFilterBytes)
	case CityStreetMT:
		reader = bytes.NewReader(RawCityStreetMTFilterBytes)
	case CityStreetNC:
		reader = bytes.NewReader(RawCityStreetNCFilterBytes)
	case CityStreetND:
		reader = bytes.NewReader(RawCityStreetNDFilterBytes)
	case CityStreetNE:
		reader = bytes.NewReader(RawCityStreetNEFilterBytes)
	case CityStreetNH:
		reader = bytes.NewReader(RawCityStreetNHFilterBytes)
	case CityStreetNJ:
		reader = bytes.NewReader(RawCityStreetNJFilterBytes)
	case CityStreetNM:
		reader = bytes.NewReader(RawCityStreetNMFilterBytes)
	case CityStreetNV:
		reader = bytes.NewReader(RawCityStreetNVFilterBytes)
	case CityStreetNY:
		reader = bytes.NewReader(RawCityStreetNYFilterBytes)
	case CityStreetOH:
		reader = bytes.NewReader(RawCityStreetOHFilterBytes)
	case CityStreetOK:
		reader = bytes.NewReader(RawCityStreetOKFilterBytes)
	case CityStreetOR:
		reader = bytes.NewReader(RawCityStreetORFilterBytes)
	case CityStreetPA:
		reader = bytes.NewReader(RawCityStreetPAFilterBytes)
	case CityStreetPR:
		reader = bytes.NewReader(RawCityStreetPRFilterBytes)
	case CityStreetRI:
		reader = bytes.NewReader(RawCityStreetRIFilterBytes)
	case CityStreetSC:
		reader = bytes.NewReader(RawCityStreetSCFilterBytes)
	case CityStreetSD:
		reader = bytes.NewReader(RawCityStreetSDFilterBytes)
	case CityStreetTN:
		reader = bytes.NewReader(RawCityStreetTNFilterBytes)
	case CityStreetTX:
		reader = bytes.NewReader(RawCityStreetTXFilterBytes)
	case CityStreetUT:
		reader = bytes.NewReader(RawCityStreetUTFilterBytes)
	case CityStreetVA:
		reader = bytes.NewReader(RawCityStreetVAFilterBytes)
	case CityStreetVI:
		reader = bytes.NewReader(RawCityStreetVIFilterBytes)
	case CityStreetVT:
		reader = bytes.NewReader(RawCityStreetVTFilterBytes)
	case CityStreetWA:
		reader = bytes.NewReader(RawCityStreetWAFilterBytes)
	case CityStreetWI:
		reader = bytes.NewReader(RawCityStreetWIFilterBytes)
	case CityStreetWV:
		reader = bytes.NewReader(RawCityStreetWVFilterBytes)
	case CityStreetWY:
		reader = bytes.NewReader(RawCityStreetWYFilterBytes)
	case ZipStreet00:
		reader = bytes.NewReader(RawZipStreet00FilterBytes)
	case ZipStreet01:
		reader = bytes.NewReader(RawZipStreet01FilterBytes)
	case ZipStreet02:
		reader = bytes.NewReader(RawZipStreet02FilterBytes)
	case ZipStreet03:
		reader = bytes.NewReader(RawZipStreet03FilterBytes)
	case ZipStreet04:
		reader = bytes.NewReader(RawZipStreet04FilterBytes)
	case ZipStreet05:
		reader = bytes.NewReader(RawZipStreet05FilterBytes)
	case ZipStreet06:
		reader = bytes.NewReader(RawZipStreet06FilterBytes)
	case ZipStreet07:
		reader = bytes.NewReader(RawZipStreet07FilterBytes)
	case ZipStreet08:
		reader = bytes.NewReader(RawZipStreet08FilterBytes)
	case ZipStreet09:
		reader = bytes.NewReader(RawZipStreet09FilterBytes)
	case ZipStreet10:
		reader = bytes.NewReader(RawZipStreet10FilterBytes)
	case ZipStreet11:
		reader = bytes.NewReader(RawZipStreet11FilterBytes)
	case ZipStreet12:
		reader = bytes.NewReader(RawZipStreet12FilterBytes)
	case ZipStreet13:
		reader = bytes.NewReader(RawZipStreet13FilterBytes)
	case ZipStreet14:
		reader = bytes.NewReader(RawZipStreet14FilterBytes)
	case ZipStreet15:
		reader = bytes.NewReader(RawZipStreet15FilterBytes)
	case ZipStreet16:
		reader = bytes.NewReader(RawZipStreet16FilterBytes)
	case ZipStreet17:
		reader = bytes.NewReader(RawZipStreet17FilterBytes)
	case ZipStreet18:
		reader = bytes.NewReader(RawZipStreet18FilterBytes)
	case ZipStreet19:
		reader = bytes.NewReader(RawZipStreet19FilterBytes)
	case ZipStreet20:
		reader = bytes.NewReader(RawZipStreet20FilterBytes)
	case ZipStreet21:
		reader = bytes.NewReader(RawZipStreet21FilterBytes)
	case ZipStreet22:
		reader = bytes.NewReader(RawZipStreet22FilterBytes)
	case ZipStreet23:
		reader = bytes.NewReader(RawZipStreet23FilterBytes)
	case ZipStreet24:
		reader = bytes.NewReader(RawZipStreet24FilterBytes)
	case ZipStreet25:
		reader = bytes.NewReader(RawZipStreet25FilterBytes)
	case ZipStreet26:
		reader = bytes.NewReader(RawZipStreet26FilterBytes)
	case ZipStreet27:
		reader = bytes.NewReader(RawZipStreet27FilterBytes)
	case ZipStreet28:
		reader = bytes.NewReader(RawZipStreet28FilterBytes)
	case ZipStreet29:
		reader = bytes.NewReader(RawZipStreet29FilterBytes)
	case ZipStreet30:
		reader = bytes.NewReader(RawZipStreet30FilterBytes)
	case ZipStreet31:
		reader = bytes.NewReader(RawZipStreet31FilterBytes)
	case ZipStreet32:
		reader = bytes.NewReader(RawZipStreet32FilterBytes)
	case ZipStreet33:
		reader = bytes.NewReader(RawZipStreet33FilterBytes)
	case ZipStreet34:
		reader = bytes.NewReader(RawZipStreet34FilterBytes)
	case ZipStreet35:
		reader = bytes.NewReader(RawZipStreet35FilterBytes)
	case ZipStreet36:
		reader = bytes.NewReader(RawZipStreet36FilterBytes)
	case ZipStreet37:
		reader = bytes.NewReader(RawZipStreet37FilterBytes)
	case ZipStreet38:
		reader = bytes.NewReader(RawZipStreet38FilterBytes)
	case ZipStreet39:
		reader = bytes.NewReader(RawZipStreet39FilterBytes)
	case ZipStreet40:
		reader = bytes.NewReader(RawZipStreet40FilterBytes)
	case ZipStreet41:
		reader = bytes.NewReader(RawZipStreet41FilterBytes)
	case ZipStreet42:
		reader = bytes.NewReader(RawZipStreet42FilterBytes)
	case ZipStreet43:
		reader = bytes.NewReader(RawZipStreet43FilterBytes)
	case ZipStreet44:
		reader = bytes.NewReader(RawZipStreet44FilterBytes)
	case ZipStreet45:
		reader = bytes.NewReader(RawZipStreet45FilterBytes)
	case ZipStreet46:
		reader = bytes.NewReader(RawZipStreet46FilterBytes)
	case ZipStreet47:
		reader = bytes.NewReader(RawZipStreet47FilterBytes)
	case ZipStreet48:
		reader = bytes.NewReader(RawZipStreet48FilterBytes)
	case ZipStreet49:
		reader = bytes.NewReader(RawZipStreet49FilterBytes)
	case ZipStreet50:
		reader = bytes.NewReader(RawZipStreet50FilterBytes)
	case ZipStreet51:
		reader = bytes.NewReader(RawZipStreet51FilterBytes)
	case ZipStreet52:
		reader = bytes.NewReader(RawZipStreet52FilterBytes)
	case ZipStreet53:
		reader = bytes.NewReader(RawZipStreet53FilterBytes)
	case ZipStreet54:
		reader = bytes.NewReader(RawZipStreet54FilterBytes)
	case ZipStreet55:
		reader = bytes.NewReader(RawZipStreet55FilterBytes)
	case ZipStreet56:
		reader = bytes.NewReader(RawZipStreet56FilterBytes)
	case ZipStreet57:
		reader = bytes.NewReader(RawZipStreet57FilterBytes)
	case ZipStreet58:
		reader = bytes.NewReader(RawZipStreet58FilterBytes)
	case ZipStreet59:
		reader = bytes.NewReader(RawZipStreet59FilterBytes)
	case ZipStreet60:
		reader = bytes.NewReader(RawZipStreet60FilterBytes)
	case ZipStreet61:
		reader = bytes.NewReader(RawZipStreet61FilterBytes)
	case ZipStreet62:
		reader = bytes.NewReader(RawZipStreet62FilterBytes)
	case ZipStreet63:
		reader = bytes.NewReader(RawZipStreet63FilterBytes)
	case ZipStreet64:
		reader = bytes.NewReader(RawZipStreet64FilterBytes)
	case ZipStreet65:
		reader = bytes.NewReader(RawZipStreet65FilterBytes)
	case ZipStreet66:
		reader = bytes.NewReader(RawZipStreet66FilterBytes)
	case ZipStreet67:
		reader = bytes.NewReader(RawZipStreet67FilterBytes)
	case ZipStreet68:
		reader = bytes.NewReader(RawZipStreet68FilterBytes)
	case ZipStreet69:
		reader = bytes.NewReader(RawZipStreet69FilterBytes)
	case ZipStreet70:
		reader = bytes.NewReader(RawZipStreet70FilterBytes)
	case ZipStreet71:
		reader = bytes.NewReader(RawZipStreet71FilterBytes)
	case ZipStreet72:
		reader = bytes.NewReader(RawZipStreet72FilterBytes)
	case ZipStreet73:
		reader = bytes.NewReader(RawZipStreet73FilterBytes)
	case ZipStreet74:
		reader = bytes.NewReader(RawZipStreet74FilterBytes)
	case ZipStreet75:
		reader = bytes.NewReader(RawZipStreet75FilterBytes)
	case ZipStreet76:
		reader = bytes.NewReader(RawZipStreet76FilterBytes)
	case ZipStreet77:
		reader = bytes.NewReader(RawZipStreet77FilterBytes)
	case ZipStreet78:
		reader = bytes.NewReader(RawZipStreet78FilterBytes)
	case ZipStreet79:
		reader = bytes.NewReader(RawZipStreet79FilterBytes)
	case ZipStreet80:
		reader = bytes.NewReader(RawZipStreet80FilterBytes)
	case ZipStreet81:
		reader = bytes.NewReader(RawZipStreet81FilterBytes)
	case ZipStreet82:
		reader = bytes.NewReader(RawZipStreet82FilterBytes)
	case ZipStreet83:
		reader = bytes.NewReader(RawZipStreet83FilterBytes)
	case ZipStreet84:
		reader = bytes.NewReader(RawZipStreet84FilterBytes)
	case ZipStreet85:
		reader = bytes.NewReader(RawZipStreet85FilterBytes)
	case ZipStreet86:
		reader = bytes.NewReader(RawZipStreet86FilterBytes)
	case ZipStreet87:
		reader = bytes.NewReader(RawZipStreet87FilterBytes)
	case ZipStreet88:
		reader = bytes.NewReader(RawZipStreet88FilterBytes)
	case ZipStreet89:
		reader = bytes.NewReader(RawZipStreet89FilterBytes)
	case ZipStreet90:
		reader = bytes.NewReader(RawZipStreet90FilterBytes)
	case ZipStreet91:
		reader = bytes.NewReader(RawZipStreet91FilterBytes)
	case ZipStreet92:
		reader = bytes.NewReader(RawZipStreet92FilterBytes)
	case ZipStreet93:
		reader = bytes.NewReader(RawZipStreet93FilterBytes)
	case ZipStreet94:
		reader = bytes.NewReader(RawZipStreet94FilterBytes)
	case ZipStreet95:
		reader = bytes.NewReader(RawZipStreet95FilterBytes)
	case ZipStreet96:
		reader = bytes.NewReader(RawZipStreet96FilterBytes)
	case ZipStreet97:
		reader = bytes.NewReader(RawZipStreet97FilterBytes)
	case ZipStreet98:
		reader = bytes.NewReader(RawZipStreet98FilterBytes)
	case ZipStreet99:
		reader = bytes.NewReader(RawZipStreet99FilterBytes)
	default:
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	_, err := filter.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s bloom filter: %w", name, err)
	}
	return filter, nil
}

//go:embed city-street-AK.bin
var RawCityStreetAKFilterBytes []byte

//go:embed city-street-AL.bin
var RawCityStreetALFilterBytes []byte

//go:embed city-street-AR.bin
var RawCityStreetARFilterBytes []byte

//go:embed city-street-AS.bin
var RawCityStreetASFilterBytes []byte

//go:embed city-street-AZ.bin
var RawCityStreetAZFilterBytes []byte

//go:embed city-street-CA.bin
var RawCityStreetCAFilterBytes []byte

//go:embed city-street-CO.bin
var RawCityStreetCOFilterBytes []byte

//go:embed city-street-CT.bin
var RawCityStreetCTFilterBytes []byte

//go:embed city-street-DC.bin
var RawCityStreetDCFilterBytes []byte

//go:embed city-street-DE.bin
var RawCityStreetDEFilterBytes []byte

//go:embed city-street-FL.bin
var RawCityStreetFLFilterBytes []byte

//go:embed city-street-GA.bin
var RawCityStreetGAFilterBytes []byte

//go:embed city-street-GU.bin
var RawCityStreetGUFilterBytes []byte

//go:embed city-street-HI.bin
var RawCityStreetHIFilterBytes []byte

//go:embed city-street-IA.bin
var RawCityStreetIAFilterBytes []byte

//go:embed city-street-ID.bin
var RawCityStreetIDFilterBytes []byte

//go:embed city-street-IL.bin
var RawCityStreetILFilterBytes []byte

//go:embed city-street-IN.bin
var RawCityStreetINFilterBytes []byte

//go:embed city-street-KS.bin
var RawCityStreetKSFilterBytes []byte

//go:embed city-street-KY.bin
var RawCityStreetKYFilterBytes []byte

//go:embed city-street-LA.bin
var RawCityStreetLAFilterBytes []byte

//go:embed city-street-MA.bin
var RawCityStreetMAFilterBytes []byte

//go:embed city-street-MD.bin
var RawCityStreetMDFilterBytes []byte

//go:embed city-street-ME.bin
var RawCityStreetMEFilterBytes []byte

//go:embed city-street-MI.bin
var RawCityStreetMIFilterBytes []byte

//go:embed city-street-MN.bin
var RawCityStreetMNFilterBytes []byte

//go:embed city-street-MO.bin
var RawCityStreetMOFilterBytes []byte

//go:embed city-street-MP.bin
var RawCityStreetMPFilterBytes []byte

//go:embed city-street-MS.bin
var RawCityStreetMSFilterBytes []byte

//go:embed city-street-MT.bin
var RawCityStreetMTFilterBytes []byte

//go:embed city-street-NC.bin
var RawCityStreetNCFilterBytes []byte

//go:embed city-street-ND.bin
var RawCityStreetNDFilterBytes []byte

//go:embed city-street-NE.bin
var RawCityStreetNEFilterBytes []byte

//go:embed city-street-NH.bin
var RawCityStreetNHFilterBytes []byte

//go:embed city-street-NJ.bin
var RawCityStreetNJFilterBytes []byte

//go:embed city-street-NM.bin
var RawCityStreetNMFilterBytes []byte

//go:embed city-street-NV.bin
var RawCityStreetNVFilterBytes []byte

//go:embed city-street-NY.bin
var RawCityStreetNYFilterBytes []byte

//go:embed city-street-OH.bin
var RawCityStreetOHFilterBytes []byte

//go:embed city-street-OK.bin
var RawCityStreetOKFilterBytes []byte

//go:embed city-street-OR.bin
var RawCityStreetORFilterBytes []byte

//go:embed city-street-PA.bin
var RawCityStreetPAFilterBytes []byte

//go:embed city-street-PR.bin
var RawCityStreetPRFilterBytes []byte

//go:embed city-street-RI.bin
var RawCityStreetRIFilterBytes []byte

//go:embed city-street-SC.bin
var RawCityStreetSCFilterBytes []byte

//go:embed city-street-SD.bin
var RawCityStreetSDFilterBytes []byte

//go:embed city-street-TN.bin
var RawCityStreetTNFilterBytes []byte

//go:embed city-street-TX.bin
var RawCityStreetTXFilterBytes []byte

//go:embed city-street-UT.bin
var RawCityStreetUTFilterBytes []byte

//go:embed city-street-VA.bin
var RawCityStreetVAFilterBytes []byte

//go:embed city-street-VI.bin
var RawCityStreetVIFilterBytes []byte

//go:embed city-street-VT.bin
var RawCityStreetVTFilterBytes []byte

//go:embed city-street-WA.bin
var RawCityStreetWAFilterBytes []byte

//go:embed city-street-WI.bin
var RawCityStreetWIFilterBytes []byte

//go:embed city-street-WV.bin
var RawCityStreetWVFilterBytes []byte

//go:embed city-street-WY.bin
var RawCityStreetWYFilterBytes []byte

//go:embed zip-street-00.bin
var RawZipStreet00FilterBytes []byte

//go:embed zip-street-01.bin
var RawZipStreet01FilterBytes []byte

//go:embed zip-street-02.bin
var RawZipStreet02FilterBytes []byte

//go:embed zip-street-03.bin
var RawZipStreet03FilterBytes []byte

//go:embed zip-street-04.bin
var RawZipStreet04FilterBytes []byte

//go:embed zip-street-05.bin
var RawZipStreet05FilterBytes []byte

//go:embed zip-street-06.bin
var RawZipStreet06FilterBytes []byte

//go:embed zip-street-07.bin
var RawZipStreet07FilterBytes []byte

//go:embed zip-street-08.bin
var RawZipStreet08FilterBytes []byte

//go:embed zip-street-09.bin
var RawZipStreet09FilterBytes []byte

//go:embed zip-street-10.bin
var RawZipStreet10FilterBytes []byte

//go:embed zip-street-11.bin
var RawZipStreet11FilterBytes []byte

//go:embed zip-street-12.bin
var RawZipStreet12FilterBytes []byte

//go:embed zip-street-13.bin
var RawZipStreet13FilterBytes []byte

//go:embed zip-street-14.bin
var RawZipStreet14FilterBytes []byte

//go:embed zip-street-15.bin
var RawZipStreet15FilterBytes []byte

//go:embed zip-street-16.bin
var RawZipStreet16FilterBytes []byte

//go:embed zip-street-17.bin
var RawZipStreet17FilterBytes []byte

//go:embed zip-street-18.bin
var RawZipStreet18FilterBytes []byte

//go:embed zip-street-19.bin
var RawZipStreet19FilterBytes []byte

//go:embed zip-street-20.bin
var RawZipStreet20FilterBytes []byte

//go:embed zip-street-21.bin
var RawZipStreet21FilterBytes []byte

//go:embed zip-street-22.bin
var RawZipStreet22FilterBytes []byte

//go:embed zip-street-23.bin
var RawZipStreet23FilterBytes []byte

//go:embed zip-street-24.bin
var RawZipStreet24FilterBytes []byte

//go:embed zip-street-25.bin
var RawZipStreet25FilterBytes []byte

//go:embed zip-street-26.bin
var RawZipStreet26FilterBytes []byte

//go:embed zip-street-27.bin
var RawZipStreet27FilterBytes []byte

//go:embed zip-street-28.bin
var RawZipStreet28FilterBytes []byte

//go:embed zip-street-29.bin
var RawZipStreet29FilterBytes []byte

//go:embed zip-street-30.bin
var RawZipStreet30FilterBytes []byte

//go:embed zip-street-31.bin
var RawZipStreet31FilterBytes []byte

//go:embed zip-street-32.bin
var RawZipStreet32FilterBytes []byte

//go:embed zip-street-33.bin
var RawZipStreet33FilterBytes []byte

//go:embed zip-street-34.bin
var RawZipStreet34FilterBytes []byte

//go:embed zip-street-35.bin
var RawZipStreet35FilterBytes []byte

//go:embed zip-street-36.bin
var RawZipStreet36FilterBytes []byte

//go:embed zip-street-37.bin
var RawZipStreet37FilterBytes []byte

//go:embed zip-street-38.bin
var RawZipStreet38FilterBytes []byte

//go:embed zip-street-39.bin
var RawZipStreet39FilterBytes []byte

//go:embed zip-street-40.bin
var RawZipStreet40FilterBytes []byte

//go:embed zip-street-41.bin
var RawZipStreet41FilterBytes []byte

//go:embed zip-street-42.bin
var RawZipStreet42FilterBytes []byte

//go:embed zip-street-43.bin
var RawZipStreet43FilterBytes []byte

//go:embed zip-street-44.bin
var RawZipStreet44FilterBytes []byte

//go:embed zip-street-45.bin
var RawZipStreet45FilterBytes []byte

//go:embed zip-street-46.bin
var RawZipStreet46FilterBytes []byte

//go:embed zip-street-47.bin
var RawZipStreet47FilterBytes []byte

//go:embed zip-street-48.bin
var RawZipStreet48FilterBytes []byte

//go:embed zip-street-49.bin
var RawZipStreet49FilterBytes []byte

//go:embed zip-street-50.bin
var RawZipStreet50FilterBytes []byte

//go:embed zip-street-51.bin
var RawZipStreet51FilterBytes []byte

//go:embed zip-street-52.bin
var RawZipStreet52FilterBytes []byte

//go:embed zip-street-53.bin
var RawZipStreet53FilterBytes []byte

//go:embed zip-street-54.bin
var RawZipStreet54FilterBytes []byte

//go:embed zip-street-55.bin
var RawZipStreet55FilterBytes []byte

//go:embed zip-street-56.bin
var RawZipStreet56FilterBytes []byte

//go:embed zip-street-57.bin
var RawZipStreet57FilterBytes []byte

//go:embed zip-street-58.bin
var RawZipStreet58FilterBytes []byte

//go:embed zip-street-59.bin
var RawZipStreet59FilterBytes []byte

//go:embed zip-street-60.bin
var RawZipStreet60FilterBytes []byte

//go:embed zip-street-61.bin
var RawZipStreet61FilterBytes []byte

//go:embed zip-street-62.bin
var RawZipStreet62FilterBytes []byte

//go:embed zip-street-63.bin
var RawZipStreet63FilterBytes []byte

//go:embed zip-street-64.bin
var RawZipStreet64FilterBytes []byte

//go:embed zip-street-65.bin
var RawZipStreet65FilterBytes []byte

//go:embed zip-street-66.bin
var RawZipStreet66FilterBytes []byte

//go:embed zip-street-67.bin
var RawZipStreet67FilterBytes []byte

//go:embed zip-street-68.bin
var RawZipStreet68FilterBytes []byte

//go:embed zip-street-69.bin
var RawZipStreet69FilterBytes []byte

//go:embed zip-street-70.bin
var RawZipStreet70FilterBytes []byte

//go:embed zip-street-71.bin
var RawZipStreet71FilterBytes []byte

//go:embed zip-street-72.bin
var RawZipStreet72FilterBytes []byte

//go:embed zip-street-73.bin
var RawZipStreet73FilterBytes []byte

//go:embed zip-street-74.bin
var RawZipStreet74FilterBytes []byte

//go:embed zip-street-75.bin
var RawZipStreet75FilterBytes []byte

//go:embed zip-street-76.bin
var RawZipStreet76FilterBytes []byte

//go:embed zip-street-77.bin
var RawZipStreet77FilterBytes []byte

//go:embed zip-street-78.bin
var RawZipStreet78FilterBytes []byte

//go:embed zip-street-79.bin
var RawZipStreet79FilterBytes []byte

//go:embed zip-street-80.bin
var RawZipStreet80FilterBytes []byte

//go:embed zip-street-81.bin
var RawZipStreet81FilterBytes []byte

//go:embed zip-street-82.bin
var RawZipStreet82FilterBytes []byte

//go:embed zip-street-83.bin
var RawZipStreet83FilterBytes []byte

//go:embed zip-street-84.bin
var RawZipStreet84FilterBytes []byte

//go:embed zip-street-85.bin
var RawZipStreet85FilterBytes []byte

//go:embed zip-street-86.bin
var RawZipStreet86FilterBytes []byte

//go:embed zip-street-87.bin
var RawZipStreet87FilterBytes []byte

//go:embed zip-street-88.bin
var RawZipStreet88FilterBytes []byte

//go:embed zip-street-89.bin
var RawZipStreet89FilterBytes []byte

//go:embed zip-street-90.bin
var RawZipStreet90FilterBytes []byte

//go:embed zip-street-91.bin
var RawZipStreet91FilterBytes []byte

//go:embed zip-street-92.bin
var RawZipStreet92FilterBytes []byte

//go:embed zip-street-93.bin
var RawZipStreet93FilterBytes []byte

//go:embed zip-street-94.bin
var RawZipStreet94FilterBytes []byte

//go:embed zip-street-95.bin
var RawZipStreet95FilterBytes []byte

//go:embed zip-street-96.bin
var RawZipStreet96FilterBytes []byte

//go:embed zip-street-97.bin
var RawZipStreet97FilterBytes []byte

//go:embed zip-street-98.bin
var RawZipStreet98FilterBytes []byte

//go:embed zip-street-99.bin
var RawZipStreet99FilterBytes []byte

// RawZipCityFilterBytes holds the pre-compiled zip-city Bloom filter
//
//go:embed zip-city.bin
var RawZipCityFilterBytes []byte
