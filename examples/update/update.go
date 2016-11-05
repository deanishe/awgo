//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-05
//

/*
This workflow is an example of how to use AwGo's update API.

Its own version (set in info.plist) is 0.1 and it's pointing to the
GitHub repo of "alfred-ssh" (a completely different workflow), which
is several version ahead.

The first time you run the workflow, it will call itself in the background
with the environment variable "check_update=true".

When this variable is set, the program calls CheckForUpdate(), which
retrieves and caches the available releases. When this is complete,
you will see an "Update available!" message in Alfred's results.

Actioning (hitting ↩ or ⌘+1) or completing it (hitting ⇥) auto-completes
the item's text to "workflow:update", which is one of AwGo's "magic"
arguments. At this point, AwGo will take control of execution, and
download & install the newer version of the workflow (except it's a
different workflow).
*/
package main

import (
	"fmt"
	"log"
	"os"

	"os/exec"

	"syscall"

	"gogs.deanishe.net/deanishe/awgo"
)

var (
	iconAvailable = &aw.Icon{Value: "update-available.png"}
	iconUpToDate  = &aw.Icon{Value: "up-to-date.png"}
	repo          = "deanishe/alfred-ssh"
	opts          *aw.Options
	wf            *aw.Workflow
)

func init() {
	opts = &aw.Options{GitHub: repo}
	wf = aw.NewWorkflow(opts)
}

func run() {
	// Alternate action: Get available releases from remote
	if os.Getenv("check_update") == "true" {
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// ----------------------------------------------------------------
	// Main script

	var query string
	args := wf.Args()
	if len(args) > 0 {
		query = args[0]
	}
	log.Printf("query=%s", query)

	// Call self in background to update local releases cache
	if wf.UpdateCheckDue() { // Run check update in background
		log.Println("Starting update checker in background...")
		cmd := exec.Command("./update")
		env := os.Environ()
		env = append(env, "check_update=true")
		cmd.Env = env
		// Ensure process isn't killed if parent (this process) is
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			wf.FatalError(err)
		}
	} else {
		log.Println("Update check not due")
	}

	if query == "" { // Only show update status if query is empty
		// Send update status to Alfred
		if wf.UpdateAvailable() {
			wf.NewItem("Update available!").
				Subtitle("↩ to install").
				Autocomplete("workflow:update").
				Valid(false).
				Icon(iconAvailable)
		}
	}

	// Actual Script Filter items
	for i := 1; i < 21; i++ {
		wf.NewItem(fmt.Sprintf("Item #%d", i)).
			Icon(aw.IconFavourite)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.WarnEmpty("No matching items", "Try a different query?")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
