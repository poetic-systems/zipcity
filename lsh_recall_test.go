package zipcity_test

import (
	"testing"
	"time"

	"github.com/poetic-systems/zipcity"
	"github.com/poetic-systems/zipcity/internal/cityabbrev"
)

// TestLSHRecall compares the LSH index to the linear scan on synthetic
// abbreviations of every 20th city in a state: last letter dropped, one
// interior letter dropped, two adjacent letters swapped.
func TestLSHRecall(t *testing.T) {
	zc, err := zipcity.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"CA", "TX", "NY"} {
		names := zc.CityNames(state)
		start := time.Now()
		ix := cityabbrev.NewIndex(names)
		build := time.Since(start)
		var total, linear, approx int
		var linT, lshT time.Duration
		for i := 0; i < len(names); i += 20 {
			n := names[i]
			if len(n) < 6 {
				continue
			}
			for _, q := range []string{n[:len(n)-1], n[:3] + n[4:], n[:2] + n[3:4] + n[2:3] + n[4:]} {
				total++
				s := time.Now()
				a := cityabbrev.Match(q, names)
				linT += time.Since(s)
				s = time.Now()
				b := ix.Query(q)
				lshT += time.Since(s)
				linear += hit(a, n)
				approx += hit(b, n)
			}
		}
		t.Logf("%s: %d cities, build %v; %d queries; linear recall %d, lsh recall %d; linear %v/q, lsh %v/q",
			state, len(names), build, total, linear, approx, linT/time.Duration(total), lshT/time.Duration(total))
	}
}

func hit(cs []cityabbrev.Candidate, want string) int {
	for _, c := range cs {
		if c.City == want {
			return 1
		}
	}
	return 0
}
