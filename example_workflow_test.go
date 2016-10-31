//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
fuzzy-simple shows how to fuzzy filter results using awgo.

It displays and filters a list of subdirectories of ~/ in Alfred, and
allows you to open the folders or browse them in Alfred.

This demo is a complete Alfred 3 workflow.
*/
package aw_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gogs.deanishe.net/deanishe/awgo"
)

var (
	startDir     string             // Directory to read
	minimumScore float64            // Search score cutoff
	wf           *aw.Workflow // Our Workflow object
)

func init() {
	// Where we'll look for directories
	startDir = os.Getenv("HOME")
	// Initialise workflow
	wf = aw.NewWorkflow(nil)
}

// readDir returns the paths to all the visible subdirectories of `dirpath`
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

	// Generate feedback for Alfred
	for _, path := range paths {

		// Convenience method. Sets Item title to filename, subtitle
		// to shortened path, arg to full path, and icon to file icon.
		it := wf.NewFileItem(path)

		// We could set this modifier via Alfred's GUI.
		it.NewModifier("cmd").
			Subtitle("Browse in Alfred")
	}

	// Sort results if query isn't empty.
	if query != "" {
		// Sort results
		res := wf.Filter(query)
		log.Printf("%d results match `%s`", len(res), query)
		for i, r := range res {
			log.Printf("%02d. score=%0.1f sortkey=%s", i+1, r.Score, wf.Feedback.SortKey(i))
		}
	}

	// Send JSON to Alfred. After calling this function, you can't send
	// any more results to Alfred.
	wf.SendFeedback()
}

/*
This is the program from the "fuzzy-simple" demo (see examples/ subdirectory).

It displays and filters a list of subdirectories of ~/ in Alfred,
and allows you to open the folders or browse them in Alfred.

The main program entry point is run(), which is called via Workflow.Run() to
catch any panics, log them, and show the user an error message in Alfred.

This is a complete Script Filter program and will not run outside of
Alfred/without an info.plist.
*/
func ExampleWorkflow_searchHomeDir() {
	// Call workflow via `Run` wrapper to catch any errors, log them
	// and display an error message in Alfred.
	wf.Run(run)
}
