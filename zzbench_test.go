package zipcity_test

import (
	"testing"

	"github.com/poetic-systems/zipcity"
)

// Simulates go-projectusat#194's finding: several candidate readings per
// address ask the exact same (zip, street) / (city, state, street) pair.
// repeatsPerQuery models that fan-out.
const repeatsPerQuery = 6

func BenchmarkMatchZipAndStreetRepeated(b *testing.B) {
	zc, err := zipcity.New(zipcity.WithFilterFS(filesFilterFS))
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
	zc, err := zipcity.New(zipcity.WithFilterFS(filesFilterFS))
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
