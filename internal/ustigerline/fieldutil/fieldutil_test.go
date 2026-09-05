package fieldutil_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline/fieldutil"
)

func TestJoinNonEmpty(t *testing.T) {
	in := []string{
		"",
		"this",
		"",
		"is a",
		"test",
		"",
	}
	want := "this is a test"

	out := fieldutil.JoinNonEmpty(in, " ")
	if out != want {
		t.Fatalf("Expected: '%s' Got: '%s' for %v", want, out, in)
	}
}

func TestAsString(t *testing.T) {
	cases := []struct {
		In   any
		Want string
	}{
		{In: "string", Want: "string"},
		{In: 1234, Want: "1234"},
		{In: 5.5, Want: "5.5"},
		{In: []string{"hi"}, Want: "[hi]"},
	}

	for _, tc := range cases {
		out := fieldutil.AsString(tc.In)
		if out != tc.Want {
			t.Fatalf("Expected: '%s' Got: '%s' for %v", tc.Want, out, tc.In)
		}
	}
}
