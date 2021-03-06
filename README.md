[docimg]:https://godoc.org/github.com/ardnew/mkgo?status.svg
[docurl]:https://godoc.org/github.com/ardnew/mkgo
[repimg]:https://goreportcard.com/badge/github.com/ardnew/mkgo
[repurl]:https://goreportcard.com/report/github.com/ardnew/mkgo

# mkgo
#### Create a Go main module using template source file

[![GoDoc][docimg]][docurl] [![Go Report Card][repimg]][repurl]

> mkgo creates a new Go main module using a source code template.
> It integrates `github.com/ardnew/version` to embed a version and changelog,
> and it uses the standard `flag` package to accept command-line arguments.
> The module is created at a given Go import path relative to the first path
> found in the user's `GOPATH` environment variable.

## Usage

Simply provide the Go import path of the desired module:

```sh
mkgo github.com/ardnew/mycmd
```

Also generate a simple `README.md` (with GoDoc and GoReportCard badges) and `LICENSE` (MIT) file:

```sh
mkgo -r -l MIT -u ardnew github.com/ardnew/mycmd
```

If there were no errors, you should see the following output:

```
mkgo: successfully created "github.com/ardnew/myapp": /home/andrew/Code/go/src/github.com/ardnew/myapp
```

Use the `-h` flag for usage summary:

```
Usage of mkgo:
  -changelog
		display change history
  -d string
		date of initial revision (default "2020 Oct 10")
  -f    force overwriting file if it already exists
  -l string
		create a LICENSE file (options: MIT)
  -r    create a simple README.md
  -s string
		semantic version of initial revision (default "0.1.0")
  -u string
		user name for license file copyright (default "andrew")
  -version
		display version information
```

## Installation

Use the builtin Go package manager:

```sh
go get -v github.com/ardnew/mkgo
```

