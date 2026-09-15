// Package diacritics folds a name to ASCII the way Project US@ Appendix A
// says to, so a key built at generation time and a key built from spec-form
// input agree.
//
// Appendix A lists the Latin-1 letters and a few others; anything it does not
// list but Unicode can decompose (č in Sčilíp, Lake County MT) has its marks
// stripped instead, so no key is left holding a non-ASCII rune the caller
// cannot type.
package diacritics

import (
	"maps"
	"unicode"

	"github.com/poetic-systems/addresstables/diacritics"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var substitutes = maps.Collect(func(yield func(rune, rune) bool) {
	for d := range diacritics.All() {
		if !yield(d.Rune, rune(d.Substitute[0])) {
			return
		}
	}
})

// The appendix prints every substitute in lowercase, even for a capital row,
// so the case comes from the input rune.
func substitute(r rune) rune {
	s, ok := substitutes[r]
	if !ok {
		return r
	}
	if unicode.IsUpper(r) {
		return unicode.ToUpper(s)
	}
	return s
}

// Fold substitutes every Appendix A character with the letter the appendix
// prints for it, keeping the input's case, then strips combining marks from
// whatever is left.
func Fold(s string) string {
	t := transform.Chain(runes.Map(substitute), norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	// None of the chained transformers can fail: invalid UTF-8 is replaced with
	// U+FFFD rather than reported.
	folded, _, _ := transform.String(t, s)
	return folded
}
