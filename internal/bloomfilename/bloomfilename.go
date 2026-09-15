package bloomfilename

import (
	"fmt"
)

func Filename(identifier string) string {
	return fmt.Sprintf("%s.bin", identifier)
}
