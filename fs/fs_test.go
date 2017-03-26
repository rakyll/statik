package fs

import (
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

// Example zipData with a single file "index.html" containing 80 bytes of HTML
const testZipData = "PK\x03\x04\x14\x00\x08\x00\x08\x00\xdc\nzJ\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\n\x00\x00\x00index.html\xb2\xc9(\xc9\xcd\xb1\xb3\xc9HML\xb1\xb3)\xc9,\xc9I\xb5\xf3H\xcd\xc9\xc9\xd7Q(\xcf/\xcaIQ\xb4\xd1\x87\x08\xda\xe8C\x94$\xe5\xa7T\xa2\xab\x00\x8b\xd9\xe8\x83M\x02\x04\x00\x00\xff\xffPK\x07\x08uR\xdd>:\x00\x00\x00P\x00\x00\x00PK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00\xdc\nzJuR\xdd>:\x00\x00\x00P\x00\x00\x00\n\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa4\x81\x00\x00\x00\x00index.htmlPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x008\x00\x00\x00r\x00\x00\x00\x00\x00"

// Contents of "index.html" file in testZipData
var testZipDataFileData = []byte("<html><head><title>Hello, world!</title></head><body>Hello, world!</body></html>")

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
			zipData:     "PK\x03\x04\x14\x00\x08\x00\x08\x00\x075zJ\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x08\x00\x00\x00file.txt\x01\x00\x00\xff\xffPK\x07\x08\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00\x075zJ\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\x81\x00\x00\x00\x00file.txtPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x006\x00\x00\x00;\x00\x00\x00\x00\x00",
			expectedFiles: map[string]expectedFile{
				"/file.txt": {
					data:    []byte{},
					isDir:   false,
					modTime: time.Unix(0, 1490510414000000000).UTC(),
					mode:    0777,
					name:    "file.txt",
					size:    0,
				},
			},
		},
		{
			description: "Images should successfully unpack",
			zipData: "PK\x03\x04\x14\x00\x08\x00\x08\x00$3zJ\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00	\x00\x00\x00pixel.gifr\xf7t\xb3\xb0Ldd`dh`\x00\x81\xff\xff\xff+\xfeda\x041u@\x04H\x86\x81\x89\xd1\x85\xc1\x1a\x10\x00\x00\xff\xffPK\x07\x08x\x13\x95''\x00\x00\x00*\x00\x00\x00PK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00$3zJx\x13\x95''\x00\x00\x00*\x00\x00\x00	\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa4\x81\x00\x00\x00\x00pixel.gifPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x007\x00\x00\x00^\x00\x00\x00\x00\x00",
			expectedFiles: map[string]expectedFile{
				"/pixel.gif": {
					data:    []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 1, 68, 0, 59},
					isDir:   false,
					modTime: time.Unix(0, 1490509508000000000).UTC(),
					mode:    0644,
					name:    "pixel.gif",
					size:    42,
				},
			},
		},
		{
			description: "'index.html' files should be returned at their original path and their directory path",
			zipData:     "PK\x03\x04\x14\x00\x08\x00\x08\x00\x8b1zJ\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\n\x00\x00\x00index.html2\x04\x04\x00\x00\xff\xffPK\x07\x08\xb7\xef\xdc\x83\x07\x00\x00\x00\x01\x00\x00\x00PK\x03\x04\x14\x00\x08\x00\x08\x00\x8c1zJ\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x00\x00\x00sub_dir/index.html2\x02\x04\x00\x00\xff\xffPK\x07\x08\x0d\xbe\xd5\x1a\x07\x00\x00\x00\x01\x00\x00\x00PK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00\x8b1zJ\xb7\xef\xdc\x83\x07\x00\x00\x00\x01\x00\x00\x00\n\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa4\x81\x00\x00\x00\x00index.htmlPK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00\x8c1zJ\x0d\xbe\xd5\x1a\x07\x00\x00\x00\x01\x00\x00\x00\x12\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa4\x81?\x00\x00\x00sub_dir/index.htmlPK\x05\x06\x00\x00\x00\x00\x02\x00\x02\x00x\x00\x00\x00\x86\x00\x00\x00\x00\x00",
			expectedFiles: map[string]expectedFile{
				"/index.html": {
					data:    []byte("1"),
					isDir:   false,
					modTime: time.Unix(0, 1490508742000000000).UTC(),
					mode:    0644,
					name:    "index.html",
					size:    1,
				},
				"/": {
					data:    []byte("1"),
					isDir:   true,
					modTime: time.Unix(0, 1490508742000000000).UTC(),
					mode:    0644,
					name:    "index.html",
					size:    1,
				},
				"/sub_dir/index.html": {
					data:    []byte("2"),
					isDir:   false,
					modTime: time.Unix(0, 1490508744000000000).UTC(),
					mode:    0644,
					name:    "index.html",
					size:    1,
				},
				"/sub_dir/": {
					data:    []byte("2"),
					isDir:   true,
					modTime: time.Unix(0, 1490508744000000000).UTC(),
					mode:    0644,
					name:    "index.html",
					size:    1,
				},
			},
		},
		{
			description: "Missing files should return os.ErrNotExist",
			zipData:     "PK\x05\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
			expectedFiles: map[string]expectedFile{
				"/missing.txt": {
					err: os.ErrNotExist,
				},
			},
		},
	}
	for _, test := range tests {
		Register(test.zipData)
		fs, err := New()
		if err != nil {
			t.Errorf("%s: Error creating new fs: %s",
				test.description,
				err)
		}
		for name, expectedFile := range test.expectedFiles {
			f, err := fs.Open(name)
			if expectedFile.err != err {
				t.Errorf(
					"%s: Expected and actual error opening file different for file %q.\nExpected:\t%s\nActual:\t\t%s",
					test.description,
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
					"%s: Error reading file %q: %s",
					test.description,
					name,
					err)
				continue
			}
			if !reflect.DeepEqual(expectedFile.data, b) {
				t.Errorf(
					"%s: Expected and actual file data different for file %q.\nExpected:\t%v\nActual:\t\t%v",
					test.description,
					name,
					expectedFile.data,
					b)
			}
			stat, _ := f.Stat()
			if expectedFile.isDir != stat.IsDir() {
				t.Errorf(
					"%s: Expected and actual file IsDir different for file %q.\nExpected:\t%t\nActual:\t\t%t",
					test.description,
					name,
					expectedFile.isDir,
					stat.IsDir())
			}
			if expectedFile.modTime != stat.ModTime() {
				t.Errorf(
					"%s: Expected and actual file ModTime different for file %q.\nExpected:\t%s (%d)\nActual:\t\t%s (%d)",
					test.description,
					name,
					expectedFile.modTime,
					expectedFile.modTime.UTC().UnixNano(),
					stat.ModTime(),
					stat.ModTime().UTC().UnixNano())
			}
			if expectedFile.mode != stat.Mode() {
				t.Errorf(
					"%s: Expected and actual file Mode different for file %q.\nExpected:\t%s\nActual:\t\t%s",
					test.description,
					name,
					expectedFile.mode,
					stat.Mode())
			}
			if expectedFile.name != stat.Name() {
				t.Errorf(
					"%s: Expected and actual file Name different for file %q.\nExpected:\t%s\nActual:\t\t%s",
					test.description,
					name,
					expectedFile.name,
					stat.Name())
			}
			if expectedFile.size != stat.Size() {
				t.Errorf(
					"%s: Expected and actual file Size different for file %q.\nExpected:\t%d\nActual:\t\t%d",
					test.description,
					name,
					expectedFile.size,
					stat.Size())
			}
		}
	}
}

// Test that calling Open by many goroutines concurrently continues
// to return the expected result.
func TestOpen_Parallel(t *testing.T) {
	Register(testZipData)
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
			if !reflect.DeepEqual(testZipDataFileData, b) {
				t.Errorf(
					"Expected and actual file data different for file '/index.html'.\nExpected:\t%v\nActual:\t\t%v",
					testZipDataFileData,
					b)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkOpen(b *testing.B) {
	Register(testZipData)
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
