// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.False(t, wf.UpdateCheckDue(), "unexpected UpdateCheckDue")
	assert.False(t, wf.UpdateAvailable(), "unexpected UpdateAvailable")
	assert.NotNil(t, wf.CheckForUpdate(), "CheckForUpdate succeeded")
	assert.NotNil(t, wf.InstallUpdate(), "InstallUpdate succeeded")

	// true/success with mockUpdater
	u := &mockUpdater{}
	_ = wf.Configure(Update(u))
	assert.True(t, wf.UpdateCheckDue(), "unexpected UpdateCheckDue")
	assert.True(t, u.checkDueCalled, "checkDue not called")
	assert.True(t, wf.UpdateAvailable(), "unexpected UpdateAvailable")
	assert.True(t, u.updateAvailableCalled, "updateAvailable not called")
	assert.Nil(t, wf.CheckForUpdate(), "CheckForUpdate failed")
	assert.True(t, u.checkForUpdateCalled, "checkForUpdate not called")
	assert.Nil(t, wf.InstallUpdate(), "InstallUpdate failed")
	assert.True(t, u.installCalled, "installCalled not called")
}
