package fs

import (
	"net/http"
	"os"
	"path/filepath"
)

type chdirFileSystem struct {
	prefix string
	system http.FileSystem
}

func (c chdirFileSystem) Open(name string) (http.File, error) {
	name = filepath.Clean(name)
	return c.system.Open(filepath.Join(c.prefix, name))
}

func Chdir(httpFS http.FileSystem, dir string) (http.FileSystem, error) {
	switch fs := httpFS.(type) {
	case *chdirFileSystem:
		return chdirOnChdirFS(fs, dir)
	default:
		return chdirOnOtherFS(fs, dir)
	}
}

func chdirOnChdirFS(fs *chdirFileSystem, dir string) (*chdirFileSystem, error) {
	if !isExistingDir(fs, dir) {
		return nil, os.ErrNotExist
	}
	prefix := filepath.Join(fs.prefix, dir)
	newFS := &chdirFileSystem{
		prefix: filepath.Clean(prefix),
		system: fs.system,
	}
	return newFS, nil
}

func chdirOnOtherFS(fs http.FileSystem, dir string) (http.FileSystem, error) {
	if !isExistingDir(fs, dir) {
		return nil, os.ErrNotExist
	}
	return &chdirFileSystem{
		prefix: dir,
		system: fs,
	}, nil
}

func isExistingDir(fs http.FileSystem, dir string) bool {
	f, err := fs.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.IsDir()
}
