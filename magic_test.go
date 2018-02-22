//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-12
//

package aw

import (
	"errors"
	"testing"
)

// Mock magic action
type testMA struct {
	keyCalled, descCalled, runTextCalled, runCalled bool
	returnError                                     bool
}

func (a *testMA) Keyword() string {
	a.keyCalled = true
	return "test"
}
func (a *testMA) Description() string {
	a.descCalled = true
	return "Just a test"
}
func (a *testMA) RunText() string {
	a.runTextCalled = true
	return "Performing test…"
}
func (a *testMA) Run() error {
	a.runCalled = true
	if a.returnError {
		return errors.New("requested error")
	}
	return nil
}

// Returns an error if the MA wasn't "shown".
// That means MagicActions didn't show a list of actions.
func (a *testMA) ValidateShown() error {

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
func (a *testMA) ValidateRun() error {

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
		in, out []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, td := range data {

		ma := MagicActions{}
		ma.Register(defaultMagicActions...)

		args, handled := ma.handleArgs(td.in, DefaultMagicPrefix)

		if handled {
			t.Error("handled")
		}

		if !slicesEqual(args, td.out) {
			t.Errorf("not equal. Expected=%v, Got=%v", td.out, args)
		}
	}

}

func TestMagicActions(t *testing.T) {

	ma := MagicActions{}
	ta := &testMA{}

	ma.Register(ta)
	// Incomplete keyword = search query
	_, v := ma.handleArgs([]string{"workflow:tes"}, DefaultMagicPrefix)
	if v != true {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	if err := ta.ValidateShown(); err != nil {
		t.Errorf("Bad MagicAction: %v", err)
	}

	// Test unregister
	ma.Unregister(ta)

	if len(ma) != 0 {
		t.Errorf("Bad MagicActions length. Expected=%v, Got=%v", 0, len(ma))
	}

	// Register a new action
	ta = &testMA{}
	ma.Register(ta)

	// Keyword of test MA
	_, v = ma.handleArgs([]string{"workflow:test"}, DefaultMagicPrefix)
	if v != true {
		t.Errorf("Bad handled. Expected=%v, Got=%v", true, v)
	}

	if err := ta.ValidateRun(); err != nil {
		t.Errorf("Bad MagicAction: %v", err)
	}
}
