package diacritics_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/bloomkeys/diacritics"
)

func TestFold(t *testing.T) {
	input := "Açaí, résumé, Ötzi, Šđčćž"
	expected := "Acai, resume, Otzi, Sdccz"

	output, err := diacritics.Fold(input)
	if err != nil {
		t.Fatalf("Error folding diacritics: %s", err)
	}

	if output != expected {
		t.Fatalf("Expected: '%s' Got: '%s'", expected, output)
	}
}
