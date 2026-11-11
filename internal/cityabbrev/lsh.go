package cityabbrev

import (
	"hash/fnv"
	"math"

	"github.com/ekzhu/lsh"
)

// Index is an approximate nearest-neighbour index over one state's city
// names: each name becomes a vector of hashed character bigrams (and the
// first-letter marker), ekzhu/lsh's Euclidean LSH buckets the vectors, and
// the survivors are reranked with Damerau-Levenshtein. It finds typos and
// truncations (GLNDALE, GLENDAL) in sub-linear time. It does not find
// initialisms or heavy vowel loss (NY, BKLYN): those share almost no bigrams
// with the full name, which is why Match's rules still exist.
type Index struct{ idx *lsh.BasicLsh }

const (
	dim    = 128
	tables = 16 // l: hash tables, more = better recall
	hashes = 3  // m: hashes per table, more = fewer false candidates
	width  = 1.0
)

// NewIndex builds an index over names.
func NewIndex(names []string) *Index {
	idx := lsh.NewBasicLsh(dim, tables, hashes, width)
	for _, n := range names {
		idx.Insert(vector(n), n)
	}
	return &Index{idx}
}

// Query returns the indexed names within MinFuzzy of abbr, best first.
func (ix *Index) Query(abbr string) []Candidate {
	var cands []string
	cands = ix.idx.Query(vector(abbr))
	return Match(abbr, cands)
}

func vector(s string) lsh.Point {
	v := make(lsh.Point, dim)
	padded := "^" + s + "$"
	for i := 0; i+2 <= len(padded); i++ {
		h := fnv.New32a()
		h.Write([]byte(padded[i : i+2]))
		v[h.Sum32()%dim]++
	}
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	for i := range v {
		v[i] /= norm
	}
	return v
}
