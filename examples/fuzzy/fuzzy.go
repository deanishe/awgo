//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
examples/fuzzy shows how to fuzzy filter results using awgo.

It displays and filters a list of subdirectories of ~/ in Alfred, and
allows you to open or reveal the folders, or browse them in Alfred.
*/
package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gogs.deanishe.net/deanishe/awgo"
	"gogs.deanishe.net/deanishe/awgo/fuzzy"
)

var (
	startDir     string  // Directory to read
	minimumScore float64 // Search score cutoff
)

// Folders is a simple slice of strings that supports fuzzy.Interface
// to allow fuzzy searching.
type Folders []string

// Default sort.Interface methods
func (f Folders) Len() int           { return len(f) }
func (f Folders) Less(i, j int) bool { return f[i] < f[j] }
func (f Folders) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Keywords implements fuzzy.Interface. Comparisons are based on the
// basename of the filepath.
func (f Folders) Keywords(i int) string { return filepath.Base(f[i]) }

func init() {
	startDir = os.Getenv("HOME")
	minimumScore = 0.3
}

// readDir returns the paths to all the visible subdirectories of `dirpath`
func readDir(dirpath string) Folders {
	paths := Folders{}
	files, _ := ioutil.ReadDir(dirpath)
	for _, fi := range files {
		// Ignore files and hidden files
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		paths = append(paths, filepath.Join(dirpath, fi.Name()))
	}
	return paths
}

// run runs the workflow
func run() {
	var query string
	paths := readDir(startDir)

	if len(os.Args) > 1 {
		// When run from a workflow, because the program is called from Alfred
		// with "{query}" or "$1", $1 will always be set, even if to an
		// emtpy string.
		query = os.Args[1]
	}

	// Filter results if query isn't empty.
	if query != "" {
		// Filter results
		for i, score := range fuzzy.Sort(paths, query) {
			log.Printf("%v\t%v", score, paths[i])
			if score < minimumScore {
				paths = paths[:i]
				break
			}
		}
	}

	// Generate feedback for Alfred
	for _, path := range paths {

		it := workflow.NewFileItem(path)

		m, _ := it.NewModifier("cmd")
		m.SetSubtitle("Reveal in Finder.")

		m, _ = it.NewModifier("alt")
		m.SetSubtitle("Browse in Alfred.")
	}

	// Send JSON to Alfred. After calling this function, you can't send
	// any more results to Alfred.
	workflow.SendFeedback()
}

func main() {
	// Call workflow via `Run` wrapper to catch any errors, log them
	// and display an error message in Alfred.
	workflow.Run(run)
}
