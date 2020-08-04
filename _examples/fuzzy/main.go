// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

/*
Workflow fuzzy is a basic demonstration of AwGo's fuzzy filtering.

It displays and filters the contents of your Downloads directory in Alfred,
and allows you to open files, reveal in Finder or browse in Alfred.
*/
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/deanishe/awgo"
)

var (
	// Where we'll look for directories
	startDir = os.ExpandEnv("${HOME}/Downloads")
	// Our Workflow object
	wf *aw.Workflow
)

type file struct {
	Path  string
	IsDir bool
}

// readDir returns the paths to all visible subdirectories of `dirpath`.
func readDir(dir string) (files []file) {
	infos, _ := ioutil.ReadDir(dir)
	for _, fi := range infos {
		// ignore hidden files
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		files = append(files, file{filepath.Join(dir, fi.Name()), fi.IsDir()})
	}
	return files
}

// run executes the Script Filter.
func run() {
	// ----------------------------------------------------------------
	// Handle CLI arguments
	// ----------------------------------------------------------------

	// You should always use wf.Args() in Script Filters. It contains the
	// same as os.Args[1:], but the arguments are first parsed for AwGo's
	// magic actions (i.e. "workflow:*" to allow the user to easily open
	// the log or data/cache directory).
	query := wf.Args()[0]

	// ----------------------------------------------------------------
	// Load data and create Alfred items
	// ----------------------------------------------------------------

	for _, file := range readDir(startDir) {
		// Convenience method. Sets Item title to filename, subtitle
		// to shortened path, arg to full path, and icon to file icon.
		it := wf.NewFileItem(file.Path)

		// Alternate actions
		it.NewModifier(aw.ModCmd).
			Subtitle("Reveal in Finder").
			Var("action", "reveal")

		if file.IsDir {
			it.NewModifier(aw.ModAlt).
				Subtitle("Browse in Alfred").
				Var("action", "browse")
		}
	}

	// ----------------------------------------------------------------
	// Filter items based on user query
	// ----------------------------------------------------------------

	if query != "" {
		wf.Filter(query)
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
	// Initialise workflow
	wf = aw.New()
	// Call workflow via `Run` wrapper to catch any errors, log them
	// and display an error message in Alfred.
	wf.Run(run)
}
