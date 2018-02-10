//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
Workflow workflows searches GitHub repos tagged with "alfred-workflow".

It demonstrates the use of the caching and background-process APIs to
provide a responsive workflow by updating the datastore in the background.

It shows results based on cached data (if available), and if the cached data
are out-of-date, starts a background process to refresh the cache.

The Script Filter reloads the results (by setting Workflow.Rerun) until the
cached data have been updated (at which point it's showing the latest data).

This is a very useful idiom for workflows that don't need data that are
absolutely bang up-to-date. The user still gets (potentially out-of-date)
results, preserving the responsiveness of the workflow, and the latest data
are shown as soon as they're available.
*/
package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/deanishe/awgo"
	"github.com/deanishe/awgo/fuzzy"
	"github.com/docopt/docopt-go"
)

var (
	cacheName   = "repos.json"      // Filename of cached repo list
	maxResults  = 200               // Number of results sent to Alfred
	minScore    = 10.0              // Minimum score for a result
	maxCacheAge = 180 * time.Minute // How long to cache repo list for
)

var (
	usage = `repos [options] [<query>]

Usage:
	repos <query>
	repos --download
	repos -h|--version

Options:
    --download     download list of workflows to cache
	-h, --help     show this message and exit
	--version      show version number and exit
`
	sopts []fuzzy.Option
	wf    *aw.Workflow
)

func init() {
	// Set some custom fuzzy search options
	sopts = []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.HelpURL("http://www.deanishe.net/"), aw.MaxResults(maxResults), aw.SortOptions(sopts...))
}

func run() {
	var query string

	// Version is read from the `alfred_workflow_version` environment variable,
	// which Alfred sets to the value specified in info.plist.
	args, err := docopt.Parse(usage, wf.Args(), true, wf.Version(), false)
	if err != nil {
		log.Fatalf("Error parsing CLI options : %v", err)
	}

	if v, ok := args["--download"].(bool); ok {
		if v == true {
			wf.TextErrors = true
			log.Printf("[main] downloading repo list...")
			repos, err := fetchRepos()
			if err != nil {
				wf.FatalError(err)
			}
			if err := wf.Cache.StoreJSON(cacheName, repos); err != nil {
				wf.FatalError(err)
			}
			log.Printf("[main] downloaded repo list")
			return
		}
	}

	// Docopt values are interface{} :(
	if s, ok := args["<query>"].(string); ok {
		query = s
	}
	log.Printf("[main] query=%s", query)

	// Try to load repos
	repos := []*Repo{}
	if wf.Cache.Exists(cacheName) {
		if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
			wf.FatalError(err)
		}
	}

	// If the cache has expired, set Rerun (which tells Alfred to re-run the
	// workflow), and start the background update process if it isn't already
	// running.
	if wf.Cache.Expired(cacheName, maxCacheAge) {
		wf.Rerun(0.3)
		if !aw.IsRunning("download") {
			cmd := exec.Command(os.Args[0], "--download")
			if err := aw.RunInBackground("download", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("download job already running.")
		}
		// Cache is also "expired" if it doesn't exist. So if there are no
		// cached data, show a corresponding message and exit.
		if len(repos) == 0 {
			wf.NewItem("Downloading repos…").
				Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}

	}

	// Add results for cached repos
	for _, r := range repos {
		wf.NewItem(r.FullName()).
			Subtitle(r.Description).
			Arg(r.URL).
			UID(r.FullName()).
			Valid(true)
	}

	// Filter results against query if user entered one
	if query != "" {
		res := wf.Filter(query)
		log.Printf("[main] %d/%d repos match \"%s\"", len(res), len(repos), query)
	}

	// Convenience method that shows a warning if there are no results to show.
	// Alfred's default behaviour if no results are returned is to show its
	// fallback searches, which is also what it does if a workflow errors out.
	//
	// As such, it's a good idea to display a message in this situation,
	// otherwise the user can't tell if the workflow failed or simply found
	// no matching results.
	wf.WarnEmpty("No repos found", "Try a different query?")

	// Send results/warning message to Alfred
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
