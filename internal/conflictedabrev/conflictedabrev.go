package conflictedabrev

import "maps"

type SourceValue struct {
	Source string
	Value  string
}

type KnownConflict struct {
	Abrev    string
	Possible []SourceValue
}

var knownConflicts = []KnownConflict{
	{
		Abrev: "CT",
		Possible: []SourceValue{
			{
				Source: "StreetType",
				Value:  "COURT",
			},
			{
				Source: "Region",
				Value:  "CONNECTICUT",
			},
		},
	},
	{
		Abrev: "MT",
		Possible: []SourceValue{
			{"CityName", "MOUNT"},
			{"Region", "MONTANA"},
		},
	},
	{
		Abrev: "NE",
		Possible: []SourceValue{
			{"Directional", "NORTHEAST"},
			{"Region", "NEBRASKA"},
		},
	},
	{
		Abrev: "ST",
		Possible: []SourceValue{
			{"CityName", "SAINT"},
			{"StreetType", "STREET"},
			{"Other", "STATE"},
		},
	},
	{
		Abrev: "DR",
		Possible: []SourceValue{
			{"StreetType", "DRIVE"},
			{"Other", "DOCTOR"},
		},
	},
}

var conflictMap = maps.Collect(func(yield func(string, KnownConflict) bool) {
	for _, kc := range knownConflicts {
		if !yield(kc.Abrev, kc) {
			return
		}
		for _, sv := range kc.Possible {
			if !yield(sv.Value, kc) {
				return
			}
		}
	}
})

func Info(token string) *KnownConflict {
	kc, ok := conflictMap[token]
	if !ok {
		return nil
	}

	return &KnownConflict{
		Abrev:    kc.Abrev,
		Possible: kc.Possible[:],
	}
}
