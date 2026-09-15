package bloomkeys

import (
	"fmt"
	"strings"

	"github.com/poetic-systems/zipcity/internal/bloomkeys/diacritics"
)

func normalize(keypart string) string {
	return strings.ToUpper(diacritics.Fold(keypart))
}

func KeyZipStreet(zip, street string) string {
	return fmt.Sprintf("%s:%s", normalize(zip), normalize(street))
}

func KeyZipCity(zip, city string) string {
	return fmt.Sprintf("%s:%s", normalize(zip), normalize(city))
}

func KeyCityStateStreet(city, state, street string) string {
	return fmt.Sprintf("%s:%s:%s", normalize(city), normalize(state), normalize(street))
}
