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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type expectedFile struct {
	data    []byte
	isDir   bool
	modTime time.Time
	mode    os.FileMode
	name    string
	size    int64
	err     error
}

func TestOpen(t *testing.T) {
	tests := []struct {
		description   string
		zipData       string
		expectedFiles map[string]expectedFile
	}{
		{
			description: "Files should retain their original file mode and modified time",
			zipData:     mustZipTree("../testdata/file"),
			expectedFiles: map[string]expectedFile{
				"/file.txt": {
					data:    mustReadFile("../testdata/file/file.txt"),
					isDir:   false,
					modTime: mustStatFile("../testdata/file/file.txt").ModTime(),
					mode:    mustStatFile("../testdata/file/file.txt").Mode(),
					name:    mustStatFile("../testdata/file/file.txt").Name,
					size:    int64(mustStatFile("../testdata/file/file.txt").UncompressedSize64),
				},
			},
		},
		{
			description: "Images should successfully unpack",
			zipData:     mustZipTree("../testdata/image"),
			expectedFiles: map[string]expectedFile{
				"/pixel.gif": {
					data:    mustReadFile("../testdata/image/pixel.gif"),
					isDir:   false,
					modTime: mustStatFile("../testdata/image/pixel.gif").ModTime(),
					mode:    mustStatFile("../testdata/image/pixel.gif").Mode(),
					name:    mustStatFile("../testdata/image/pixel.gif").Name,
					size:    int64(mustStatFile("../testdata/image/pixel.gif").UncompressedSize64),
				},
			},
		},
		{
			description: "'index.html' files should be returned at their original path and their directory path",
			zipData:     mustZipTree("../testdata/index"),
			expectedFiles: map[string]expectedFile{
				"/index.html": {
					data:    mustReadFile("../testdata/index/index.html"),
					isDir:   false,
					modTime: mustStatFile("../testdata/index/index.html").ModTime(),
					mode:    mustStatFile("../testdata/index/index.html").Mode(),
					name:    mustStatFile("../testdata/index/index.html").Name,
					size:    int64(mustStatFile("../testdata/index/index.html").UncompressedSize64),
				},
				"/": {
					data:    mustReadFile("../testdata/index/index.html"),
					isDir:   true,
					modTime: mustStatFile("../testdata/index/index.html").ModTime(),
					mode:    mustStatFile("../testdata/index/index.html").Mode(),
					name:    mustStatFile("../testdata/index/index.html").Name,
					size:    int64(mustStatFile("../testdata/index/index.html").UncompressedSize64),
				},
				"/sub_dir/index.html": {
					data:    mustReadFile("../testdata/index/sub_dir/index.html"),
					isDir:   false,
					modTime: mustStatFile("../testdata/index/sub_dir/index.html").ModTime(),
					mode:    mustStatFile("../testdata/index/sub_dir/index.html").Mode(),
					name:    mustStatFile("../testdata/index/sub_dir/index.html").Name,
					size:    int64(mustStatFile("../testdata/index/sub_dir/index.html").UncompressedSize64),
				},
				"/sub_dir/": {
					data:    mustReadFile("../testdata/index/sub_dir/index.html"),
					isDir:   true,
					modTime: mustStatFile("../testdata/index/sub_dir/index.html").ModTime(),
					mode:    mustStatFile("../testdata/index/sub_dir/index.html").Mode(),
					name:    mustStatFile("../testdata/index/sub_dir/index.html").Name,
					size:    int64(mustStatFile("../testdata/index/sub_dir/index.html").UncompressedSize64),
				},
			},
		},
		{
			description: "Missing files should return os.ErrNotExist",
			zipData:     mustZipTree("../testdata/file"),
			expectedFiles: map[string]expectedFile{
				"/missing.txt": {
					err: os.ErrNotExist,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			Register(tc.zipData)
			fs, err := New()
			if err != nil {
				t.Errorf("Error creating new fs: %s", err)
				return
			}
			for name, expectedFile := range tc.expectedFiles {
				f, err := fs.Open(name)
				if expectedFile.err != err {
					t.Errorf(
						"Expected and actual error opening file different for file %q.\nExpected:\t%s\nActual:\t\t%s",
						name,
						expectedFile.err,
						err)
				}
				if err != nil {
					continue
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					t.Errorf(
						"Error reading file %q: %s", name, err)
					continue
				}
				if !reflect.DeepEqual(expectedFile.data, b) {
					t.Errorf(
						"Expected and actual file data different for file %q.\nExpected:\t%v\nActual:\t\t%v",
						name,
						expectedFile.data,
						b)
				}
				stat, _ := f.Stat()
				if expectedFile.isDir != stat.IsDir() {
					t.Errorf(
						"Expected and actual file IsDir different for file %q.\nExpected:\t%t\nActual:\t\t%t",
						name,
						expectedFile.isDir,
						stat.IsDir())
				}
				if expectedFile.modTime != stat.ModTime() {
					t.Errorf(
						"Expected and actual file ModTime different for file %q.\nExpected:\t%s (%d)\nActual:\t\t%s (%d)",
						name,
						expectedFile.modTime,
						expectedFile.modTime.UnixNano(),
						stat.ModTime(),
						stat.ModTime().UnixNano())
				}
				if expectedFile.mode != stat.Mode() {
					t.Errorf(
						"Expected and actual file Mode different for file %q.\nExpected:\t%s\nActual:\t\t%s",
						name,
						expectedFile.mode,
						stat.Mode())
				}
				if expectedFile.name != stat.Name() {
					t.Errorf(
						"Expected and actual file Name different for file %q.\nExpected:\t%s\nActual:\t\t%s",
						name,
						expectedFile.name,
						stat.Name())
				}
				if expectedFile.size != stat.Size() {
					t.Errorf(
						"Expected and actual file Size different for file %q.\nExpected:\t%d\nActual:\t\t%d",
						name,
						expectedFile.size,
						stat.Size())
				}
			}
		})
	}
}

// Test that calling Open by many goroutines concurrently continues
// to return the expected result.
func TestOpen_Parallel(t *testing.T) {
	indexFileContents := mustReadFile("../testdata/index/index.html")
	Register(mustZipTree("../testdata/index"))
	fs, err := New()
	if err != nil {
		t.Fatalf("Error creating new fs: %s", err)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := fs.Open("/index.html")
			if err != nil {
				t.Errorf("Error opening file '/index.html': %s", err)
				return
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("Error reading file '/index.html': %s", err)
				return
			}
			if !reflect.DeepEqual(indexFileContents, b) {
				t.Errorf(
					"Expected and actual file data different for file '/index.html'.\nExpected:\t%v\nActual:\t\t%v",
					indexFileContents,
					b)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkOpen(b *testing.B) {
	Register(mustZipTree("../testdata/index"))
	fs, err := New()
	if err != nil {
		b.Fatalf("Error creating new fs: %s", err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := fs.Open("/index.html")
			if err != nil {
				b.Errorf("Error opening file '/index.html': %s", err)
			}
		}
	})
}

// mustZipTree walks on the source path and returns the zipped file contents
// as a string. Panics on any errors.
func mustZipTree(srcPath string) string {
	var out bytes.Buffer
	absPath, err := filepath.Abs(srcPath)
	if err != nil {
		panic(err)
	}
	absPathWithoutSymlinks, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		panic(err)
	}
	w := zip.NewWriter(&out)
	if err := filepath.Walk(absPathWithoutSymlinks, func(path string, fi os.FileInfo, err error) error {
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
		relPath, err := filepath.Rel(absPathWithoutSymlinks, path)
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
	absPath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}
	absPathWithoutSymlinks, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(absPathWithoutSymlinks)
	if err != nil {
		panic(err)
	}
	return b
}

// mustStatFile returns the zip file info header. Panics on any errors.
func mustStatFile(filename string) *zip.FileHeader {
	info, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	zipInfo, err := zip.FileInfoHeader(info)
	if err != nil {
		panic(err)
	}
	return zipInfo
}
