package benchmark

import (
	"testing"

	"github.com/poetic-systems/zipcity"
)

// repeatsPerQuery simulates go-projectusat#194's finding: several candidate
// readings per address ask the exact same (zip, street) / (city, state,
// street) pair.
const repeatsPerQuery = 6

var testData = []struct {
	Street string
	City   string
	State  string
	Zip    string
}{
	{"BLUE RIDGE DR", "MARTINEZ", "CA", "94553"},
	{"W 200 N", "NORTH SALT LAKE", "UT", "84054"},
	{"CALLE LOIZA", "San Juan", "PR", "00909"},
}

func BenchmarkMatchZipAndStreetRepeated(b *testing.B) {
	zc, err := zipcity.New()
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d := testData[i%len(testData)]
		for r := 0; r < repeatsPerQuery; r++ {
			if _, err := zc.MatchZipAndStreet(d.Zip, d.Street); err != nil {
				b.Fatalf("MatchZipAndStreet: %v", err)
			}
		}
	}
}

func BenchmarkMatchCityStateAndStreetRepeated(b *testing.B) {
	zc, err := zipcity.New()
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d := testData[i%len(testData)]
		for r := 0; r < repeatsPerQuery; r++ {
			if _, err := zc.MatchCityStateAndStreet(d.City, d.State, d.Street); err != nil {
				b.Fatalf("MatchCityStateAndStreet: %v", err)
			}
		}
	}
}
