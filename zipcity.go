package zipcity

import (
	"fmt"
	"regexp"

	"github.com/poetic-systems/zipcity/generated/compiled_filter"
	"github.com/poetic-systems/zipcity/internal/bloomkeys"
)

var zip5pattern = regexp.MustCompile(`^\d{5}$`)

func CheckZipAndCity(zip, city string) (bool, error) {
	if !zip5pattern.MatchString(zip) {
		return false, fmt.Errorf("5-digit zip code required")
	}

	if len(city) < 1 {
		return false, fmt.Errorf("city required")
	}

	f, err := compiled_filter.LoadFilter(compiled_filter.ZipCity)
	if err != nil {
		return false, fmt.Errorf("Unable to load bloom filter: %w", err)
	}

	key, err := bloomkeys.KeyZipCity(zip, city)
	if err != nil {
		return false, err
	}

	return f.TestString(key), nil
}

func CheckZipAndStreet(zip, street string) (bool, error) {
	if !zip5pattern.MatchString(zip) {
		return false, fmt.Errorf("5-digit zip code required")
	}

	if len(street) < 1 {
		return false, fmt.Errorf("street required")
	}

	filterId, err := compiled_filter.ZipStreetFilterForZip(zip)
	if err != nil {
		return false, fmt.Errorf("Unable to identify bloom filter for zip: %w", err)
	}

	f, err := compiled_filter.LoadFilter(filterId)
	if err != nil {
		return false, fmt.Errorf("Unable to load bloom filter: %w", err)
	}

	key, err := bloomkeys.KeyZipStreet(zip, street)
	if err != nil {
		return false, err
	}

	return f.TestString(key), nil
}

func CheckCityStateAndStreet(city, state, street string) (bool, error) {
	if len(city) < 1 {
		return false, fmt.Errorf("city required")
	}

	if len(state) != 2 {
		return false, fmt.Errorf("2-letter state abbreviation required")
	}

	if len(street) < 1 {
		return false, fmt.Errorf("street required")
	}

	f, err := compiled_filter.LoadFilter(compiled_filter.CityStreet)
	if err != nil {
		return false, fmt.Errorf("Unable to load bloom filter: %w", err)
	}

	key, err := bloomkeys.KeyCityStateStreet(city, state, street)
	if err != nil {
		return false, err
	}

	return f.TestString(key), nil
}
