package fieldutil

import (
	"fmt"
	"strings"
)

func AsString(input interface{}) string {
	s, ok := input.(string)
	if ok {
		return fmt.Sprintf("%s", s)
	}

	d, ok := input.(int)
	if ok {
		return fmt.Sprintf("%d", d)
	}

	return fmt.Sprintf("%v", input)
}

func JoinNonEmpty(elems []string, sep string) string {
	var filtered []string
	for _, str := range elems {
		if len(str) > 0 { // Skip empty strings
			filtered = append(filtered, str)
		}
	}
	return strings.Join(filtered, sep)
}
