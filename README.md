# statik

[![Build Status](https://travis-ci.org/rakyll/statik.png?branch=master)](https://travis-ci.org/rakyll/statik)

statik allows you to embed a directory of static files into your Go binary to be later served from an http.FileSystem.

Is this a crazy idea? No, not necessarily. If you're building a tool that has a Web component, you typically want to serve some images, CSS and JavaScript. You like the comfort of distributing a single binary, so you don't want to mess with deploying them elsewhere. If your static files are not large in size and will be browsed by a few people, statik is a solution you are looking for.

## Usage

Install the command line tool first.

	go get github.com/rakyll/statik

statik is a tiny program that reads a directory and generates a source file contains its contents. The generated source file registers the directory contents to be used by statik file system.

The command below will walk on the public path and generate a package called `statik` under the current working directory.

    $ statik -src=/path/to/your/project/public
    
Let's take a look at the generated source file.

	$ cat statik/statik.go

~~~ go
package statik

import (
  "github.com/rakyll/statik/fs"
)

func init() {
  data := // ...
  fs.Register(data)
}
~~~

In your program, all your need to do is to import the generated package, initialize a new statik file system and serve.

~~~ go
import (
  _ "./statik"
  "github.com/rakyll/statik/fs"
)

// ...

statikFS, _ := fs.New()
http.ListenAndServe(":8080", http.FileServer(statikFS))
~~~

Visit http://localhost:8080/path/to/file to see your file.


Note: The idea and the implementation are hijacked from [camlistore](http://camlistore.org/). I decided to decouple it from its codebase due to the fact I'm actively in need of a similar solution for many of my projects.
