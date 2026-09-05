package featnames_test

import (
	"testing"

	"github.com/poetic-systems/zipcity/internal/ustigerline/featnames"
)

func TestPub28FeatureName(t *testing.T) {
	cases := []struct {
		In   map[string]any
		Want string
	}{
		{
			In: map[string]any{
				"FULLNAME":   "O Calle de la Cruz",
				"NAME":       "de la Cruz",
				"PREDIR":     22,
				"PREDIRABRV": "O",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     156,
				"PRETYPABRV": "Cll",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "W CALLE DE LA CRUZ",
		},
		{
			In: map[string]any{
				"FULLNAME":   "Pleasant Hill Rd",
				"NAME":       "Pleasant Hill",
				"PREDIR":     "",
				"PREDIRABRV": "",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     "",
				"PRETYPABRV": "",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     531,
				"SUFTYPABRV": "Rd",
			},
			Want: "PLEASANT HILL RD",
		},
		{
			In: map[string]any{
				"FULLNAME":   "W Fox Park Dr",
				"NAME":       "Fox Park",
				"PREDIR":     "14",
				"PREDIRABRV": "W",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     "",
				"PRETYPABRV": "",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     247,
				"SUFTYPABRV": "Dr",
			},
			Want: "W FOX PARK DR",
		},
		{
			In: map[string]any{
				"FULLNAME":   "W 9200 S",
				"NAME":       "9200",
				"PREDIR":     14,
				"PREDIRABRV": "W",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     "",
				"PRETYPABRV": "",
				"SUFDIR":     12,
				"SUFDIRABRV": "S",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "W 9200 S",
		},
		{
			In: map[string]any{
				"FULLNAME":   "Ó CALLE C",
				"NAME":       "Ó CALLE C",
				"PREDIR":     "",
				"PREDIRABRV": "",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     "",
				"PRETYPABRV": "",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "CALLE C",
		},
		{
			In: map[string]any{
				"FULLNAME":   "W Ave G",
				"NAME":       "G",
				"PREDIR":     "14",
				"PREDIRABRV": "W",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     124,
				"PRETYPABRV": "Ave",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "W AVENIDA G",
		},
		{
			In: map[string]any{
				"FULLNAME":   "Via Donato W",
				"NAME":       "Donato",
				"PREDIR":     "",
				"PREDIRABRV": "",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     655,
				"PRETYPABRV": "Via",
				"SUFDIR":     14,
				"SUFDIRABRV": "W",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "VIA DONATO W",
		},
		{
			In: map[string]any{
				"FULLNAME":   "W Cll del Ciervo",
				"NAME":       "del Ciervo",
				"PREDIR":     14,
				"PREDIRABRV": "W",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     156,
				"PRETYPABRV": "Cll",
				"SUFDIR":     "",
				"SUFDIRABRV": "",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "W CALLE DEL CIERVO",
		},
		{
			In: map[string]any{
				"FULLNAME":   "Via Montebello E",
				"NAME":       "Montebello",
				"PREDIR":     "",
				"PREDIRABRV": "",
				"PREQUAL":    "",
				"PREQUALABR": "",
				"PRETYP":     655,
				"PRETYPABRV": "Via",
				"SUFDIR":     13,
				"SUFDIRABRV": "E",
				"SUFQUAL":    "",
				"SUFQUALABR": "",
				"SUFTYP":     "",
				"SUFTYPABRV": "",
			},
			Want: "VIA MONTEBELLO E",
		},
	}

	for _, tc := range cases {
		out := featnames.Pub28FeatureName(tc.In)
		if out != tc.Want {
			t.Fatalf("Expected: '%s' Got: '%s' for %v", tc.Want, out, tc.In)
		}
	}
}
