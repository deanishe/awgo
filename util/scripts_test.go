// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutableRunner(t *testing.T) {
	t.Parallel()

	data := []struct {
		in string
		x  bool
	}{
		{"", false},
		{"non-existent", false},
		// Directories
		{"/Applications", false},
		{"/var", false}, // symlink on macOS
		{"/", false},
		{"/bin", false},
		// Existing paths
		{"/usr/bin/python2.7", true}, // symlink on El Cap
		{"/bin/cp", true},
		{"/bin/ls", true},
		{"/bin/mv", true},
	}

	r := ExecRunner{}

	for _, td := range data {
		td := td // capture variable
		t.Run(fmt.Sprintf("CanRun(%s)", td.in), func(t *testing.T) {
			t.Parallel()
			v := r.CanRun(td.in)
			if v != td.x {
				t.Errorf("Bad CanRun for %#v. Expected=%v, Got=%v", td.in, td.x, v)
			}
		})
	}
}

func TestScriptRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		valid bool
	}{
		{"testdata/applescript.applescript", true},
		{"testdata/applescript.scpt", true},
		{"testdata/bash.sh", true},
		{"testdata/jxa.js", true},
		{"testdata/python.py", true},
		// ScriptRunner can't run executables
		{"testdata/bash_exe", false},
		{"testdata/python_exe", false},
		// Not scripts
		{"testdata/non-executable", false},
		{"testdata/non-existent", false},
		{"testdata/plain.txt", false},
		{"testdata/perl.pl", false},
	}

	r := ScriptRunner{DefaultInterpreters}
	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("CanRun(%s)", td.in), func(t *testing.T) {
			t.Parallel()
			v := r.CanRun(td.in)
			if v != td.valid {
				t.Errorf("Expected=%v, Got=%v", td.valid, v)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	scripts := []string{
		"testdata/applescript.applescript",
		"testdata/applescript.scpt",
		"testdata/bash.sh",
		"testdata/bash_exe",
		"testdata/jxa.js",
		"testdata/python.py",
		"testdata/python_exe",
	}

	for _, script := range scripts {
		script := script // capture variable
		t.Run(fmt.Sprintf("Run(%s)", script), func(t *testing.T) {
			t.Parallel()
			x := filepath.Base(script)
			out, err := Run(script, x)
			if err != nil {
				t.Errorf("failed: %v", err)
			}
			v := strings.TrimSpace(string(out))
			if v != x {
				t.Errorf("Bad output. Expected=%v, Got=%v", x, v)
			}
		})
	}
}

func TestNoRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		unknown bool
		missing bool
	}{
		{"testdata/non-executable", true, false},
		{"testdata/non-existent", false, true},
		{"testdata/plain.txt", true, false},
		{"testdata/perl.pl", true, false},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("Run(%s)", td.in), func(t *testing.T) {
			t.Parallel()
			_, err := Run(td.in, "blah")
			if err == nil {
				t.Errorf("Ran bad script %q", td.in)
			}
			if td.unknown && err != ErrUnknownFileType {
				t.Errorf("Unknown file recognised %q. Expected=%v, Got=%v", td.in, ErrUnknownFileType, err)
			}
			if td.missing && !os.IsNotExist(err) {
				t.Errorf("Missing file found %q. Expected=ErrNotExist, Got=%v", td.in, err)
			}
		})
	}
}

// TestNewScriptRunner verifies that ScriptRunner accepts the correct filetypes.
func TestNewScriptRunner(t *testing.T) {
	t.Parallel()

	data := []struct {
		good, bad int
		m         map[string][]string
	}{
		// Python scripts only
		{1, 6, map[string][]string{
			".py": []string{"/usr/bin/python"},
		}},
		// AppleScripts
		{3, 4, map[string][]string{
			".scpt":        []string{"/usr/bin/osascript"},
			".applescript": []string{"/usr/bin/osascript"},
			".js":          []string{"/usr/bin/osascript", "-l", "JavaScript"},
		}},
	}

	scripts := []string{
		"testdata/applescript.applescript",
		"testdata/applescript.scpt",
		"testdata/bash.sh",
		"testdata/bash_exe",
		"testdata/jxa.js",
		"testdata/python.py",
		"testdata/python_exe",
	}

	for i, td := range data {
		td := td // capture variable
		t.Run(fmt.Sprintf("ScriptRunner(%d)", i), func(t *testing.T) {
			t.Parallel()

			r := NewScriptRunner(td.m)
			var good, bad int

			for _, script := range scripts {
				if v := r.CanRun(script); v {
					good++
				} else {
					bad++
				}
			}

			if good != td.good {
				t.Errorf("Bad good. Expected=%d, Got=%d", td.good, good)
			}
			if bad != td.bad {
				t.Errorf("Bad bad. Expected=%d, Got=%d", td.bad, bad)
			}
		})
	}
}

// TestQuoteJS verifies QuoteJS quoting.
func TestQuoteJS(t *testing.T) {
	t.Parallel()

	data := []struct {
		in  interface{}
		out string
	}{
		{"", `""`},
		{"onions", `"onions"`},
		{"", `""`},
		{[]string{"one", "two", "three"}, `["one","two","three"]`},
	}

	for _, td := range data {

		s := QuoteJS(td.in)

		if s != td.out {
			t.Errorf("Bad JS for %#v. Expected=%v, Got=%v", td.in, td.out, s)
		}

	}
}
