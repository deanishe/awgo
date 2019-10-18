// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
// That means magicActions didn't show a list of actions.
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
	t.Parallel()

	data := []struct {
		in, x []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, td := range data {
		td := td
		t.Run(fmt.Sprintf("%v", td.in), func(t *testing.T) {
			t.Parallel()
			wf := New()
			ma := wf.magicActions

			args, handled := ma.handleArgs(td.in, DefaultMagicPrefix)
			assert.False(t, handled, "handled")
			assert.Equal(t, td.x, args, "unexpected non-magic arguments")
		})
	}

}

func TestMagicDefaults(t *testing.T) {
	helpURL := "https://github.com/deanishe/awgo"

	withTestWf(func(wf *Workflow) {
		wf.Configure(HelpURL(helpURL))
		ma := wf.magicActions

		x := 7
		v := len(ma.actions)
		if v != x {
			t.Errorf("Bad MagicAction count. Expected=%d, Got=%d", x, v)
		}

		tests := []struct {
			in   string
			name string
			args []string
		}{
			{"workflow:cache", "open", []string{"open", wf.CacheDir()}},
			{"workflow:log", "open", []string{"open", wf.LogFile()}},
			{"workflow:data", "open", []string{"open", wf.DataDir()}},
			{"workflow:help", "open", []string{"open", helpURL}},
		}

		for _, td := range tests {
			me := &mockExec{}
			wf.execFunc = me.Run
			exitFunc = func(int) {}
			_ = wf.magicActions.args([]string{td.in}, "workflow:")
			assert.Equal(t, td.name, me.name, "Unexpected command name")
			assert.Equal(t, td.args, me.args, "Unexpected command args")
		}
	})
}

func TestMagicActions(t *testing.T) {

	tests := []struct {
		in    string
		shown bool
		run   bool
	}{
		{"workflow:tes", true, false},
		{"workflow:test", false, true},
		//{"workflow:test", true, false, true},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("MagicAction(%q)", td.in), func(t *testing.T) {
			var (
				wf = New()
				ta = &mockMA{}
			)
			exitFunc = func(int) {}
			wf.magicActions.register(ta)
			_ = wf.magicActions.args([]string{td.in}, DefaultMagicPrefix)
			if err := ta.ValidateShown(); err != nil && td.shown {
				t.Error("Not Shown")
			}
			if err := ta.ValidateRun(); err != nil && td.run {
				t.Error("Not Run")
			}
		})
	}
}

// Test MagicArgs call os.Exit.
func TestMagicExits(t *testing.T) {
	tests := []struct {
		in   string
		exit bool
	}{
		{"prefix:", true},
		{"prefix", false},
	}

	defer func() { exitFunc = os.Exit }()

	// test wf.magicActions
	for _, td := range tests {
		td := td
		withTestWf(func(wf *Workflow) {
			me := &mockExit{}
			exitFunc = me.Exit
			wf.magicActions.args([]string{td.in}, "prefix:")
			assert.Equal(t, 0, me.code, "MagicArgs did not exit")
		})
	}

	origArgs := os.Args[:]
	defer func() {
		os.Args = origArgs
	}()

	// test wf.args
	for _, td := range tests {
		td := td
		withTestWf(func(wf *Workflow) {
			me := &mockExit{}
			exitFunc = me.Exit
			os.Args = []string{"blah", td.in}
			wf.Configure(MagicPrefix("prefix:"))
			wf.Args()
			assert.Equal(t, 0, me.code, "wf.args did not exit")
		})
	}
}

// Test automatically-added updateMA.
func TestMagicUpdate(t *testing.T) {
	t.Parallel()

	u := &mockUpdater{}
	// Workflow automatically adds a MagicAction to call the Updater
	wf := New(Update(u))
	ma := wf.magicActions

	// Incomplete keyword = search query
	_, v := ma.handleArgs([]string{"workflow:upda"}, DefaultMagicPrefix)
	assert.True(t, v, "non-magic arguments handled")

	// Keyword of update MA
	_, v = ma.handleArgs([]string{"workflow:update"}, DefaultMagicPrefix)
	assert.True(t, v, "non-magic arguments handled")

	assert.True(t, u.checkForUpdateCalled, "CheckForUpdate not called")
	assert.True(t, u.updateAvailableCalled, "UpdateAvailable not called")
	assert.True(t, u.installCalled, "Install not called")
}
