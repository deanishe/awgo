/*
examples/fuzzy shows how to fuzzy filter results using awgo.

It displays and filters a list of subdirectories of ~/ in Alfred, and
allows you to open or reveal the folders, or browse them in Alfred.
*/
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gogs.deanishe.net/deanishe/awgo"
)

var (
	StartDir     string  // Directory to read
	MinimumScore float64 // Search score cutoff
)

// Folders is a simple slice of strings that supports workflow.Fuzzy
// to allow fuzzy searching.
type Folders []string

// Default sort.Interface methods
func (f Folders) Len() int           { return len(f) }
func (f Folders) Less(i, j int) bool { return f[i] < f[j] }
func (f Folders) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Keywords implements workflow.Fuzzy. Comparisons are based on the
// basename of the filepath.
func (f Folders) Keywords(i int) string { return filepath.Base(f[i]) }

func init() {
	StartDir = os.Getenv("HOME")
	MinimumScore = 0.3
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
	paths := readDir(StartDir)

	// Because the program is called from Alfred with "{query}",
	// $1 will always be set, even if to an emtpy string.
	query := os.Args[1]

	// Filter results if query isn't empty.
	if query != "" {
		// Filter results
		for i, score := range workflow.SortFuzzy(paths, query) {
			if score < MinimumScore {
				paths = paths[:i]
				break
			}
		}
	}

	// Generate feedback for Alfred
	for _, path := range paths {
		it := workflow.NewFileItem(path)
		it.SetAlternateSubtitle("cmd", "Reveal in Finder.")
		it.SetAlternateSubtitle("alt", "Browse in Alfred.")
	}

	// Send XML to Alfred. After calling this function, you can't send
	// any more results to Alfred.
	workflow.SendFeedback()
}

func main() {
	// Call workflow via `Run` wrapper to catch any errors, log them
	// and display an error message in Alfred.
	workflow.Run(run)
}
