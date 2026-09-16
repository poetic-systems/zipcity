package bloomkeys

import (
	"fmt"
	"maps"
	"strings"

	"github.com/poetic-systems/addresstables/cityabbreviations"
	"github.com/poetic-systems/zipcity/internal/bloomkeys/diacritics"
)

func normalize(keypart string) string {
	return strings.ToUpper(diacritics.Fold(keypart))
}

// Punctuation may be omitted from an address (Publication 28 §222, Project
// US@ p. 15), and a caller in spec form has omitted it. A hyphen or slash
// joins two words and is dropped as a space so they stay two words;
// everything else is dropped outright. WINSTON-SALEM is WINSTON SALEM, as the
// USPS City State file prints it, and LEE'S SUMMIT is LEES SUMMIT.
var punctuation = strings.NewReplacer("-", " ", "/", " ", ".", "", "'", "", ",", "", "(", "", ")", "")

var spelledOut = maps.Collect(func(yield func(string, string) bool) {
	for a := range cityabbreviations.All() {
		if !yield(a.Short, a.Full) {
			return
		}
	}
})

// City normalizes a city name to the form Publication 28 §223 wants it
// written, on top of what normalize does: punctuation dropped, and ST, STE,
// MT and FT spelled out. The sources abbreviate inconsistently — GeoNames has
// ST. ALBANS where TIGER has SAINT ALBANS, and 242 ZIP Codes carried both —
// while a caller in spec form has spelled the city out, so the key is built
// spelled out on both sides. See #41.
//
// A word is spelled out only when another word follows it: ST is SAINT at the
// head of ST LOUIS or in the middle of BAY ST LOUIS. No name in the data ends
// in one of these, and a trailing ST would as likely be a street.
//
// Exported so the ZIP-to-city table (#24) can hold the same names the filter
// answers to.
func City(name string) string {
	words := strings.Fields(punctuation.Replace(normalize(name)))
	for i := range len(words) - 1 {
		if full, ok := spelledOut[words[i]]; ok {
			words[i] = full
		}
	}
	return strings.Join(words, " ")
}

func KeyZipStreet(zip, street string) string {
	return fmt.Sprintf("%s:%s", normalize(zip), normalize(street))
}

func KeyZipCity(zip, city string) string {
	return fmt.Sprintf("%s:%s", normalize(zip), City(city))
}

func KeyCityStateStreet(city, state, street string) string {
	return fmt.Sprintf("%s:%s:%s", City(city), normalize(state), normalize(street))
}
