package zipcities

import (
	"compress/gzip"
	"fmt"
	"io"
)

// Size is what a generated table costs: the bytes it is written as, and the
// bytes it would be as gzip, which is what the two candidate formats for a
// committed artifact would weigh.
type Size struct {
	Zips    int
	Names   int
	Bytes   int
	Gzipped int
}

func (s Size) String() string {
	return fmt.Sprintf(
		"zip-city names: %d ZIP Codes, %d names, %d bytes (%d gzipped)",
		s.Zips, s.Names, s.Bytes, s.Gzipped,
	)
}

// Measure encodes the table twice, plain and gzipped, counting bytes and
// keeping neither.
//
// The number decides whether the table ships, so it has to come from a full
// generation rather than from a count of the counties a working tree happens
// to cache. See poetic-systems/zipcity#17.
func Measure(t Table) (Size, error) {
	size := Size{Zips: len(t)}
	for _, names := range t {
		size.Names += len(names)
	}

	plain := &counter{}
	if err := Encode(plain, t); err != nil {
		return Size{}, err
	}
	size.Bytes = plain.n

	compressed := &counter{}
	gz := gzip.NewWriter(compressed)
	if err := Encode(gz, t); err != nil {
		return Size{}, err
	}
	if err := gz.Close(); err != nil {
		return Size{}, err
	}
	size.Gzipped = compressed.n

	return size, nil
}

// counter is an io.Writer that keeps the count and drops the bytes.
type counter struct{ n int }

func (c *counter) Write(p []byte) (int, error) {
	c.n += len(p)
	return len(p), nil
}

var _ io.Writer = (*counter)(nil)
