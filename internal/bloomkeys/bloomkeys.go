package bloomkeys

import "fmt"

func KeyZipStreet(zip, street string) string {
	return fmt.Sprintf("%s:%s", zip, street)
}

func KeyZipCity(zip, city string) string {
	return fmt.Sprintf("%s:%s", zip, city)
}

func KeyCityStateStreet(city, state, street string) string {
	return fmt.Sprintf("%s:%s:%s", city, state, street)
}
