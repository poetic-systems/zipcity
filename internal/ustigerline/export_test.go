package ustigerline

import "path/filepath"

// FixturePrefix names the county the checked-in TIGER fixtures were cut from.
const FixturePrefix = "tl_2025_01001"

// UseFixtures points the readers at the checked-in fixtures for the duration
// of a test, and returns a function that puts the download cache back.
//
// The fixtures deliberately do not live in the cache. A download is skipped
// whenever the file is already there, so a trimmed archive sitting in
// data/us_census_tiger would silently stand in for the real one on any machine
// that ran a generation.
func UseFixtures() func() {
	previous := storagedir
	storagedir = filepath.Join("testdata", "us_census_tiger")
	return func() { storagedir = previous }
}
