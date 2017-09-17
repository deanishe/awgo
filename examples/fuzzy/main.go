//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
Workflow fuzzy demonstrates AwGo's fuzzy filtering.

It displays and filters a list of subdirectories of your home directory
in Alfred, and allows you to open the folders in Finder or browse them
in Alfred.
*/
package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/deanishe/awgo"
)

var (
	startDir     string       // Directory to read
	minimumScore float64      // Search score cutoff
	wf           *aw.Workflow // Our Workflow object
)

func init() {
	startDir = os.Getenv("HOME") // Where we'll look for directories
	wf = aw.New()                // Initialise workflow
}

// readDir returns the paths to all the visible subdirectories of `dirpath`.
func readDir(dirpath string) []string {
	paths := []string{}
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

// run executes the Script Filter.
func run() {
	var (
		query       string
		args, paths []string
	)

	// ----------------------------------------------------------------
	// Handle CLI arguments
	// ----------------------------------------------------------------
	args = wf.Args()
	if len(args) > 0 {
		// When run from a workflow, because the program is called from
		// Alfred with "{query}" or "$1", $1 will always be set, even
		// if to an emtpy string.
		query = args[0]
	}

	// ----------------------------------------------------------------
	// Load data and create Alfred items
	// ----------------------------------------------------------------
	paths = readDir(startDir)
	for _, path := range paths {
		// Convenience method. Sets Item title to filename, subtitle
		// to shortened path, arg to full path, and icon to file icon.
		it := wf.NewFileItem(path)

		// We could also set this modifier via Alfred's GUI.
		it.NewModifier("cmd").
			Subtitle("Browse in Alfred")
	}

	// ----------------------------------------------------------------
	// Filter items based on user query
	// ----------------------------------------------------------------
	if query != "" {
		res := wf.Filter(query)
		log.Printf("%d results match \"%s\"", len(res), query)
		for i, r := range res {
			log.Printf("%02d. score=%0.1f sortkey=%s", i+1, r.Score, wf.Feedback.SortKey(i))
		}
	}

	// ----------------------------------------------------------------
	// Send results to Alfred
	// ----------------------------------------------------------------
	// Show a warning in Alfred if there are no items
	wf.WarnEmpty("No matching folders found", "Try a different query?")

	// Send JSON to Alfred. After calling this function, you can't send
	// any more results to Alfred.
	wf.SendFeedback()
}

func main() {
	// Call workflow via `Run` wrapper to catch any errors, log them
	// and display an error message in Alfred.
	wf.Run(run)
}
