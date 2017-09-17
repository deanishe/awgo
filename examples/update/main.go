//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-05
//

/*
Workflow update is an example of how to use AwGo's update API.

It demonstrates best practices for handling updates, in particular
loading the list of releases in a background process and only showing
an "Update is available!" message if the user query is empty.

Details

Its own version (set in info.plist via Alfred's UI) is 0.2 and it's
pointing to the GitHub repo "deanishe/alfred-ssh" (a completely
different workflow), which is several versions ahead.

The Script Filter code loads the time of the last update check, and if a
check is due, it calls itself with the "check" command via AwGo's
background job API.

When run with "check", the program calls Workflow.CheckForUpdate() to
cache the available releases.

After that has completed, subsequent runs of the Script Filter will show
an "Update available!" item (if the query is empty).

Actioning (hitting ↩ or ⌘+1) or completing the item (hitting ⇥)
auto-completes the query to "workflow:update", which is the keyword for
one of AwGo's "magic" actions.

At this point, AwGo will take control of execution, and download and
install the newer version of the workflow (but as it's pointing to a
different workflow's repo, Alfred will install a different workflow
rather than actually updating this one).
*/
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/docopt/docopt-go"
)

// Name of the background job that checks for updates
const updateJobName = "checkForUpdate"

var usage = `update (search|check) [<query>]

Demonstrates self-updating using AwGo.

Usage:
    update search [<query>]
    update check
    update -h

Options:
    -h, --help    Show this message and exit.
`

var (
	// Icon to show if an update is available
	iconAvailable = &aw.Icon{Value: "update-available.png"}
	repo          = "deanishe/alfred-ssh" // GitHub repo
	wf            *aw.Workflow            // Our Workflow struct
)

func init() {
	wf = aw.New(update.GitHub(repo))
}

func run() {
	// Pass wf.Args() to docopt because our update logic relies on AwGo's magic actions.
	args, _ := docopt.Parse(usage, wf.Args(), true, wf.Version(), false, true)

	// Alternate action: Get available releases from remote.
	if args["check"] != false {
		wf.TextErrors = true
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// ----------------------------------------------------------------
	// Script Filter
	// ----------------------------------------------------------------

	var query string
	if args["<query>"] != nil {
		query = args["<query>"].(string)
	}

	log.Printf("query=%s", query)

	// Call self with "check" command if an update is due and a check
	// job isn't already running.
	if wf.UpdateCheckDue() && !aw.IsRunning(updateJobName) {
		log.Println("Running update check in background...")

		cmd := exec.Command(os.Args[0], "check")
		if err := aw.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	// Only show update status if query is empty.
	if query == "" && wf.UpdateAvailable() {
		wf.NewItem("Update available!").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconAvailable)
	}

	// Script Filter results
	for i := 1; i <= 20; i++ {
		t := fmt.Sprintf("Item #%d", i)
		wf.NewItem(t).
			Arg(t).
			Icon(aw.IconFavourite).
			Valid(true)
	}

	// Add an extra item to reset update status for demo purposes
	wf.NewItem("Reset update status").
		Autocomplete("workflow:delcache").
		Icon(aw.IconTrash).
		Valid(false)

	// Filter results on user query if present
	if query != "" {
		wf.Filter(query)
	}

	wf.WarnEmpty("No matching items", "Try a different query?")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
