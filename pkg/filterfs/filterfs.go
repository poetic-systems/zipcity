package filterfs

import "io/fs"

type InitFn func() (fs.FS, error)

func (init InitFn) PrepareFS() (fs.FS, error) {
	return init()
}

type InitFunc interface {
	PrepareFS() (fs.FS, error)
}
