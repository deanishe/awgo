// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"log"
)

// Updater can check for and download & install newer versions of the workflow.
// There is a concrete implementation and documentation in subpackage update.
type Updater interface {
	UpdateAvailable() bool // Return true if a newer version is available
	CheckDue() bool        // Return true if a check for a newer version is due
	CheckForUpdate() error // Retrieve available releases, e.g. from a URL
	Install() error        // Install the latest version
}

// --------------------------------------------------------------------
// Updating

// setUpdater sets an updater for the workflow.
func (wf *Workflow) setUpdater(u Updater) {
	wf.Updater = u
	wf.MagicActions.Register(&updateMA{wf.Updater})
}

// UpdateCheckDue returns true if an update is available.
func (wf *Workflow) UpdateCheckDue() bool {
	if wf.Updater == nil {
		log.Println("No updater configured")
		return false
	}
	return wf.Updater.CheckDue()
}

// CheckForUpdate retrieves and caches the list of available releases.
func (wf *Workflow) CheckForUpdate() error {
	if wf.Updater == nil {
		return errors.New("No updater configured")
	}
	return wf.Updater.CheckForUpdate()
}

// UpdateAvailable returns true if a newer version is available to install.
func (wf *Workflow) UpdateAvailable() bool {
	if wf.Updater == nil {
		log.Println("No updater configured")
		return false
	}
	return wf.Updater.UpdateAvailable()
}

// InstallUpdate downloads and installs the latest version of the workflow.
func (wf *Workflow) InstallUpdate() error {
	if wf.Updater == nil {
		return errors.New("No updater configured")
	}
	return wf.Updater.Install()
}
