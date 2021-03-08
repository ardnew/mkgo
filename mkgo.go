// mkgo creates a new Go main module using a source code template.
// It integrates `github.com/ardnew/version` to embed a version and changelog,
// and it uses the standard `flag` package to accept command-line arguments.
// The module is created at a given Go import path relative to the first path
// found in the user's `GOPATH` environment variable.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ardnew/version"
)

func init() {
	version.ChangeLog = []version.Change{{
		Package: "mkgo",
		Version: "0.1.0",
		Date:    "2020 Oct 10",
		Description: []string{
			"initial implementation",
		},
	}, {
		Package: "mkgo",
		Version: "0.2.0",
		Date:    "2020 Oct 10",
		Description: []string{
			"add support for README and LICENSE generation",
		},
	}, {
		Package: "mkgo",
		Version: "0.2.1",
		Date:    "2021 Mar 8",
		Description: []string{
			"use shorter, abbreviated flags in generated source",
		},
	}}
}

func main() {

	var (
		argChanges   bool
		argVersion   bool
		argMkDate    string
		argMkVersion string
		argReadme    bool
		argLicense   string
		argUser      string
		argOverwrite bool
	)

	currDate := time.Now().Format(dateFormat)
	currUser := os.Getenv("USER")

	knownLicense := []string{}
	for name := range licenseTemplate {
		knownLicense = append(knownLicense, name)
	}

	flag.BoolVar(&argChanges, "changelog", false, "display change history")
	flag.BoolVar(&argVersion, "version", false, "display version information")
	flag.StringVar(&argMkDate, "d", currDate, "date of initial revision")
	flag.StringVar(&argMkVersion, "s", semVersion, "semantic version of initial revision")
	flag.BoolVar(&argOverwrite, "f", false, "force overwriting file if it already exists")
	flag.BoolVar(&argReadme, "r", false, "create a simple README.md")
	flag.StringVar(&argLicense, "l", "", "create a LICENSE file (options: "+strings.Join(knownLicense, " ")+")")
	flag.StringVar(&argUser, "u", currUser, "user name for license file copyright")
	flag.Parse()

	if argChanges {
		version.PrintChangeLog()
	} else if argVersion {
		fmt.Printf("mkgo version %s\n", version.String())
	} else {

		if len(flag.Args()) == 0 {
			fmt.Println("error: no package path specified (use -h for help)")
			os.Exit(1)
		}

		path, name := packagePath(flag.Arg(0))
		if err := os.MkdirAll(path, os.ModePerm); nil != err {
			fmt.Printf("error: %s\n", err.Error())
			os.Exit(2)
		}

		sourcePath := filepath.Join(path, name+".go")
		if exists, isDir := fileExists(sourcePath); !exists || argOverwrite {
			if isDir {
				fmt.Printf("error: output file is a directory: %s\n", sourcePath)
				os.Exit(3)
			}
			template.insert(flag.Arg(0), name, argMkDate, argMkVersion, argUser)
			if err := ioutil.WriteFile(sourcePath, []byte(template.String()), 0664); nil != err {
				fmt.Printf("error: %s\n", err.Error())
				os.Exit(4)
			}
			if out, err := execCmd(path, "goimports", "-w", name+".go"); nil != err {
				fmt.Print(out)
				os.Exit(5)
			}
			if out, err := execCmd(path, "go", "mod", "init"); nil != err {
				fmt.Print(out)
				os.Exit(6)
			}
		} else {
			fmt.Printf("error: file exists (use -f to overwrite): %s\n", sourcePath)
			os.Exit(7)
		}

		licensePath := filepath.Join(path, "LICENSE")
		if license, ok := licenseTemplate[argLicense]; !ok {
			fmt.Printf("error: unsupported license (use -h to view options): %s\n", argLicense)
			os.Exit(8)
		} else {
			if exists, isDir := fileExists(licensePath); !exists || argOverwrite {
				if isDir {
					fmt.Printf("error: output file is a directory: %s\n", licensePath)
					os.Exit(9)
				}
				license.insert(flag.Arg(0), name, argMkDate, argMkVersion, argUser)
				if err := ioutil.WriteFile(licensePath, []byte(license.String()), 0664); nil != err {
					fmt.Printf("error: %s\n", err.Error())
					os.Exit(10)
				}
			} else {
				fmt.Printf("error: file exists (use -f to overwrite): %s\n", licensePath)
				os.Exit(11)
			}
		}

		readmePath := filepath.Join(path, "README.md")
		if exists, isDir := fileExists(readmePath); !exists || argOverwrite {
			if isDir {
				fmt.Printf("error: output file is a directory: %s\n", readmePath)
				os.Exit(9)
			}
			readme.insert(flag.Arg(0), name, argMkDate, argMkVersion, argUser)
			if err := ioutil.WriteFile(readmePath, []byte(readme.String()), 0664); nil != err {
				fmt.Printf("error: %s\n", err.Error())
				os.Exit(10)
			}
		} else {
			fmt.Printf("error: file exists (use -f to overwrite): %s\n", readmePath)
			os.Exit(11)
		}

		fmt.Printf("mkgo: successfully created %q: %s\n", flag.Arg(0), path)
	}
}

