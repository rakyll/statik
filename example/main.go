//go:generate statik -src=./public

package main

import (
	"log"
	"net/http"

	_ "github.com/rakyll/statik/example/statik"
	"github.com/rakyll/statik/fs"
	"github.com/rakyll/statik/fs/httpfs"
)

// Before buildling, run go generate.
// Then, run the main program and visit http://localhost:8080/public/hello.txt
func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(httpfs.System(statikFS))))
	http.ListenAndServe(":8080", nil)
}
