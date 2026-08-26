package bloomkeys

import (
	"fmt"
	"strings"
)

func normalize(keypart string) string {
	return strings.ToUpper(keypart)
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
