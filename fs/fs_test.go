// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fs

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type wantFile struct {
	data    []byte
	isDir   bool
	modTime time.Time
	mode    os.FileMode
	name    string
	size    int64
	err     error
}

func TestOpen(t *testing.T) {
	fileTxtHeader := mustFileHeader("../testdata/file/file.txt")
	pixelGifHeader := mustFileHeader("../testdata/image/pixel.gif")
	indexHTMLHeader := mustFileHeader("../testdata/index/index.html")
	subdirIndexHTMLHeader := mustFileHeader("../testdata/index/sub_dir/index.html")
	deepAHTMLHeader := mustFileHeader("../testdata/deep/a")
	deepCHTMLHeader := mustFileHeader("../testdata/deep/aa/bb/c")
	tests := []struct {
		description string
		zipData     string
		wantFiles   map[string]wantFile
	}{
		{
			description: "Files should retain their original file mode and modified time",
			zipData:     mustZipTree("../testdata/file"),
			wantFiles: map[string]wantFile{
				"/file.txt": {
					data:    mustReadFile("../testdata/file/file.txt"),
					isDir:   false,
					modTime: fileTxtHeader.ModTime(),
					mode:    fileTxtHeader.Mode(),
					name:    fileTxtHeader.Name,
					size:    int64(fileTxtHeader.UncompressedSize64),
				},
			},
		},
		{
			description: "Images should successfully unpack",
			zipData:     mustZipTree("../testdata/image"),
			wantFiles: map[string]wantFile{
				"/pixel.gif": {
					data:    mustReadFile("../testdata/image/pixel.gif"),
					isDir:   false,
					modTime: pixelGifHeader.ModTime(),
					mode:    pixelGifHeader.Mode(),
					name:    pixelGifHeader.Name,
					size:    int64(pixelGifHeader.UncompressedSize64),
				},
			},
		},
		{
			description: "'index.html' files should be returned at their original path and their directory path",
			zipData:     mustZipTree("../testdata/index"),
			wantFiles: map[string]wantFile{
				"/index.html": {
					data:    mustReadFile("../testdata/index/index.html"),
					isDir:   false,
					modTime: indexHTMLHeader.ModTime(),
					mode:    indexHTMLHeader.Mode(),
					name:    indexHTMLHeader.Name,
					size:    int64(indexHTMLHeader.UncompressedSize64),
				},
				"/sub_dir/index.html": {
					data:    mustReadFile("../testdata/index/sub_dir/index.html"),
					isDir:   false,
					modTime: subdirIndexHTMLHeader.ModTime(),
					mode:    subdirIndexHTMLHeader.Mode(),
					name:    subdirIndexHTMLHeader.Name,
					size:    int64(subdirIndexHTMLHeader.UncompressedSize64),
				},
				"/": {
					isDir: true,
					mode:  os.ModeDir | 0755,
					name:  "/",
				},
				"/sub_dir": {
					isDir: true,
					mode:  os.ModeDir | 0755,
					name:  "/sub_dir",
				},
			},
		},
		{
			description: "Missing files should return os.ErrNotExist",
			zipData:     mustZipTree("../testdata/file"),
			wantFiles: map[string]wantFile{
				"/missing.txt": {
					err: os.ErrNotExist,
				},
			},
		},
		{
			description: "listed all sub directories in deep directory",
			zipData:     mustZipTree("../testdata/deep"),
			wantFiles: map[string]wantFile{
				"/a": {
					data:    mustReadFile("../testdata/deep/a"),
					isDir:   false,
					modTime: deepAHTMLHeader.ModTime(),
					mode:    deepAHTMLHeader.Mode(),
					name:    deepAHTMLHeader.Name,
					size:    int64(deepAHTMLHeader.UncompressedSize64),
				},
				"/aa/bb/c": {
					data:    mustReadFile("../testdata/deep/aa/bb/c"),
					isDir:   false,
					modTime: deepCHTMLHeader.ModTime(),
					mode:    deepCHTMLHeader.Mode(),
					name:    deepCHTMLHeader.Name,
					size:    int64(deepCHTMLHeader.UncompressedSize64),
				},
				"/": {
					isDir: true,
					mode:  os.ModeDir | 0755,
					name:  "/",
				},
				"/aa": {
					isDir: true,
					mode:  os.ModeDir | 0755,
					name:  "/aa",
				},
				"/aa/bb": {
					isDir: true,
					mode:  os.ModeDir | 0755,
					name:  "/aa/bb",
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			Register(tc.zipData)
			fs, err := New()
			if err != nil {
				t.Errorf("New() = %v", err)
				return
			}
			for name, wantFile := range tc.wantFiles {
				f, err := fs.Open(name)
				if wantFile.err != err {
					t.Errorf("fs.Open(%v) = %v; want %v", name, err, wantFile.err)
				}
				if err != nil {
					continue
				}
				if !wantFile.isDir {
					b, err := ioutil.ReadAll(f)
					if err != nil {
						t.Errorf("ioutil.ReadAll(%v) = %v", name, err)
						continue
					}
					if !reflect.DeepEqual(wantFile.data, b) {
						t.Errorf("%v data = %q; want %q", name, b, wantFile.data)
					}
				}
				stat, err := f.Stat()
				if err != nil {
					t.Errorf("Stat(%v) = %v", name, err)
				}
				if got, want := stat.IsDir(), wantFile.isDir; got != want {
					t.Errorf("IsDir(%v) = %t; want %t", name, got, want)
				}
				if got, want := stat.ModTime(), wantFile.modTime; got != want {
					t.Errorf("ModTime(%v) = %v; want %v", name, got, want)
				}
				if got, want := stat.Mode(), wantFile.mode; got != want {
					t.Errorf("Mode(%v) = %v; want %v", name, got, want)
				}
				if got, want := stat.Name(), path.Base(wantFile.name); got != want {
					t.Errorf("Name(%v) = %v; want %v", name, got, want)
				}
				if got, want := stat.Size(), wantFile.size; got != want {
					t.Errorf("Size(%v) = %v; want %v", name, got, want)
				}
			}
		})
	}
}

func TestWalk(t *testing.T) {
	Register(mustZipTree("../testdata/deep"))
	fs, err := New()
	if err != nil {
		t.Errorf("New() = %v", err)
		return
	}
	var files []string
	err = Walk(fs, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Errorf("Walk(fs, /) = %v", err)
		return
	}
	wantDirs := []string{
		"/",
		"/a",
		"/aa",
		"/aa/bb",
		"/aa/bb/c",
	}
	sort.Strings(files)
	if !reflect.DeepEqual(files, wantDirs) {
		t.Errorf("got:    %v\nexpect: %v", files, wantDirs)
	}
}

func TestHTTPFile_Readdir(t *testing.T) {
	Register(mustZipTree("../testdata/readdir"))
	fs, err := New()
	if err != nil {
		t.Errorf("New() = %v", err)
		return
	}
	t.Run("Readdir(-1)", func(t *testing.T) {
		dir, err := fs.Open("/")
		if err != nil {
			t.Errorf("fs.Open(/) = %v", err)
			return
		}
		fis, err := dir.Readdir(-1)
		if err != nil {
			t.Errorf("dir.Readdir(-1) = %v", err)
			return
		}
		if len(fis) != 3 {
			t.Errorf("got: %d, expect: 3", len(fis))
		}
	})
	t.Run("Readdir(>0)", func(t *testing.T) {
		dir, err := fs.Open("/")
		if err != nil {
			t.Errorf("fs.Open(/) = %v", err)
			return
		}
		fis, err := dir.Readdir(1)
		if err != nil {
			t.Errorf("dir.Readdir(1) = %v", err)
			return
		}
		if len(fis) != 1 {
			t.Errorf("got: %d, expect: 1", len(fis))
		}
		if fis[0].Name() != "aa" {
			t.Errorf("got: %s, expect: aa", fis[0].Name())
		}
		fis, err = dir.Readdir(1)
		if err != nil {
			t.Errorf("dir.Readdir(1) = %v", err)
			return
		}
		if len(fis) != 1 {
			t.Errorf("got: %d, expect: 1", len(fis))
		}
		if fis[0].Name() != "bb" {
			t.Errorf("got: %s, expect: bb", fis[0].Name())
		}
		fis, err = dir.Readdir(-1) // take rest entries
		if err != nil {
			t.Errorf("dir.Readdir(1) = %v", err)
			return
		}
		if len(fis) != 1 {
			t.Errorf("got: %d, expect: 1", len(fis))
		}
		if fis[0].Name() != "cc" {
			t.Errorf("got: %s, expect: cc", fis[0].Name())
		}
		fis, err = dir.Readdir(-1)
		if err != nil {
			t.Errorf("dir.Readdir(1) = %v", err)
			return
		}
		if len(fis) != 0 {
			t.Errorf("got: %d, expect: 0", len(fis))
		}
		fis, err = dir.Readdir(1)
		if err != io.EOF {
			t.Errorf("error should be io.EOF, but: %s", err)
			return
		}
		if len(fis) != 0 {
			t.Errorf("got: %d, expect: 0", len(fis))
		}
	})
}

// Test that calling Open by many goroutines concurrently continues
// to return the expected result.
func TestOpen_Parallel(t *testing.T) {
	indexHTMLData := mustReadFile("../testdata/index/index.html")
	Register(mustZipTree("../testdata/index"))
	fs, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := "/index.html"
			f, err := fs.Open(name)
			if err != nil {
				t.Errorf("fs.Open(%v) = %v", name, err)
				return
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("ioutil.ReadAll(%v) = %v", name, err)
				return
			}
			if !reflect.DeepEqual(indexHTMLData, b) {
				t.Errorf("%v data = %q; want %q", name, b, indexHTMLData)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkOpen(b *testing.B) {
	Register(mustZipTree("../testdata/index"))
	fs, err := New()
	if err != nil {
		b.Fatalf("New() = %v", err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			name := "/index.html"
			_, err := fs.Open(name)
			if err != nil {
				b.Errorf("fs.Open(%v) = %v", name, err)
			}
		}
	})
}

// mustZipTree walks on the source path and returns the zipped file contents
// as a string. Panics on any errors.
func mustZipTree(srcPath string) string {
	var out bytes.Buffer
	w := zip.NewWriter(&out)
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
		fHeader.Name = filepath.ToSlash(relPath)
		fHeader.Method = zip.Deflate
		f, err := w.CreateHeader(fHeader)
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		return err
	}); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return string(out.Bytes())
}

// mustReadFile returns the file contents. Panics on any errors.
func mustReadFile(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}

// mustFileHeader returns the zip file info header. Panics on any errors.
func mustFileHeader(filename string) *zip.FileHeader {
	info, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		panic(err)
	}
	return header
}
