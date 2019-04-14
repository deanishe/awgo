// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"testing"
)

// Mock magic action
type mockMA struct {
	keyCalled     bool
	descCalled    bool
	runTextCalled bool
	runCalled     bool
	returnError   bool

	keyword string
}

func (a *mockMA) Keyword() string {
	a.keyCalled = true
	if a.keyword != "" {
		return a.keyword
	}
	return "test"
}
func (a *mockMA) Description() string {
	a.descCalled = true
	return "Just a test"
}
func (a *mockMA) RunText() string {
	a.runTextCalled = true
	return "Performing test…"
}
func (a *mockMA) Run() error {
	a.runCalled = true
	if a.returnError {
		return errors.New("requested error")
	}
	return nil
}

// Returns an error if the MA wasn't "shown".
// That means MagicActions didn't show a list of actions.
func (a *mockMA) ValidateShown() error {

	if !a.keyCalled {
		return errors.New("Keyword() not called")
	}

	if !a.descCalled {
		return errors.New("Description() not called")
	}

	if a.runCalled {
		return errors.New("Run() called")
	}

	if a.runTextCalled {
		return errors.New("RunText() called")
	}

	return nil
}

// Returns an error if the MA wasn't run.
func (a *mockMA) ValidateRun() error {

	if !a.keyCalled {
		return errors.New("Keyword() not called")
	}

	if a.descCalled {
		return errors.New("Description() called")
	}

	if !a.runCalled {
		return errors.New("Run() not called")
	}

	if !a.runTextCalled {
		return errors.New("RunText() not called")
	}

	return nil
}

// TestNonMagicArgs tests that normal arguments aren't ignored
func TestNonMagicArgs(t *testing.T) {

	data := []struct {
		in, x []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, td := range data {

		wf := New()
		ma := wf.MagicActions

		args, handled := ma.handleArgs(td.in, DefaultMagicPrefix)

		if handled {
			t.Error("handled")
		}

		if !slicesEqual(args, td.x) {
			t.Errorf("not equal. Expected=%v, Got=%v", td.x, args)
		}
	}

}

func TestMagicDefaults(t *testing.T) {

	wf := New()
	ma := wf.MagicActions

	x := 6
	v := len(ma.actions)
	if v != x {
		t.Errorf("Bad MagicAction count. Expected=%d, Got=%d", x, v)
	}
}

func TestMagicActions(t *testing.T) {
	t.Parallel()

	wf := New()
	ma := wf.MagicActions
	ta := &mockMA{}

	ma.Register(ta)
	// Incomplete keyword = search query
	_, v := ma.handleArgs([]string{"workflow:tes"}, DefaultMagicPrefix)
	if !v {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	if err := ta.ValidateShown(); err != nil {
		t.Errorf("Bad MagicAction: %v", err)
	}

	// Test unregister
	ma.Unregister(ta)

	n := len(ma.actions)
	x := 6
	if n != x {
		t.Errorf("Bad MagicActions length. Expected=%v, Got=%v", x, n)
	}

	// Register a new action
	ta = &mockMA{}
	ma.Register(ta)

	// Keyword of test MA
	_, v = ma.handleArgs([]string{"workflow:test"}, DefaultMagicPrefix)
	if !v {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	if err := ta.ValidateRun(); err != nil {
		t.Errorf("Bad MagicAction: %v", err)
	}
}

// Test automatically-added updateMA.
func TestMagicUpdate(t *testing.T) {
	t.Parallel()

	u := &mockUpdater{}
	// Workflow automatically adds a MagicAction to call the Updater
	wf := New(Update(u))
	ma := wf.MagicActions

	// Incomplete keyword = search query
	_, v := ma.handleArgs([]string{"workflow:upda"}, DefaultMagicPrefix)
	if !v {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	// Keyword of update MA
	_, v = ma.handleArgs([]string{"workflow:update"}, DefaultMagicPrefix)
	if !v {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	if !u.checkForUpdateCalled {
		t.Errorf("Bad update. CheckForUpdate not called")
	}
	if !u.updateAvailableCalled {
		t.Errorf("Bad update. UpdateAvailable not called")
	}
	if !u.installCalled {
		t.Errorf("Bad update. Install not called")
	}
}
