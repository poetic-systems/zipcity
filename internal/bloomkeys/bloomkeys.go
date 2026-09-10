package bloomkeys

import (
	"fmt"
	"strings"

	"github.com/poetic-systems/zipcity/internal/bloomkeys/diacritics"
)

func normalize(keypart string) (string, error) {
	upper := strings.ToUpper(keypart)
	folded, err := diacritics.Fold(upper)
	if err != nil {
		return "", fmt.Errorf("unable to create key part: %w", err)
	}
	return folded, nil
}

func KeyZipStreet(zip, street string) (string, error) {
	nzip, err := normalize(zip)
	if err != nil {
		return "", err
	}

	nstreet, err := normalize(street)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", nzip, nstreet), nil
}

func KeyZipCity(zip, city string) (string, error) {
	nzip, err := normalize(zip)
	if err != nil {
		return "", err
	}

	ncity, err := normalize(city)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", nzip, ncity), nil
}

func KeyCityStateStreet(city, state, street string) (string, error) {
	ncity, err := normalize(city)
	if err != nil {
		return "", err
	}

	nstate, err := normalize(state)
	if err != nil {
		return "", err
	}

	nstreet, err := normalize(street)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s:%s", ncity, nstate, nstreet), nil
}
