package fs

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var zipData string
var zipModTime time.Time

type statikFS struct {
	files map[string]*zip.File
}

func Register(modTime time.Time, data string) {
	zipModTime = modTime
	zipData = data
}

func New() (http.FileSystem, error) {
	if zipData == "" {
		return nil, os.ErrNotExist
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
	fi, _ := newFileInfo(f)
	return &file{fileInfo: fi}, nil
}

type file struct {
	*fileInfo
	once sync.Once // for making the SectionReader
	sr   *io.SectionReader
}

func (f *file) Read(p []byte) (n int, err error) {
	f.once.Do(f.initReader)
	return f.sr.Read(p)
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	f.once.Do(f.initReader)
	return f.sr.Seek(offset, whence)
}

func (f *file) initReader() {
	f.sr = io.NewSectionReader(f.fileInfo.ra, 0, f.Size())
}

func newFileInfo(zf *zip.File) (*fileInfo, error) {
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	all, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	rc.Close()
	return &fileInfo{
		fullName: zf.Name,
		regdata:  all,
		Closer:   nopCloser,
		ra:       bytes.NewReader(all),
	}, nil
}

var nopCloser = ioutil.NopCloser(nil)

type fileInfo struct {
	fullName string
	regdata  []byte      // non-nil if regular file
	ra       io.ReaderAt // over regdata
	io.Closer
}

func (f *fileInfo) IsDir() bool {
	return f.regdata == nil
}

func (f *fileInfo) Size() int64 {
	return int64(len(f.regdata))
}

func (f *fileInfo) ModTime() time.Time {
	return zipModTime
}

func (f *fileInfo) Name() string {
	return path.Base(f.fullName)
}

func (f *fileInfo) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *fileInfo) Sys() interface{} {
	return nil
}

func (f *fileInfo) Readdir(count int) ([]os.FileInfo, error) {
	// directory listing is disabled.
	var files []os.FileInfo
	return files, nil
}

func (f *fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return 0755 | os.ModeDir
	}
	return 0644
}
