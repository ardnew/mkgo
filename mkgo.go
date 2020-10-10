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
	}}
}

func main() {

	var (
		argChangeLog bool
		argVersion   bool
		argMkDate    string
		argMkVersion string
		argOverwrite bool
	)

	currDate := time.Now().Format(dateFormat)

	flag.BoolVar(&argChangeLog, "changelog", false, "display change history")
	flag.BoolVar(&argVersion, "version", false, "display version information")
	flag.StringVar(&argMkDate, "d", currDate, "date of initial revision")
	flag.StringVar(&argMkVersion, "s", semVersion, "semantic version of initial revision")
	flag.BoolVar(&argOverwrite, "f", false, "force overwriting file if it already exists")
	flag.Parse()

	if argChangeLog {
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
			err := ioutil.WriteFile(sourcePath,
				[]byte(template.insert(name, argMkDate, argMkVersion).String()), 0664)
			if nil != err {
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
			fmt.Printf("mkgo: successfully created %q: %s\n", flag.Arg(0), path)
		} else {
			fmt.Printf("error: file exists (use -f to overwrite): %s\n", sourcePath)
			os.Exit(7)
		}
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

// Template represents a Go source code file whose elements are individual lines
// of the file.
type Template []string

// insert replaces all placeholder tokens in the receiver Template's elements
// with the given replacement values, returning the resulting Template.
func (tmpl *Template) insert(name, date, version string) *Template {
	for i, s := range *tmpl {
		(*tmpl)[i] =
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(s,
						"__NAME__", name),
					"__DATE__", date),
				"__VERSION__", version)
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
		`		argChangeLog bool`,
		`		argVersion   bool`,
		`	)`,
		``,
		`	flag.BoolVar(&argChangeLog, "changelog", false, "display change history")`,
		`	flag.BoolVar(&argVersion, "version", false, "display version information")`,
		`	flag.Parse()`,
		``,
		`	if argChangeLog {`,
		`		version.PrintChangeLog()`,
		`	} else if argVersion {`,
		`		fmt.Printf("__NAME__ version %s\n", version.String())`,
		`	} else {`,
		`		// main`,
		`	}`,
		`}`,
	}
)
