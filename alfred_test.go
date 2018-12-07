// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"
)

// Whether to run tests that actually call Alfred
var (
	testSearch   = false
	testAction   = false
	testBrowse   = false
	testTrigger  = false
	testSetTheme = false
)

func TestAlfred(t *testing.T) {

	a := NewAlfred()

	if testSearch {

		if err := a.Search(""); err != nil {
			t.Error(err)
		}

		if err := a.Search("awgo alfred"); err != nil {
			t.Error(err)
		}
	}

	if testAction {

		h := os.ExpandEnv("$HOME")

		if err := a.Action(h+"/Desktop", "."); err != nil {
			t.Error(err)
		}
	}

	if testBrowse {

		if err := a.Browse("."); err != nil {
			t.Error(err)
		}
	}

	if testTrigger {

		if err := a.RunTrigger("test", "AwGo, yo!"); err != nil {
			t.Error(err)
		}
	}

	if testSetTheme {

		if err := a.SetTheme("Alfred Notepad"); err != nil {
			t.Error(err)
		}
	}
}
