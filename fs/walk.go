// Copyright 2018 Tamás Gulácsi. All Rights Reserved.
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
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
)

var SkipDir = errors.New("skip dir")

//  WalkFunc is the type of the function called for each file or directory visited by Walk.
// The path argument contains the argument to Walk as a prefix;
// that is, if Walk is called with "dir", which is a directory containing the file "a",
// the walk function will be called with argument "dir/a".
// The info argument is the os.FileInfo for the named path.
//
// If there was a problem walking to the file or directory named by path,
// the incoming error will describe the problem and the function
// can decide how to handle that error (and Walk will not descend into that directory).
// If an error is returned, processing stops.
// The sole exception is when the function returns the special value SkipDir.
// If the function returns SkipDir when invoked on a directory,
// Walk skips the directory's contents entirely.
// If the function returns SkipDir when invoked on a non-directory file,
// Walk skips the remaining files in the containing directory.
type WalkFunc func(path string, info os.FileInfo, err error) error

// Walk walks the file tree rooted at root,
// calling walkFn for each file or directory in the tree, including root.
// All errors that arise visiting files and directories are filtered by walkFn.
func Walk(hfs http.FileSystem, root string, walkFn WalkFunc) error {
	dh, err := hfs.Open(root)
	if err != nil {
		return err
	}
	di, err := dh.Stat()
	if err != nil {
		return err
	}
	fis, err := dh.Readdir(-1)
	dh.Close()
	if err = walkFn(root, di, err); err != nil {
		if err == SkipDir {
			return nil
		}
		return err
	}
	for _, fi := range fis {
		fn := path.Join(root, fi.Name())
		if fi.IsDir() {
			if err = Walk(hfs, fn, walkFn); err != nil {
				if err == SkipDir {
					continue
				}
				return err
			}
			continue
		}
		if err = walkFn(fn, fi, nil); err != nil {
			if err == SkipDir {
				continue
			}
			return err
		}
	}
	return nil
}

// ReadFile reads the contents of the file of hfs specified by name.
// Just as ioutil.ReadFile does.
func ReadFile(hfs http.FileSystem, name string) ([]byte, error) {
	fh, err := hfs.Open(name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, fh)
	fh.Close()
	return buf.Bytes(), err
}
