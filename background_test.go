// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	"github.com/deanishe/awgo/util"
)

// Background jobs.
func TestWorkflow_RunInBackground(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {

		jobName := "sleep"

		cmd := exec.Command("sleep", "5")
		// Sanity check
		if wf.IsRunning(jobName) {
			t.Fatalf("Job %q is already running", jobName)
		}

		// Start job
		if err := wf.RunInBackground(jobName, cmd); err != nil {
			t.Fatalf("Error starting job %q: %s", jobName, err)
		}

		// Job running?
		if !wf.IsRunning(jobName) {
			t.Fatalf("Job %q is not running", jobName)
		}
		pid, err := wf.getPid(jobName)
		if err != nil {
			t.Fatalf("get PID for job: %v", err)
		}
		p := wf.pidFile(jobName)
		if !util.PathExists(p) {
			t.Fatalf("No PID file for %q", jobName)
		}

		// Duplicate job fails?
		cmd = exec.Command("sleep", "5")
		err = wf.RunInBackground("sleep", cmd)
		if err == nil {
			t.Fatal("Starting duplicate 'sleep' job didn't error")
		}
		if _, ok := err.(ErrJobExists); !ok {
			t.Fatal("RunInBackground didn't return ErrAlreadyRunning")
		}
		if !IsJobExists(err) {
			t.Errorf("IsAlreadyRunning didn't identify ErrAlreadyRunning")
		}
		if strings.Index(err.Error(), fmt.Sprintf("%d", pid)) == -1 {
			t.Errorf(`PID not found in error`)
		}

		// Job killed OK?
		if err := wf.Kill("sleep"); err != nil {
			t.Fatalf("Error killing job %q: %s", jobName, err)
		}
		// Killing dead job fails?
		if err := wf.Kill("sleep"); err == nil {
			t.Fatalf("No error killing dead job %q", jobName)
		}

		// Job has exited and tidied up?
		if wf.IsRunning("sleep") {
			t.Fatalf("%q job still running", jobName)
		}
		if util.PathExists(p) {
			t.Fatalf("PID file for %q not deleted", jobName)
		}

		// Invalid PID returns error?
		if err := ioutil.WriteFile(p, []byte("bad PID"), 0600); err != nil {
			t.Fatalf("failed to write PID file %q: %v", p, err)
		}

		if err := wf.Kill(jobName); err == nil {
			t.Fatal("invalid PID did not cause error")
		}
	})
}

// invalid command fails
func TestWorkflow_RunInBackground_badJob(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		cmd := exec.Command("/does/not/exist")
		if err := wf.RunInBackground("badJob", cmd); err == nil {
			t.Fatal(`run "/does/not/exist" succeeded`)
		}
	})
}
