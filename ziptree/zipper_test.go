package ziptree_test

import (
	"archive/zip"
	"bytes"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/rakyll/statik/ziptree"
)

var zipData = []byte{
	0x50, 0x4b, 0x03, 0x04, 0x14, 0x00, 0x08, 0x00,
	0x08, 0x00, 0x00, 0x00, 0x21, 0x28, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x05, 0x00, 0x09, 0x00, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x55, 0x54, 0x05, 0x00, 0x01,
	0x80, 0x43, 0x6d, 0x38, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x50, 0x4b, 0x07, 0x08, 0x00, 0x00, 0x00,
	0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x50, 0x4b, 0x01, 0x02, 0x14, 0x03, 0x14,
	0x00, 0x08, 0x00, 0x08, 0x00, 0x00, 0x00, 0x21,
	0x28, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x09,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0xa4, 0x81, 0x00, 0x00, 0x00, 0x00, 0x68,
	0x65, 0x6c, 0x6c, 0x6f, 0x55, 0x54, 0x05, 0x00,
	0x01, 0x80, 0x43, 0x6d, 0x38, 0x50, 0x4b, 0x05,
	0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01,
	0x00, 0x3c, 0x00, 0x00, 0x00, 0x41, 0x00, 0x00,
	0x00, 0x00, 0x00}
var zipData4Print = zipData

func TestZip(t *testing.T) {
	err := os.Chmod("../testdata/ziptree/hello", 0644)
	if err != nil {
		t.Fatal(err)
	}
	out, err := ziptree.Zip(
		"../testdata/ziptree/",
		ziptree.FixMtime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	wantData := zipData
	if !reflect.DeepEqual(out, wantData) {
		t.Errorf("got: %#v\nexpect: %#v", out, wantData)
	}
}

func TestFprintZipData(t *testing.T) {
	buf := &bytes.Buffer{}
	err := ziptree.FprintZipData(buf, zipData4Print)
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	out := buf.String()
	want := `PK\x03\x04\x14\x00\x08\x00\x08\x00\x00\x00!(\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00	\x00helloUT\x05\x00\x01\x80Cm8\x01\x00\x00\xff\xffPK\x07\x08\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x03\x14\x00\x08\x00\x08\x00\x00\x00!(\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\x05\x00	\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa4\x81\x00\x00\x00\x00helloUT\x05\x00\x01\x80Cm8PK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x00<\x00\x00\x00A\x00\x00\x00\x00\x00`
	if out != want {
		t.Errorf("got: %s\nexpect: %s", out, want)
	}
}

func TestZip_CollectFile(t *testing.T) {
	tests := []struct {
		description string
		dir         string
		opts        []ziptree.Option
		wantFiles   []string
	}{
		{
			description: "dot files and files inside dot dirs are not collected",
			dir:         "../testdata/ziptree-skipdir",
			opts:        nil,
			wantFiles:   []string{"general-file"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			out, err := ziptree.Zip(tc.dir, tc.opts...)
			if err != nil {
				t.Errorf("error should be nil, but: %s", err)
			}
			zipReader, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
			l := len(zipReader.File)
			files := make([]string, l)
			for i := 0; i < l; i++ {
				files[i] = zipReader.File[i].Name
			}
			sort.Strings(files)
			if !reflect.DeepEqual(tc.wantFiles, files) {
				t.Errorf("got: %v\nwant: %v", files, tc.wantFiles)
			}
		})
	}
}
