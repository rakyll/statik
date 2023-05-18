# statik

[![Build Status](https://travis-ci.org/rakyll/statik.svg?branch=master)](https://travis-ci.org/rakyll/statik)

statik allows you to embed a directory of static files into your Go binary to be later served from an http.FileSystem.

Is this a crazy idea? No, not necessarily. If you're building a tool that has a Web component, you typically want to serve some images, CSS and JavaScript. You like the comfort of distributing a single binary, so you don't want to mess with deploying them elsewhere. If your static files are not large in size and will be browsed by a few people, statik is a solution you are looking for.

## Usage

Install the command line tool first.

	go install github.com/rakyll/statik@latest

statik is a tiny program that reads a directory and generates a source file that contains its contents. The generated source file registers the directory contents to be used by statik file system.

The command below will walk on the public path and generate a package called `statik` under the current working directory.

    $ statik -src=/path/to/your/project/public

The command below will filter only files on listed extensions.

    $ statik -include=*.jpg,*.txt,*.html,*.css,*.js

In your program, all your need to do is to import the generated package, initialize a new statik file system and serve.

~~~ go
import (
  "github.com/rakyll/statik/fs"

  _ "./statik" // TODO: Replace with the absolute import path
)

  // ...

  statikFS, err := fs.New()
  if err != nil {
    log.Fatal(err)
  }
  
  // Serve the contents over HTTP.
  http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(statikFS)))
  http.ListenAndServe(":8080", nil)
~~~

Visit http://localhost:8080/public/path/to/file to see your file.

You can also read the content of a single file:

~~~ go
import (
  "github.com/rakyll/statik/fs"

  _ "./statik" // TODO: Replace with the absolute import path
)

  // ...

  statikFS, err := fs.New()
  if err != nil {
    log.Fatal(err)
  }
  
  // Access individual files by their paths.
  r, err := statikFS.Open("/hello.txt")
  if err != nil {
    log.Fatal(err)
  }    
  defer r.Close()
  contents, err := ioutil.ReadAll(r)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(string(contents))
~~~

There is also a working example under [example](https://github.com/rakyll/statik/tree/master/example) directory, follow the instructions to build and run it.

Note: The idea and the implementation are hijacked from [camlistore](http://camlistore.org/). I decided to decouple it from its codebase due to the fact I'm actively in need of a similar solution for many of my projects.

## Deterministic output

By default, statik includes the "last modified" (mtime) time on files that it packs. This allows an HTTP FileServer to present the correct file modification times to clients.

However, if you have a continuous integration task that checks that your checked-in static files in a git repository match the code that is generated on your CI system, you'll run into a problem: The mtime on the git checkout does not match what you have locally, causing tests to fail.

You can fix the test in one of two ways:

1. In CI, manually set the mtime on the freshly checked out tree: [here's a stackoverflow answer](https://stackoverflow.com/a/22638823/93405) that provides a shell command to do that; or,
2. Instruct statik not to store the "last modified" time.

To ignore the last modified time, use the `-m` to statik, like so:

    $ statik -m -include=*.jpg,*.txt,*.html,*.css,*.js

Note that this will cause http.FileServer to consider the file to always have changed & serve it with a "Last-Modified" of the time of the request.
