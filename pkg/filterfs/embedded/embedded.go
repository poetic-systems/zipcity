package embedded

import (
	"io/fs"

	"github.com/poetic-systems/zipcity/generated/embedded_filter"
)

func PrepareFS() (fs.FS, error) {
	return embedded_filter.PrepareFS()
}
