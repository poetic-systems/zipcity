// DO NOT EDIT! Code generated at 2026-08-29T18:28:22Z by internal/bloomgenerator/bloomgenerator.go
package compiled_filter

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"fmt"
	"regexp"
	bloom "github.com/bits-and-blooms/bloom/v3"
)

var zip5pattern = regexp.MustCompile(`^\d{5}$`)

type CompiledFilter string

const (
	Unrecognized  CompiledFilter = ""
  ZipCity       CompiledFilter = "zip-city"
  CityStreet    CompiledFilter = "city-street"
	ZipStreet00   CompiledFilter = "zip-street-00"
	ZipStreet01   CompiledFilter = "zip-street-01"
	ZipStreet02   CompiledFilter = "zip-street-02"
	ZipStreet03   CompiledFilter = "zip-street-03"
	ZipStreet04   CompiledFilter = "zip-street-04"
	ZipStreet05   CompiledFilter = "zip-street-05"
	ZipStreet06   CompiledFilter = "zip-street-06"
	ZipStreet07   CompiledFilter = "zip-street-07"
	ZipStreet08   CompiledFilter = "zip-street-08"
	ZipStreet09   CompiledFilter = "zip-street-09"
	ZipStreet10   CompiledFilter = "zip-street-10"
	ZipStreet11   CompiledFilter = "zip-street-11"
	ZipStreet12   CompiledFilter = "zip-street-12"
	ZipStreet13   CompiledFilter = "zip-street-13"
	ZipStreet14   CompiledFilter = "zip-street-14"
	ZipStreet15   CompiledFilter = "zip-street-15"
	ZipStreet16   CompiledFilter = "zip-street-16"
	ZipStreet17   CompiledFilter = "zip-street-17"
	ZipStreet18   CompiledFilter = "zip-street-18"
	ZipStreet19   CompiledFilter = "zip-street-19"
	ZipStreet20   CompiledFilter = "zip-street-20"
	ZipStreet21   CompiledFilter = "zip-street-21"
	ZipStreet22   CompiledFilter = "zip-street-22"
	ZipStreet23   CompiledFilter = "zip-street-23"
	ZipStreet24   CompiledFilter = "zip-street-24"
	ZipStreet25   CompiledFilter = "zip-street-25"
	ZipStreet26   CompiledFilter = "zip-street-26"
	ZipStreet27   CompiledFilter = "zip-street-27"
	ZipStreet28   CompiledFilter = "zip-street-28"
	ZipStreet29   CompiledFilter = "zip-street-29"
	ZipStreet30   CompiledFilter = "zip-street-30"
	ZipStreet31   CompiledFilter = "zip-street-31"
	ZipStreet32   CompiledFilter = "zip-street-32"
	ZipStreet33   CompiledFilter = "zip-street-33"
	ZipStreet34   CompiledFilter = "zip-street-34"
	ZipStreet35   CompiledFilter = "zip-street-35"
	ZipStreet36   CompiledFilter = "zip-street-36"
	ZipStreet37   CompiledFilter = "zip-street-37"
	ZipStreet38   CompiledFilter = "zip-street-38"
	ZipStreet39   CompiledFilter = "zip-street-39"
	ZipStreet40   CompiledFilter = "zip-street-40"
	ZipStreet41   CompiledFilter = "zip-street-41"
	ZipStreet42   CompiledFilter = "zip-street-42"
	ZipStreet43   CompiledFilter = "zip-street-43"
	ZipStreet44   CompiledFilter = "zip-street-44"
	ZipStreet45   CompiledFilter = "zip-street-45"
	ZipStreet46   CompiledFilter = "zip-street-46"
	ZipStreet47   CompiledFilter = "zip-street-47"
	ZipStreet48   CompiledFilter = "zip-street-48"
	ZipStreet49   CompiledFilter = "zip-street-49"
	ZipStreet50   CompiledFilter = "zip-street-50"
	ZipStreet51   CompiledFilter = "zip-street-51"
	ZipStreet52   CompiledFilter = "zip-street-52"
	ZipStreet53   CompiledFilter = "zip-street-53"
	ZipStreet54   CompiledFilter = "zip-street-54"
	ZipStreet55   CompiledFilter = "zip-street-55"
	ZipStreet56   CompiledFilter = "zip-street-56"
	ZipStreet57   CompiledFilter = "zip-street-57"
	ZipStreet58   CompiledFilter = "zip-street-58"
	ZipStreet59   CompiledFilter = "zip-street-59"
	ZipStreet60   CompiledFilter = "zip-street-60"
	ZipStreet61   CompiledFilter = "zip-street-61"
	ZipStreet62   CompiledFilter = "zip-street-62"
	ZipStreet63   CompiledFilter = "zip-street-63"
	ZipStreet64   CompiledFilter = "zip-street-64"
	ZipStreet65   CompiledFilter = "zip-street-65"
	ZipStreet66   CompiledFilter = "zip-street-66"
	ZipStreet67   CompiledFilter = "zip-street-67"
	ZipStreet68   CompiledFilter = "zip-street-68"
	ZipStreet69   CompiledFilter = "zip-street-69"
	ZipStreet70   CompiledFilter = "zip-street-70"
	ZipStreet71   CompiledFilter = "zip-street-71"
	ZipStreet72   CompiledFilter = "zip-street-72"
	ZipStreet73   CompiledFilter = "zip-street-73"
	ZipStreet74   CompiledFilter = "zip-street-74"
	ZipStreet75   CompiledFilter = "zip-street-75"
	ZipStreet76   CompiledFilter = "zip-street-76"
	ZipStreet77   CompiledFilter = "zip-street-77"
	ZipStreet78   CompiledFilter = "zip-street-78"
	ZipStreet79   CompiledFilter = "zip-street-79"
	ZipStreet80   CompiledFilter = "zip-street-80"
	ZipStreet81   CompiledFilter = "zip-street-81"
	ZipStreet82   CompiledFilter = "zip-street-82"
	ZipStreet83   CompiledFilter = "zip-street-83"
	ZipStreet84   CompiledFilter = "zip-street-84"
	ZipStreet85   CompiledFilter = "zip-street-85"
	ZipStreet86   CompiledFilter = "zip-street-86"
	ZipStreet87   CompiledFilter = "zip-street-87"
	ZipStreet88   CompiledFilter = "zip-street-88"
	ZipStreet89   CompiledFilter = "zip-street-89"
	ZipStreet90   CompiledFilter = "zip-street-90"
	ZipStreet91   CompiledFilter = "zip-street-91"
	ZipStreet92   CompiledFilter = "zip-street-92"
	ZipStreet93   CompiledFilter = "zip-street-93"
	ZipStreet94   CompiledFilter = "zip-street-94"
	ZipStreet95   CompiledFilter = "zip-street-95"
	ZipStreet96   CompiledFilter = "zip-street-96"
	ZipStreet97   CompiledFilter = "zip-street-97"
	ZipStreet98   CompiledFilter = "zip-street-98"
	ZipStreet99   CompiledFilter = "zip-street-99"
)

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
	var filter bloom.BloomFilter
	var buf *bytes.Buffer
	switch name {
	case ZipCity:
		buf = bytes.NewBuffer(RawZipCityFilterBytes)
	case CityStreet:
		buf = bytes.NewBuffer(RawCityStreetFilterBytes)
	case ZipStreet00:
		buf = bytes.NewBuffer(RawZipStreet00FilterBytes)
	case ZipStreet01:
		buf = bytes.NewBuffer(RawZipStreet01FilterBytes)
	case ZipStreet02:
		buf = bytes.NewBuffer(RawZipStreet02FilterBytes)
	case ZipStreet03:
		buf = bytes.NewBuffer(RawZipStreet03FilterBytes)
	case ZipStreet04:
		buf = bytes.NewBuffer(RawZipStreet04FilterBytes)
	case ZipStreet05:
		buf = bytes.NewBuffer(RawZipStreet05FilterBytes)
	case ZipStreet06:
		buf = bytes.NewBuffer(RawZipStreet06FilterBytes)
	case ZipStreet07:
		buf = bytes.NewBuffer(RawZipStreet07FilterBytes)
	case ZipStreet08:
		buf = bytes.NewBuffer(RawZipStreet08FilterBytes)
	case ZipStreet09:
		buf = bytes.NewBuffer(RawZipStreet09FilterBytes)
	case ZipStreet10:
		buf = bytes.NewBuffer(RawZipStreet10FilterBytes)
	case ZipStreet11:
		buf = bytes.NewBuffer(RawZipStreet11FilterBytes)
	case ZipStreet12:
		buf = bytes.NewBuffer(RawZipStreet12FilterBytes)
	case ZipStreet13:
		buf = bytes.NewBuffer(RawZipStreet13FilterBytes)
	case ZipStreet14:
		buf = bytes.NewBuffer(RawZipStreet14FilterBytes)
	case ZipStreet15:
		buf = bytes.NewBuffer(RawZipStreet15FilterBytes)
	case ZipStreet16:
		buf = bytes.NewBuffer(RawZipStreet16FilterBytes)
	case ZipStreet17:
		buf = bytes.NewBuffer(RawZipStreet17FilterBytes)
	case ZipStreet18:
		buf = bytes.NewBuffer(RawZipStreet18FilterBytes)
	case ZipStreet19:
		buf = bytes.NewBuffer(RawZipStreet19FilterBytes)
	case ZipStreet20:
		buf = bytes.NewBuffer(RawZipStreet20FilterBytes)
	case ZipStreet21:
		buf = bytes.NewBuffer(RawZipStreet21FilterBytes)
	case ZipStreet22:
		buf = bytes.NewBuffer(RawZipStreet22FilterBytes)
	case ZipStreet23:
		buf = bytes.NewBuffer(RawZipStreet23FilterBytes)
	case ZipStreet24:
		buf = bytes.NewBuffer(RawZipStreet24FilterBytes)
	case ZipStreet25:
		buf = bytes.NewBuffer(RawZipStreet25FilterBytes)
	case ZipStreet26:
		buf = bytes.NewBuffer(RawZipStreet26FilterBytes)
	case ZipStreet27:
		buf = bytes.NewBuffer(RawZipStreet27FilterBytes)
	case ZipStreet28:
		buf = bytes.NewBuffer(RawZipStreet28FilterBytes)
	case ZipStreet29:
		buf = bytes.NewBuffer(RawZipStreet29FilterBytes)
	case ZipStreet30:
		buf = bytes.NewBuffer(RawZipStreet30FilterBytes)
	case ZipStreet31:
		buf = bytes.NewBuffer(RawZipStreet31FilterBytes)
	case ZipStreet32:
		buf = bytes.NewBuffer(RawZipStreet32FilterBytes)
	case ZipStreet33:
		buf = bytes.NewBuffer(RawZipStreet33FilterBytes)
	case ZipStreet34:
		buf = bytes.NewBuffer(RawZipStreet34FilterBytes)
	case ZipStreet35:
		buf = bytes.NewBuffer(RawZipStreet35FilterBytes)
	case ZipStreet36:
		buf = bytes.NewBuffer(RawZipStreet36FilterBytes)
	case ZipStreet37:
		buf = bytes.NewBuffer(RawZipStreet37FilterBytes)
	case ZipStreet38:
		buf = bytes.NewBuffer(RawZipStreet38FilterBytes)
	case ZipStreet39:
		buf = bytes.NewBuffer(RawZipStreet39FilterBytes)
	case ZipStreet40:
		buf = bytes.NewBuffer(RawZipStreet40FilterBytes)
	case ZipStreet41:
		buf = bytes.NewBuffer(RawZipStreet41FilterBytes)
	case ZipStreet42:
		buf = bytes.NewBuffer(RawZipStreet42FilterBytes)
	case ZipStreet43:
		buf = bytes.NewBuffer(RawZipStreet43FilterBytes)
	case ZipStreet44:
		buf = bytes.NewBuffer(RawZipStreet44FilterBytes)
	case ZipStreet45:
		buf = bytes.NewBuffer(RawZipStreet45FilterBytes)
	case ZipStreet46:
		buf = bytes.NewBuffer(RawZipStreet46FilterBytes)
	case ZipStreet47:
		buf = bytes.NewBuffer(RawZipStreet47FilterBytes)
	case ZipStreet48:
		buf = bytes.NewBuffer(RawZipStreet48FilterBytes)
	case ZipStreet49:
		buf = bytes.NewBuffer(RawZipStreet49FilterBytes)
	case ZipStreet50:
		buf = bytes.NewBuffer(RawZipStreet50FilterBytes)
	case ZipStreet51:
		buf = bytes.NewBuffer(RawZipStreet51FilterBytes)
	case ZipStreet52:
		buf = bytes.NewBuffer(RawZipStreet52FilterBytes)
	case ZipStreet53:
		buf = bytes.NewBuffer(RawZipStreet53FilterBytes)
	case ZipStreet54:
		buf = bytes.NewBuffer(RawZipStreet54FilterBytes)
	case ZipStreet55:
		buf = bytes.NewBuffer(RawZipStreet55FilterBytes)
	case ZipStreet56:
		buf = bytes.NewBuffer(RawZipStreet56FilterBytes)
	case ZipStreet57:
		buf = bytes.NewBuffer(RawZipStreet57FilterBytes)
	case ZipStreet58:
		buf = bytes.NewBuffer(RawZipStreet58FilterBytes)
	case ZipStreet59:
		buf = bytes.NewBuffer(RawZipStreet59FilterBytes)
	case ZipStreet60:
		buf = bytes.NewBuffer(RawZipStreet60FilterBytes)
	case ZipStreet61:
		buf = bytes.NewBuffer(RawZipStreet61FilterBytes)
	case ZipStreet62:
		buf = bytes.NewBuffer(RawZipStreet62FilterBytes)
	case ZipStreet63:
		buf = bytes.NewBuffer(RawZipStreet63FilterBytes)
	case ZipStreet64:
		buf = bytes.NewBuffer(RawZipStreet64FilterBytes)
	case ZipStreet65:
		buf = bytes.NewBuffer(RawZipStreet65FilterBytes)
	case ZipStreet66:
		buf = bytes.NewBuffer(RawZipStreet66FilterBytes)
	case ZipStreet67:
		buf = bytes.NewBuffer(RawZipStreet67FilterBytes)
	case ZipStreet68:
		buf = bytes.NewBuffer(RawZipStreet68FilterBytes)
	case ZipStreet69:
		buf = bytes.NewBuffer(RawZipStreet69FilterBytes)
	case ZipStreet70:
		buf = bytes.NewBuffer(RawZipStreet70FilterBytes)
	case ZipStreet71:
		buf = bytes.NewBuffer(RawZipStreet71FilterBytes)
	case ZipStreet72:
		buf = bytes.NewBuffer(RawZipStreet72FilterBytes)
	case ZipStreet73:
		buf = bytes.NewBuffer(RawZipStreet73FilterBytes)
	case ZipStreet74:
		buf = bytes.NewBuffer(RawZipStreet74FilterBytes)
	case ZipStreet75:
		buf = bytes.NewBuffer(RawZipStreet75FilterBytes)
	case ZipStreet76:
		buf = bytes.NewBuffer(RawZipStreet76FilterBytes)
	case ZipStreet77:
		buf = bytes.NewBuffer(RawZipStreet77FilterBytes)
	case ZipStreet78:
		buf = bytes.NewBuffer(RawZipStreet78FilterBytes)
	case ZipStreet79:
		buf = bytes.NewBuffer(RawZipStreet79FilterBytes)
	case ZipStreet80:
		buf = bytes.NewBuffer(RawZipStreet80FilterBytes)
	case ZipStreet81:
		buf = bytes.NewBuffer(RawZipStreet81FilterBytes)
	case ZipStreet82:
		buf = bytes.NewBuffer(RawZipStreet82FilterBytes)
	case ZipStreet83:
		buf = bytes.NewBuffer(RawZipStreet83FilterBytes)
	case ZipStreet84:
		buf = bytes.NewBuffer(RawZipStreet84FilterBytes)
	case ZipStreet85:
		buf = bytes.NewBuffer(RawZipStreet85FilterBytes)
	case ZipStreet86:
		buf = bytes.NewBuffer(RawZipStreet86FilterBytes)
	case ZipStreet87:
		buf = bytes.NewBuffer(RawZipStreet87FilterBytes)
	case ZipStreet88:
		buf = bytes.NewBuffer(RawZipStreet88FilterBytes)
	case ZipStreet89:
		buf = bytes.NewBuffer(RawZipStreet89FilterBytes)
	case ZipStreet90:
		buf = bytes.NewBuffer(RawZipStreet90FilterBytes)
	case ZipStreet91:
		buf = bytes.NewBuffer(RawZipStreet91FilterBytes)
	case ZipStreet92:
		buf = bytes.NewBuffer(RawZipStreet92FilterBytes)
	case ZipStreet93:
		buf = bytes.NewBuffer(RawZipStreet93FilterBytes)
	case ZipStreet94:
		buf = bytes.NewBuffer(RawZipStreet94FilterBytes)
	case ZipStreet95:
		buf = bytes.NewBuffer(RawZipStreet95FilterBytes)
	case ZipStreet96:
		buf = bytes.NewBuffer(RawZipStreet96FilterBytes)
	case ZipStreet97:
		buf = bytes.NewBuffer(RawZipStreet97FilterBytes)
	case ZipStreet98:
		buf = bytes.NewBuffer(RawZipStreet98FilterBytes)
	case ZipStreet99:
		buf = bytes.NewBuffer(RawZipStreet99FilterBytes)
	default:
		return nil, fmt.Errorf("Unsupported compiled filter: %s", name)
	}
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&filter); err != nil {
		return nil, err
	}
	return &filter, nil
}


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
//go:embed zip-city.bin
var RawZipCityFilterBytes []byte

// RawCityStreetFilterBytes holds the pre-compiled city-street Bloom filter
//go:embed city-street.bin
var RawCityStreetFilterBytes []byte

