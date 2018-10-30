package fs

import (
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	type wantPath struct {
		isDir bool
	}
	tests := []struct {
		description string
		zipData     string
		wantPaths   map[string]wantPath
	}{
		{
			zipData: mustZipTree("../testdata/index"),
			wantPaths: map[string]wantPath{
				"/":                   wantPath{isDir: true},
				"/index.html":         wantPath{isDir: false},
				"/sub_dir":            wantPath{isDir: true},
				"/sub_dir/index.html": wantPath{isDir: false},
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

			err = Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					t.Errorf("unexpected error = %v", err)
				}
				if wantPath, ok := tc.wantPaths[path]; ok {
					if got, want := info.IsDir(), wantPath.isDir; got != want {
						t.Errorf("IsDir(%v) = %t; want %t", path, got, want)
					}
					delete(tc.wantPaths, path)
				} else {
					t.Errorf("unexpected path = %v (info = %#v)", path, info)
				}

				return nil
			})

			if err != nil {
				t.Errorf("Walk(fs, \"/\", WalkFunc) = %v", err)
			}

			if len(tc.wantPaths) != 0 {
				t.Errorf("ignored paths: %v", tc.wantPaths)
			}
		})
	}
}
