package ustigerline

import (
	"maps"
	"reflect"
	"slices"
	"testing"
)

func areas(codes ...string) map[string]bool {
	out := make(map[string]bool, len(codes))
	for _, c := range codes {
		out[c] = true
	}

	return out
}

func TestAbsentSources(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   map[string]map[string]bool
		want AbsentSources
	}{
		{
			name: "every type published for every area",
			in: map[string]map[string]bool{
				"featnames": areas("02100", "60010"),
				"addr":      areas("02100", "60010"),
			},
			want: AbsentSources{},
		},
		{
			// The 2025 release: American Samoa's county equivalents are
			// described by every county file except ADDR.
			name: "one type missing for one area",
			in: map[string]map[string]bool{
				"featnames": areas("02100", "60010"),
				"edges":     areas("02100", "60010"),
				"faces":     areas("02100", "60010"),
				"addr":      areas("02100"),
			},
			want: AbsentSources{"60010": {"addr"}},
		},
		{
			name: "an area missing more than one type",
			in: map[string]map[string]bool{
				"featnames": areas("02100", "69085"),
				"edges":     areas("02100"),
				"addr":      areas("02100"),
			},
			want: AbsentSources{"69085": {"addr", "edges"}},
		},
		{
			// A county equivalent is not missing the national state file, and
			// a state is not missing a county file.
			name: "granularities do not contaminate each other",
			in: map[string]map[string]bool{
				"featnames": areas("02100", "60010"),
				"addr":      areas("02100", "60010"),
				"place":     areas("02", "60"),
				"state":     areas("us"),
			},
			want: AbsentSources{},
		},
		{
			name: "absence within a granularity is still reported",
			in: map[string]map[string]bool{
				"featnames": areas("02100", "60010"),
				"addr":      areas("02100"),
				"place":     areas("02", "60"),
				"state":     areas("us"),
			},
			want: AbsentSources{"60010": {"addr"}},
		},
		{
			name: "nothing published at all",
			in:   map[string]map[string]bool{},
			want: AbsentSources{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := absentSources(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("absentSources() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The report is data a generated package is written from, so two runs over the
// same index must produce the same bytes. Map iteration order is the risk.
func TestAbsentSourcesIsOrdered(t *testing.T) {
	in := map[string]map[string]bool{
		"featnames": areas("02100", "69085"),
		"edges":     areas("02100"),
		"addr":      areas("02100"),
		"faces":     areas("02100"),
	}

	first := absentSources(in)
	for range 20 {
		if got := absentSources(in); !reflect.DeepEqual(got, first) {
			t.Fatalf("absentSources() = %v on a later call, %v on the first", got, first)
		}
	}

	for area, filetypes := range first {
		if !slices.IsSorted(filetypes) {
			t.Errorf("%s: %v is not sorted", area, filetypes)
		}
	}
	if got := slices.Sorted(maps.Keys(first)); !reflect.DeepEqual(got, []string{"69085"}) {
		t.Errorf("areas = %v, want [69085]", got)
	}
}

func TestAreaOf(t *testing.T) {
	for _, tc := range []struct{ basename, suffix, want string }{
		{"tl_2025_02100_addr", "_addr", "02100"},
		{"tl_2025_02100_featnames", "_featnames", "02100"},
		{"tl_2025_02_place", "_place", "02"},
		{"tl_2025_us_state", "_state", "us"},
	} {
		if got := areaOf(tc.basename, tc.suffix); got != tc.want {
			t.Errorf("areaOf(%q, %q) = %q, want %q", tc.basename, tc.suffix, got, tc.want)
		}
	}
}

// The policy the Set field used to carry: ADDR may be absent, the county trio
// may not. It is now one statement — MayBeAbsent — read here.
func TestAbsentSourcesUnexpected(t *testing.T) {
	optional := optionalTypes(allRequiredTigerfiles())

	for _, tc := range []struct {
		name string
		in   AbsentSources
		want []string
	}{
		{
			name: "the eight ADDR-absent county equivalents are allowed",
			in: AbsentSources{
				"60010": {"addr"}, "60020": {"addr"}, "60030": {"addr"},
				"60040": {"addr"}, "60050": {"addr"}, "69085": {"addr"},
				"69100": {"addr"}, "69120": {"addr"},
			},
			want: []string{},
		},
		{
			name: "a county short a required type is not",
			in:   AbsentSources{"02100": {"edges"}},
			want: []string{"02100: edges"},
		},
		{
			name: "an allowed absence does not excuse the others on the same area",
			in:   AbsentSources{"69085": {"addr", "faces", "featnames"}},
			want: []string{"69085: faces", "69085: featnames"},
		},
		{
			name: "a state missing its place file is reported like any other",
			in:   AbsentSources{"60": {"place"}},
			want: []string{"60: place"},
		},
		{
			name: "nothing absent",
			in:   AbsentSources{},
			want: []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.unexpected(optional); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("unexpected() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOptionalTypes(t *testing.T) {
	got := optionalTypes(allRequiredTigerfiles())
	if !reflect.DeepEqual(got, map[string]bool{"addr": true}) {
		t.Errorf("optionalTypes() = %v, want map[addr:true]", got)
	}
}

func TestCountyPrefixes(t *testing.T) {
	idx := &tigerfileIndex{
		release: "tl_2025",
		areasByType: map[string]map[string]bool{
			"featnames": areas("02100", "01001", "60010"),
			"addr":      areas("02100", "01001"),
			"place":     areas("02", "01", "60"),
			"state":     areas("us"),
		},
	}

	// Only the five digit codes, each county named once however many of its
	// files the index lists, in an order two runs agree on.
	want := []string{"tl_2025_01001", "tl_2025_02100", "tl_2025_60010"}
	if got := idx.countyPrefixes(); !reflect.DeepEqual(got, want) {
		t.Errorf("countyPrefixes() = %v, want %v", got, want)
	}
}
