package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var zipData string

type statikFS struct {
	files map[string]*zip.File
}

func Register(data string) {
	zipData = data
}

func New() (http.FileSystem, error) {
	if zipData == "" {
		return nil, errors.New("statik/fs: No zip data registered.")
	}
	zipReader, err := zip.NewReader(strings.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}
	files := make(map[string]*zip.File)
	for _, file := range zipReader.File {
		// name should represent a path to the file
		file.Name = "/" + file.Name
		files[file.Name] = file
	}
	return &statikFS{files: files}, nil
}

func (fs *statikFS) Open(name string) (http.File, error) {
	f, ok := fs.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return newFile(f)
}

var nopCloser = ioutil.NopCloser(nil)

func newFile(zf *zip.File) (*file, error) {
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	all, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	rc.Close()

	return &file{
		FileInfo: zf.FileInfo(),
		data:     all,
		readerAt: bytes.NewReader(all),
		Closer:   nopCloser,
	}, nil
}

type file struct {
	os.FileInfo
	io.Closer

	data     []byte // non-nil if regular file
	reader   *io.SectionReader
	readerAt io.ReaderAt // over data

	once sync.Once
}

func (f *file) newReader() {
	f.reader = io.NewSectionReader(f.readerAt, 0, f.FileInfo.Size())
}

func (f *file) Read(p []byte) (n int, err error) {
	f.once.Do(f.newReader)
	return f.reader.Read(p)
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	f.once.Do(f.newReader)
	return f.reader.Seek(offset, whence)
}

func (f *file) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	// directory listing is disabled.
	return make([]os.FileInfo, 0), nil
}
