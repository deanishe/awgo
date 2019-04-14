// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"testing"
	"time"
)

// ensure mockUpdater implements Updater
var _ Updater = (*mockUpdater)(nil)

type mockUpdater struct {
	updateIntervalCalled  bool
	checkDueCalled        bool
	checkForUpdateCalled  bool
	updateAvailableCalled bool
	installCalled         bool

	checkShouldFail   bool
	installShouldFail bool
}

// UpdateInterval implements Updater.
func (d *mockUpdater) UpdateInterval(_ time.Duration) {
	d.updateIntervalCalled = true
}

// UpdateAvailable implements Updater.
func (d *mockUpdater) UpdateAvailable() bool {
	d.updateAvailableCalled = true
	return true
}

// CheckDue implements Updater.
func (d *mockUpdater) CheckDue() bool {
	d.checkDueCalled = true
	return true
}

// CheckForUpdate implements Updater.
func (d *mockUpdater) CheckForUpdate() error {
	d.checkForUpdateCalled = true
	if d.checkShouldFail {
		return errors.New("check failed")
	}
	return nil
}

// Install implements Updater.
func (d *mockUpdater) Install() error {
	d.installCalled = true
	if d.installShouldFail {
		return errors.New("install failed")
	}
	return nil
}

// Test that Workflow API responses match configured Updater's.
func TestWorkflowUpdater(t *testing.T) {
	t.Parallel()

	wf := New()
	// false/fail when Updater is unset
	if wf.UpdateCheckDue() {
		t.Error("Bad UpdateCheckDue. Expected=false, Got=true")
	}
	if wf.UpdateAvailable() {
		t.Error("Bad UpdateAvailable. Expected=false, Got=true")
	}
	if err := wf.CheckForUpdate(); err == nil {
		t.Error("CheckForUpdate() succeeded, expected failure")
	}
	if err := wf.InstallUpdate(); err == nil {
		t.Error("InstallUpdate() succeeded, expected failure")
	}

	// true/success with mockUpdater
	u := &mockUpdater{}
	_ = wf.Configure(Update(u))
	if !wf.UpdateCheckDue() {
		t.Error("Bad UpdateCheckDue. Expected=true, Got=false")
	}
	if !u.checkDueCalled {
		t.Error("Bad Update. CheckDue not called")
	}
	if !wf.UpdateAvailable() {
		t.Error("Bad UpdateAvailable. Expected=true, Got=false")
	}
	if !u.updateAvailableCalled {
		t.Error("Bad Update. UpdateAvailable not called")
	}
	if err := wf.CheckForUpdate(); err != nil {
		t.Error("CheckForUpdate() failed")
	}
	if !u.checkForUpdateCalled {
		t.Error("Bad Update. CheckForUpdate not called")
	}
	if err := wf.InstallUpdate(); err != nil {
		t.Error("InstallUpdate() failed")
	}
	if !u.installCalled {
		t.Error("Bad Update. Install not called")
	}
}
