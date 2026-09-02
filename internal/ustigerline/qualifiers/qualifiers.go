package qualifiers

import (
	"fmt"
	"maps"
)

/*
QualifierCode	ExpandedFullText	DisplayNameAbbreviation	PrefixQualifier	SuffixQualifier
11 Access Acc N Y
12 Alternate Alt Y Y
13 Business Bus Y Y
14 Bypass Byp Y Y
15 Connector Con N Y
16 Extended Exd Y Y
17 Extension Exn N Y
18 Historic Hst Y N
19 Loop Lp Y Y
20 Old Old Y N
21 Private Pvt Y Y
22 Public Pub Y Y
23 Scenic Scn N Y
24 Spur Spr Y Y
25 Ramp Rmp N Y
26 Underpass Unp N Y
27 Overpass Ovp N Y
*/

type QualifierInfo struct {
	Code   string
	Full   string
	Short  string
	Prefix bool
	Suffix bool
}

var qualifiers = []QualifierInfo{
	{"11", "Access", "Acc", false, true},
	{"12", "Alternate", "Alt", true, true},
	{"13", "Business", "Bus", true, true},
	{"14", "Bypass", "Byp", true, true},
	{"15", "Connector", "Con", false, true},
	{"16", "Extended", "Exd", true, true},
	{"17", "Extension", "Exn", false, true},
	{"18", "Historic", "Hst", true, false},
	{"19", "Loop", "Lp", true, true},
	{"20", "Old", "Old", true, false},
	{"21", "Private", "Pvt", true, true},
	{"22", "Public", "Pub", true, true},
	{"23", "Scenic", "Scn", false, true},
	{"24", "Spur", "Spr", true, true},
	{"25", "Ramp", "Rmp", false, true},
	{"26", "Underpass", "Unp", false, true},
	{"27", "Overpass", "Ovp", false, true},
}

var qualifierMap = maps.Collect(func(yield func(string, QualifierInfo) bool) {
	for _, v := range qualifiers {
		if !yield(v.Code, v) {
			return
		}
	}
})

func Info(code string) (*QualifierInfo, error) {
	q, ok := qualifierMap[code]
	if !ok {
		return nil, fmt.Errorf("'%s' is not a recognized qualifier code", code)
	}

	return &QualifierInfo{
		Code:   q.Code,
		Full:   q.Full,
		Short:  q.Short,
		Prefix: q.Prefix,
		Suffix: q.Suffix,
	}, nil
}
