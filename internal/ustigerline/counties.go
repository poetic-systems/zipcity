package ustigerline

import "iter"

// County is the street sides read for one county prefix, or the error that
// stopped them being read.
type County struct {
	Prefix string
	Sides  map[string]*StreetSide
	Err    error
}

// ReadCounties reads the street sides of every prefix, up to workers at a
// time, and yields them in the order the prefixes were given. Counties are
// read independently of one another, so how many are in flight changes only
// how long the whole takes, never what any one of them says; yielding in
// order keeps a generation the same however many workers it had. At most
// workers counties are read ahead of the one being yielded, which is what
// bounds the memory in flight. See #46.
func ReadCounties(prefixes []string, workers int) iter.Seq[County] {
	return func(yield func(County) bool) {
		results := make([]chan County, len(prefixes))
		started := 0
		for i := range prefixes {
			for started < len(prefixes) && started < i+workers {
				results[started] = make(chan County, 1)
				go func(i int) {
					sides, err := ReadStreetSides(prefixes[i])
					results[i] <- County{prefixes[i], sides, err}
				}(started)
				started++
			}
			if !yield(<-results[i]) {
				return
			}
		}
	}
}
