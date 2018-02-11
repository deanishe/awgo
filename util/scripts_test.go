//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/*
func TestAppleScriptify(t *testing.T) {
	data := []struct {
		in, out string
	}{
		{"", ""},
		{"simple", "simple"},
		{"with spaces", "with spaces"},
		{`has "quotes" within`, `has " & quote & "quotes" & quote & " within`},
		{`"within quotes"`, `" & quote & "within quotes" & quote & "`},
		{`"`, `" & quote & "`},
	}

	for _, td := range data {
		s := AppleScriptify(td.in)
		if s != td.out {
			t.Errorf("Bad AppleScript escape. Expected=%v, Got=%v", td.out, s)
		}

	}
}
*/

// Scripts in various language that write $1 to STDOUT.
var testScripts = []struct {
	name, script string
}{
	{"python.py", "import sys; print(sys.argv[1])"},
	{"python_exe", `#!/usr/bin/python
import sys
print(sys.argv[1])`},

	{"bash.sh", `echo "$1"`},
	{"bash_exe", `#!/bin/bash
echo "$1"`},

	{"applescript.scpt", `on run(argv)
	return first item of argv
end run`},
	{"applescript.applescript", `on run(argv)
	return first item of argv
end run`},
	{"jxa.js", `function run(argv) { return argv[0]; }`},
}

// Files that should be ignored
var invalidFiles = []struct {
	name, content string
}{
	{"word.doc", "blah"},
	{"plain.txt", "hello!"},
	{"non-executable", "dummy"},
	{"perl.pl", "$$@£@..21!55-"},
}

// Create and execute function in a directory containing testScripts.
func withTestScripts(fun func(names []string)) {

	err := inTempDir(func(dir string) {

		names := []string{}

		// Create test scripts
		for _, ts := range testScripts {

			names = append(names, ts.name)

			perms := os.FileMode(0600)

			// Make scripts w/o extensions executable
			if filepath.Ext(ts.name) == "" {
				perms = 0700
			}
			if err := ioutil.WriteFile(ts.name, []byte(ts.script), perms); err != nil {
				panic(err)
			}
		}

		for _, ts := range invalidFiles {

			names = append(names, ts.name)

			if err := ioutil.WriteFile(ts.name, []byte(ts.content), 0600); err != nil {
				panic(err)
			}
		}

		// Execute function
		fun(names)
	})

	if err != nil {
		panic(err)
	}
}

func TestExecutableRunnerCanRun(t *testing.T) {
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
		v := r.CanRun(td.in)
		if v != td.x {
			t.Errorf("Bad CanRun for %#v. Expected=%v, Got=%v", td.in, td.x, v)
		}
	}
}

func TestScriptRunnerCanRun(t *testing.T) {

	withTestScripts(func(names []string) {

		invalid := map[string]bool{}
		for _, ts := range invalidFiles {
			invalid[ts.name] = true
		}
		for _, ts := range testScripts { // can't run extension-less files
			if strings.Index(ts.name, ".") < 0 {
				invalid[ts.name] = true
			}
		}

		r := ScriptRunner{DefaultInterpreters}

		for _, name := range names {

			v := r.CanRun(name)
			if v && invalid[name] {
				t.Errorf("Invalid accepted: %s", name)
			}
			if !v && !invalid[name] {
				t.Errorf("Valid rejected: %s", name)
			}
		}

	})

}

func TestRun(t *testing.T) {

	withTestScripts(func(names []string) {

		invalid := map[string]bool{}
		for _, ts := range invalidFiles {
			invalid[ts.name] = true
		}

		var good, bad int
		// Execute scripts and compare output
		for _, name := range names {

			out, err := Run(name, name)
			if err != nil {

				if err == ErrUnknownFileType {
					if !invalid[name] {
						t.Errorf("Failed to run valid script (%s): %v", name, err)
						continue
					}
					// correctly rejected
					bad++
					continue
				}

				t.Fatalf("Error running %v: %v", name, err)
			}

			s := strings.TrimSpace(string(out))
			if s != name {
				t.Errorf("Bad output. Expected=%v, Got=%v", name, s)
			} else {
				good++
			}

		}

		if good != len(testScripts) {
			t.Errorf("Bad script count. Expected=%v, Got=%v", len(testScripts), good)
		}

		if bad != len(invalidFiles) {
			t.Errorf("Bad invalid file count. Expected=%v, Got=%v", len(invalidFiles), bad)
		}

	})
}

// TestNewScriptRunner verifies that ScriptRunner accepts the correct filetypes.
func TestNewScriptRunner(t *testing.T) {

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

	withTestScripts(func(names []string) {

		for _, td := range data {

			r := NewScriptRunner(td.m)
			var good, bad int

			for _, ts := range testScripts {

				if v := r.CanRun(ts.name); v {
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
		}

	})

}
