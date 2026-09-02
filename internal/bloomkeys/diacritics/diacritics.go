package diacritics

import (
	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
)

func Fold(s string) (string, error) {
	result := godiacritics.Normalize(s)
	return result, nil
}
