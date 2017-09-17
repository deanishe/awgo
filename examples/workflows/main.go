//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
Workflow workflows demonstrates using cached data for maximum speed.



It demonstrates using AwGo's caching API.
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
    --download  Download list of books to cache.
	-h, --help  Show this message and exit.
	--version   Show version number and exit.
`
	sopts []fuzzy.Option
	wf    *aw.Workflow
)

func init() {
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

	// Version is parsed from info.plist
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
		// Exit if there are no repos to show
		if len(repos) == 0 {
			wf.NewItem("Downloading repos…").
				Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}

	}

	for _, r := range repos {
		wf.NewItem(r.FullName()).
			Subtitle(r.Description).
			Arg(r.URL).
			UID(r.FullName()).
			Valid(true)
	}

	if query != "" {
		res := wf.Filter(query)
		log.Printf("[main] %d/%d repos match \"%s\"", len(res), len(repos), query)
	}

	wf.WarnEmpty("No repos found", "Try a different query?")

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
