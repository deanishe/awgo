//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-12
//

package aw

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var testArgs = []struct{ in, out []string }{
	{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
}

var testMagic = []struct{ in, out []string }{
	{[]string{"workflow:invalid", "b", "c"}, []string{"b", "c"}},
	{[]string{"workflow:log", "b", "c"}, []string{"b", "c"}},
}

// ssEq tests if 2 string slices are equal.
func ssEq(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestNonMagicArgs tests that normal arguments aren't ignored
func TestNonMagicArgs(t *testing.T) {
	for _, td := range testArgs {
		args := parseArgs(td.in, DefaultMagicPrefix)
		if !ssEq(args, td.out) {
			t.Errorf("not equal. Expected=%v, Got=%v", td.out, args)
		}
	}
}

// TestMagicArgs tests that normal arguments aren't ignored
func TestMagicArgs(t *testing.T) {
	if os.Getenv("MAGIC") != "" {
		args := strings.Split(os.Getenv("ARGS"), ":")
		args = parseArgs(args, DefaultMagicPrefix)
		// log.Printf("args=%v", args)
		return
	}

	for _, td := range testMagic {
		cmd := exec.Command(os.Args[0], "-test.run=TestMagicArgs")
		args := "ARGS=" + strings.Join(td.in, ":")
		cmd.Env = append(os.Environ(), "MAGIC=1", args)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Errorf("couldn't get STDOUT of command (%v): %v", cmd, err)
		}

		err = cmd.Start()
		if err != nil {
			t.Errorf("couldn't run magic args \"%v\": %v", td.in, err)
		}
		out, err := ioutil.ReadAll(stdout)
		if err != nil {
			t.Errorf("couldn't read workflow JSON: %v", err)
		}

		s := fmt.Sprintf("%s", out)
		if !strings.HasPrefix(s, "PASS") {
			t.Errorf("magic command failed %v: %v", td.in, s)
		}
		err = cmd.Wait()
		if err != nil {
			t.Errorf("error running test command: %v", err)
		}
	}
}
