package httpfs

import (
	"net/http"
	"github.com/rakyll/statik/fs"
)

type system struct {
	*fs.StatikFS
}

func System(fs *fs.StatikFS) system {
	return system{
		fs,
	}
}

func (fs system) Open(name string) (http.File, error) {
	return fs.StatikFS.Open(name)
}
