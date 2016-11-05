//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-05
//

package main

import (
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
	updater       *aw.Updater
	opts          *aw.Options
	wf            *aw.Workflow
)

func init() {
	opts = &aw.Options{}
	wf = aw.NewWorkflow(opts)
}

func run() {
	// Create updater within run, as it may panic if the workflow
	// version is invalid.
	updater = aw.NewUpdater(&aw.GitHub{Repo: repo})

	// Alternate action: Get available releases from remote
	if os.Getenv("check_update") == "true" {
		// Tell Workflow to print any errors as simple text messages to
		// STDOUT, so they'll be shown in the Post Notification
		wf.TextErrors = true
		log.Println("Checking for updates...")
		if err := updater.CheckUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// Alternate action: Download and install update
	if os.Getenv("do_update") == "true" {
		// Not a Script Filter action
		wf.TextErrors = true
		if err := updater.Install(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// ----------------------------------------------------------------
	// Main script

	// Call self in background to update local releases cache
	if updater.CheckDue() { // Run check update in background
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
	}

	// Send update status to Alfred
	if updater.UpdateAvailable() {
		wf.NewItem("Update available!").
			Subtitle("↩ to install").
			Valid(true).
			Icon(iconAvailable).
			Var("do_update", "true")
	} else {
		wf.NewItem("Your workflow is up to date").
			Valid(false).
			Icon(iconUpToDate)
	}
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
