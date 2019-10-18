// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deanishe/awgo/util"
)

// Background jobs.
func TestWorkflow_RunInBackground(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {

		jobName := "sleep"

		cmd := exec.Command("sleep", "5")
		// Sanity check
		assert.False(t, wf.IsRunning(jobName), "job %q is already running", jobName)

		// Start job
		assert.Nil(t, wf.RunInBackground(jobName, cmd), "failed to start job %q", jobName)

		// Job running?
		assert.True(t, wf.IsRunning(jobName), "job %q is not running", jobName)

		pid, err := wf.getPid(jobName)
		assert.Nil(t, err, "get PID for job %q failed", jobName)

		p := wf.pidFile(jobName)
		assert.True(t, util.PathExists(p), "no PID file for %q", jobName)

		// Duplicate job fails?
		cmd = exec.Command("sleep", "5")
		err = wf.RunInBackground("sleep", cmd)
		assert.NotNil(t, err, "start duplicate job did not fail")

		_, ok := err.(ErrJobExists)
		assert.True(t, ok, "RunInBackground didn't return ErrAlreadyRunning")
		assert.True(t, IsJobExists(err), "IsAlreadyRunning did not identity ErrAlreadyRunning")

		assert.NotEqual(t, -1, strings.Index(err.Error(), fmt.Sprintf("%d", pid)), "PID not found in error")

		// Job killed OK?
		assert.Nil(t, wf.Kill("sleep"), "failed to kill job")

		// Killing dead job fails?
		assert.NotNil(t, wf.Kill("sleep"), "no error killing dead job %q", jobName)

		// Job has exited and tidied up?
		assert.False(t, wf.IsRunning("sleep"), "job %q still running", jobName)
		assert.False(t, util.PathExists(p), "PID file for %q not deleted", jobName)

		// Invalid PID returns error?
		assert.Nil(t, ioutil.WriteFile(p, []byte("bad PID"), 0600), "failed to write PID file")
		assert.NotNil(t, wf.Kill(jobName), "invalid PID did not cause error")
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
