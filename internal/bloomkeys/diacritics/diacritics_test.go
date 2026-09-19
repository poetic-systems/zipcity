package diacritics_test

import (
	"strings"
	"testing"
	"unicode"

	appendixa "github.com/poetic-systems/addresstables/diacritics"
	"github.com/poetic-systems/zipcity/internal/bloomkeys/diacritics"
)

// TestFoldMatchesAppendixA checks every row of the appendix, including the
// ones Unicode cannot decompose (Æ, Ð, Þ, ß, Œ), which are the rows a
// mark-stripping fold gets wrong.
func TestFoldMatchesAppendixA(t *testing.T) {
	for d := range appendixa.All() {
		want := d.Substitute
		if unicode.IsUpper(d.Rune) {
			want = strings.ToUpper(want)
		}
		if got := diacritics.Fold(string(d.Rune)); got != want {
			t.Errorf("%q (%U): want %q got %q", d.Rune, d.Rune, want, got)
		}
	}
}

func TestFold(t *testing.T) {
	cases := []struct{ In, Want string }{
		{"Açaí, résumé, Ötzi", "Acai, resume, Otzi"},
		{"CALLEJÓN MONSO MÉNDEZ", "CALLEJON MONSO MENDEZ"},
		// Not in Appendix A, but decomposable. The only such name in the 703
		// counties measured.
		{"Sčilíp CDP", "Scilip CDP"},
		{"Straße", "Strase"},
		{"ÆØÐÞŒ", "AOEPO"},
	}
	for _, tc := range cases {
		if got := diacritics.Fold(tc.In); got != tc.Want {
			t.Errorf("Fold(%q): want %q got %q", tc.In, tc.Want, got)
		}
	}
}

func BenchmarkFold(b *testing.B) {
	for _, tc := range []struct{ Name, In string }{
		{"ascii", "CALLEJON MONSO MENDEZ"},
		{"folded", "CALLEJÓN MONSO MÉNDEZ"},
	} {
		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				diacritics.Fold(tc.In)
			}
		})
	}
}
