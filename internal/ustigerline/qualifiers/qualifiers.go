// Package qualifiers reads the PREQUAL and SUFQUAL fields of a TIGER
// FEATNAMES record.
//
// The rows are Appendix C of the Census Bureau's technical documentation and
// live in addresstables; what is here is the lookup a reader of those two
// fields needs, because a record gives you the code and nothing else.
package qualifiers

import (
	"fmt"
	"maps"

	table "github.com/poetic-systems/addresstables/ustigerfile/qualifiers"
)

// QualifierInfo is one Appendix C row. Rows are uppercase, as the whole of
// addresstables is; the published table prints them in title case.
type QualifierInfo = table.Qualifier

var qualifierMap = maps.Collect(func(yield func(string, QualifierInfo) bool) {
	for q := range table.All() {
		if !yield(q.Code, q) {
			return
		}
	}
})

func Info(code string) (*QualifierInfo, error) {
	q, ok := qualifierMap[code]
	if !ok {
		return nil, fmt.Errorf("'%s' is not a recognized qualifier code", code)
	}

	return &q, nil
}