// execCmd runs the given system command cmd with given arguments arg from the
// given working directory dir, returning the combined stdout/stderr output.
func execCmd(dir, cmd string, arg ...string) (string, error) {
	c := exec.Command(cmd, arg...)
	c.Dir = dir
	o, err := c.CombinedOutput()
	return string(o), err
}

// fileExists returns whether or not a file exists, and if it exists whether or
// not it is a directory.
func fileExists(path string) (exists, isDir bool) {

	stat, err := os.Stat(path)
	exists = err == nil || !os.IsNotExist(err)
	isDir = exists && stat.IsDir()
	return
}

// splitPath returns a string slice whose elements are each of the components in
// a given file path.
func splitPath(path string) []string {
	part := []string{}
	// extract the last component from path until no components remain.
	for len(path) > 0 {
		d, f := filepath.Split(path)
		if len(f) > 0 {
			part = append(part, f)
		}
		if len(d) > 0 {
			path = filepath.Clean(d)
		} else {
			break
		}
	}
	// reverse the components
	n := len(part)
	for i := 0; i < n/2; i++ {
		part[i], part[n-i-1] = part[n-i-1], part[i]
	}
	return part
}

// packagePath returns the absolute file path of given Go package's import path
// relative to the first path found in the user's GOPATH environment variable.
func packagePath(path string) (full, name string) {
	gopath := filepath.SplitList(os.Getenv("GOPATH"))
	part := splitPath(path)
	if len(gopath) > 0 {
		full = filepath.Join(gopath[0], "src", filepath.Join(part...))
	}
	if len(part) > 0 {
		name = part[len(part)-1]
	}
	return full, name
}

// Template represents a file whose elements are individual lines of the file.
type Template []string

// insert replaces all placeholder tokens in the receiver Template's elements
// with the given replacement values, returning the resulting Template.
func (tmpl *Template) insert(path, name, date, version, user string) *Template {
	for i, s := range *tmpl {
		(*tmpl)[i] =
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(
							strings.ReplaceAll(s,
								"__IMPORT__", path),
							"__NAME__", name),
						"__DATE__", date),
					"__VERSION__", version),
				"__USER__", user)
	}
	return tmpl
}

// String returns all elements of the receiver Template joined by newline.
func (tmpl *Template) String() string {
	return strings.Join(*tmpl, "\n")
}

var (
	dateFormat = "2006 Jan 02"
	semVersion = "0.1.0"
	template   = Template{
		`package main`,
		``,
		`import (`,
		`	"flag"`,
		`	"fmt"`,
		``,
		`	"github.com/ardnew/version"`,
		`)`,
		``,
		`func init() {`,
		`	version.ChangeLog = []version.Change{{`,
		`		Package: "__NAME__",`,
		`		Version: "__VERSION__",`,
		`		Date:    "__DATE__",`,
		`		Description: []string{`,
		`			"initial implementation",`,
		`		},`,
		`	}}`,
		`}`,
		``,
		`func main() {`,
		``,
		`	var (`,
		`		argVersion bool`,
		`		argChanges bool`,
		`	)`,
		``,
		`	flag.BoolVar(&argVersion, "v", false, "Display version information")`,
		`	flag.BoolVar(&argChanges, "V", false, "Display change history")`,
		`	flag.Parse()`,
		``,
		`	if argChanges {`,
		`		version.PrintChangeLog()`,
		`	} else if argVersion {`,
		`		fmt.Printf("__NAME__ version %s\n", version.String())`,
		`	} else {`,
		`		// main`,
		`	}`,
		`}`,
	}
	licenseTemplate = map[string]Template{
		"MIT": Template{
			`MIT License`,
			``,
			`Copyright (c) 2020 __USER__`,
			``,
			`Permission is hereby granted, free of charge, to any person obtaining a copy`,
			`of this software and associated documentation files (the "Software"), to deal`,
			`in the Software without restriction, including without limitation the rights`,
			`to use, copy, modify, merge, publish, distribute, sublicense, and/or sell`,
			`copies of the Software, and to permit persons to whom the Software is`,
			`furnished to do so, subject to the following conditions:`,
			``,
			`The above copyright notice and this permission notice shall be included in all`,
			`copies or substantial portions of the Software.`,
			``,
			`THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR`,
			`IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,`,
			`FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE`,
			`AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER`,
			`LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,`,
			`OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE`,
			`SOFTWARE.			`,
		},
	}
	readme = Template{
		`[docimg]:https://godoc.org/__IMPORT__?status.svg`,
		`[docurl]:https://godoc.org/__IMPORT__`,
		`[repimg]:https://goreportcard.com/badge/__IMPORT__`,
		`[repurl]:https://goreportcard.com/report/__IMPORT__`,
		``,
		`# __NAME__`,
		`#### __NAME__`,
		``,
		`[![GoDoc][docimg]][docurl] [![Go Report Card][repimg]][repurl]`,
		``,
		`## Usage`,
		``,
		`How to use:`,
		``,
		"```sh",
		`__NAME__ ...`,
		"```",
		``,
		"Use the `-h` flag for usage summary:",
		``,
		"```",
		`Usage of __NAME__:`,
		`  -changelog`,
		`		display change history`,
		`  -version`,
		`		display version information`,
		"```",
		``,
		`## Installation`,
		``,
		`Use the builtin Go package manager:`,
		``,
		"```sh",
		`go get -v __IMPORT__`,
		"```",
	}
)
