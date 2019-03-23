// Package ziptree contains code to zip a directory tree and write it out.
package ziptree

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Option for zipping
type Option func(*config)

type config struct {
	Compress bool
	Mtime    time.Time
}

func newConfig() *config {
	return &config{Compress: true}
}

// Compress is an Option to compress a zip or not
func Compress(f bool) Option {
	return func(c *config) {
		c.Compress = f
	}
}

// FixMtime is an Option to fix mtimes of the file in the zip
func FixMtime(t time.Time) Option {
	return func(c *config) {
		c.Mtime = t
	}
}

// Zip a directory tree
func Zip(srcPath string, opts ...Option) ([]byte, error) {
	c := newConfig()
	for _, opt := range opts {
		opt(c)
	}

	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)
	if err := filepath.Walk(srcPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore directories and hidden files.
		// No entry is needed for directories in a zip file.
		// Each file is represented with a path, no directory
		// entities are required to build the hierarchy.
		if fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			return nil
		}
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		fHeader, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		if !c.Mtime.IsZero() {
			// Always use the same modification time so that
			// the output is deterministic with respect to the file contents.
			// Do NOT use fHeader.Modified as it only works on go >= 1.10
			fHeader.SetModTime(c.Mtime)
		}
		fHeader.Name = filepath.ToSlash(relPath)
		if c.Compress {
			fHeader.Method = zip.Deflate
		}
		f, err := w.CreateHeader(fHeader)
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		return err
	}); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
