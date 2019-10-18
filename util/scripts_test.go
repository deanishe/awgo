// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutableRunner(t *testing.T) {
	t.Parallel()

	data := []struct {
		in    string
		valid bool
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
			assert.Equal(t, td.valid, r.CanRun(td.in), "unexpected validity")

			// Also test runners
			cmd := runners.Cmd(td.in)
			if td.valid {
				assert.NotNil(t, cmd, "valid command rejected")
			} else {
				assert.Nil(t, cmd, "invalid command accepted")
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
			assert.Equal(t, td.valid, r.CanRun(td.in), "unexpected validity")
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

			// test Run
			out, err := Run(script, x)
			assert.Nil(t, err, "script  %q failed: %v", script, err)
			assert.Equal(t, strings.TrimSpace(string(out)), x, "bad output")

			// test runners
			out, err = runners.Run(script, x)
			assert.Nil(t, err, "script  %q failed: %v", script, err)
			assert.Equal(t, strings.TrimSpace(string(out)), x, "bad output")
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
		{"testdata", true, false},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("Run(%s)", td.in), func(t *testing.T) {
			t.Parallel()

			_, err := Run(td.in, "blah")
			assert.NotNil(t, err, "ran invalid script %q", td.in)
			if td.unknown {
				assert.Equal(t, ErrUnknownFileType, err, "invalid file recognised")
			}
			if td.missing {
				assert.True(t, os.IsNotExist(err), "non-existent file accepted")
			}

			_, err = runners.Run(td.in, "blah")
			assert.NotNil(t, err, "ran invalid script %q", td.in)
			if td.unknown {
				assert.Equal(t, ErrUnknownFileType, err, "invalid file recognised")
			}
			if td.missing {
				assert.True(t, os.IsNotExist(err), "non-existent file accepted")
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
			".py": {"/usr/bin/python"},
		}},
		// AppleScripts
		{3, 4, map[string][]string{
			".scpt":        {"/usr/bin/osascript"},
			".applescript": {"/usr/bin/osascript"},
			".js":          {"/usr/bin/osascript", "-l", "JavaScript"},
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
			assert.Equal(t, td.good, good, "unexpected good count")
			assert.Equal(t, td.bad, bad, "unexpected bad count")
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
		assert.Equal(t, td.out, QuoteJS(td.in), "unexpected quoted JS")
	}
}
