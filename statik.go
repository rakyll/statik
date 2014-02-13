package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	flagSrc  = flag.String("src", ".", "The path of the source directory.")
	flagDest = flag.String("dest", ".", "The destination path of the generated source file.")
)

func main() {
	flag.Parse()
	file, err := createSourceFile(*flagSrc)
	if err != nil {
		exitWithError(err)
	}

	destDir := *flagDest + "/statik"
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		exitWithError(err)
	}

	err = os.Rename(file.Name(), destDir+"/data.go")
	if err != nil {
		exitWithError(err)
	}
}

func createSourceFile(srcPath string) (file *os.File, err error) {
	var buffer bytes.Buffer
	var zipdest io.Writer = &buffer

	f, err := ioutil.TempFile("", "statik-archive")
	if err != nil {
		return
	}

	zipdest = io.MultiWriter(zipdest, f)
	defer f.Close()
	var modTime time.Time

	w := zip.NewWriter(zipdest)
	if err = filepath.Walk(srcPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore empty directories and hidden files.
		if fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			return nil
		}
		suffix, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}
		if mt := fi.ModTime(); mt.After(modTime) {
			modTime = mt
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := w.Create(filepath.ToSlash(suffix))
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		return err
	}); err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}

	// then embed it as a quoted string
	var qb bytes.Buffer
	fmt.Fprint(&qb, "package statik\n\n")
	fmt.Fprint(&qb, "import \"time\"\n\n")
	fmt.Fprint(&qb, "var (\n")
	fmt.Fprint(&qb, "\tStatikDataModTime time.Time\n")
	fmt.Fprint(&qb, "\tStatikData        string\n")
	fmt.Fprint(&qb, ")\n\n")
	fmt.Fprint(&qb, "func init() {\n")
	fmt.Fprintf(&qb, "\tStatikDataModTime = time.Unix(%d, 0)\n", modTime.Unix())
	fmt.Fprint(&qb, "\tStatikData = ")
	quote(&qb, buffer.Bytes())
	fmt.Fprint(&qb, "\n}\n")
	err = ioutil.WriteFile(f.Name(), qb.Bytes(), 0x700)
	if err != nil {
		return
	}
	return f, nil
}

func quote(dest *bytes.Buffer, bs []byte) {
	dest.WriteByte('"')
	for _, b := range bs {
		switch b {
		case '\n':
			dest.WriteString(`\n`)
		case '\\':
			dest.WriteString(`\\`)
		case '"':
			dest.WriteString(`\"`)
		default:
			if (b >= 32 && b <= 126) || b == '\t' {
				dest.WriteByte(b)
			}
		}
		fmt.Fprintf(dest, "\\x%02x", b)
	}
	dest.WriteByte('"')
}

func exitWithError(err error) {
	fmt.Println(err)
	os.Exit(1)
}
