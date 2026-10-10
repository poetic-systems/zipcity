package filterfs

import "io/fs"

// InitFunc and the Init interface mirrors the http.HandlerFunc adapter pattern
// to allow functions to be used to instantiate a fliter filesystem
type InitFunc func() (fs.FS, error)

func (init InitFunc) PrepareFS() (fs.FS, error) {
	return init()
}

type Init interface {
	PrepareFS() (fs.FS, error)
}
