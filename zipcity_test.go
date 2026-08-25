package zipcity_test

import (
	"testing"

	"github.com/poetic-systems/zipcity"
)

func TestCheckZipAndCity(t *testing.T) {
	zip := "94553"
	city := "MARTINEZ"
	found, err := zipcity.CheckZipAndCity(zip, city)
	if err != nil {
		t.Fatalf("Error checking zip: '%s' and city: '%s': %s", zip, city, err)
	}

	if !found {
		t.Fatalf("Expected zip: '%s' and city: '%s' to be found", zip, city)
	}
}

func TestCheckZipAndStreet(t *testing.T) {
	zip := "94553"
	street := "BLUE RIDGE DR"

	found, err := zipcity.CheckZipAndStreet(zip, street)
	if err != nil {
		t.Fatalf("Error checking zip: '%s' and street: '%s': %s", zip, street, err)
	}

	if !found {
		t.Fatalf("Expected zip: '%s' and street: '%s' to be found", zip, street)
	}
}

func TestCheckCityStateAndStreet(t *testing.T) {
	city := "MARTINEZ"
	state := "CA"
	street := "BLUE RIDGE DR"

	found, err := zipcity.CheckCityStateAndStreet(city, state, street)
	if err != nil {
		t.Fatalf("Error checking city: '%s' state: '%s' and street: '%s': %s", city, state, street, err)
	}

	if !found {
		t.Fatalf("Expected city: '%s' state: '%s' and street: '%s' to be found", city, state, street)
	}
}
