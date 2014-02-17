package main

import (
	"log"
	"net/http"

	_ "github.com/rakyll/statik/example/statik"
	"github.com/rakyll/statik/fs"
)

// Before buildling, run `statik -src=./public`
// to generate the statik package.
// Then, run the main program and visit http://localhost:8080/hello.txt
func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf(err.Error())
	}

	http.ListenAndServe(":8080", http.FileServer(statikFS))
}
