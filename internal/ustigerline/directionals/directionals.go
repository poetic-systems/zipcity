package directionals

import "maps"

type DirectionalInfo struct {
	Dir         string
	Short       string
	Spanish     bool
	Translation string
}

/*
Direction Abbreviation Spanish Translation

North N - -
South S - -
East E - -
West W - -
Northeast NE - -
Northwest NW - -
Southeast SE - -
Southwest SW - -
Norte N Y North
Sur S Y South
Este E Y East
Oeste O Y West
Noreste NE Y Northeast
Noroeste NO Y Northwest
Sudeste SE Y Southeast
Sudoeste SO Y Southwest
*/

var directionals = []DirectionalInfo{
	{"North", "N", false, ""},
	{"South", "S", false, ""},
	{"East", "E", false, ""},
	{"West", "W", false, ""},
	{"Northeast", "NE", false, ""},
	{"Northwest", "NW", false, ""},
	{"Southeast", "SE", false, ""},
	{"Southwest", "SW", false, ""},
	{"Norte", "N", true, "North"},
	{"Sur", "S", true, "South"},
	{"Este", "E", true, "East"},
	{"Oeste", "O", true, "West"},
	{"Noreste", "NE", true, "Northeast"},
	{"Noroeste", "NO", true, "Northwest"},
	{"Sudeste", "SE", true, "Southeast"},
	{"Sudoeste", "SO", true, "Southwest"},
}

var spanishFullMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range directionals {
		if v.Spanish {
			if !yield(v.Short, v.Dir) {
				return
			}
			if !yield(v.Dir, v.Dir) {
				return
			}
		}
	}
})

var spanishShortMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range directionals {
		if v.Spanish {
			if !yield(v.Dir, v.Short) {
				return
			}
			if !yield(v.Short, v.Short) {
				return
			}
		}
	}
})

var englishFullMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range directionals {
		if !v.Spanish {
			if !yield(v.Short, v.Dir) {
				return
			}
			if !yield(v.Dir, v.Dir) {
				return
			}
		}
	}
})

var englishShortMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range directionals {
		if !v.Spanish {
			if !yield(v.Short, v.Short) {
				return
			}
			if !yield(v.Dir, v.Short) {
				return
			}
		}
	}
})

func Expand(abrev string, isSpanish bool) string {
	if isSpanish {
		r, ok := spanishFullMap[abrev]
		if ok {
			return r
		}
		return abrev
	}

	r, ok := englishFullMap[abrev]
	if ok {
		return r
	}
	return abrev
}

func Abbreviate(full string, isSpanish bool) string {
	if isSpanish {
		r, ok := spanishShortMap[full]
		if ok {
			return r
		}
		return full
	}

	r, ok := englishShortMap[full]
	if ok {
		return r
	}
	return full
}
