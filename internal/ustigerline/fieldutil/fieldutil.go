package fieldutil

import "fmt"

func AsString(input interface{}) string {
	s, ok := input.(string)
	if ok {
		// fmt.Printf("Formatting %v as string\n", input)
		return fmt.Sprintf("%s", s)
	}

	d, ok := input.(int)
	if ok {
		// fmt.Printf("Formatting %v as int\n", input)
		return fmt.Sprintf("%d", d)
	}

	// fmt.Printf("Formatting %v using Sprintf(v)\n", input)
	return fmt.Sprintf("%v", input)
}
