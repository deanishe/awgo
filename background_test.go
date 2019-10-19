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
	"github.com/stretchr/testify/require"

	"github.com/deanishe/awgo/util"
)

// Background jobs.
func TestWorkflow_RunInBackground(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		jobName := "sleep"

		cmd := exec.Command("sleep", "5")
		// Sanity check
		require.False(t, wf.IsRunning(jobName), "job already running")

		// Start job
		require.Nil(t, wf.RunInBackground(jobName, cmd), "start job failed")

		// Job running?
		assert.True(t, wf.IsRunning(jobName), "job is not running")

		pid, err := wf.getPid(jobName)
		assert.Nil(t, err, "get PID failed")

		p := wf.pidFile(jobName)
		assert.True(t, util.PathExists(p), "PID file does not exist")

		// Duplicate job fails?
		cmd = exec.Command("sleep", "5")
		err = wf.RunInBackground("sleep", cmd)
		require.NotNil(t, err, "start duplicate job did not fail")
		_, ok := err.(ErrJobExists)
		require.True(t, ok, "RunInBackground did not return ErrJobExists")
		assert.True(t, IsJobExists(err), "IsJobExist failed to identify ErrJobExists")
		assert.NotEqual(t, -1, strings.Index(err.Error(), fmt.Sprintf("%d", pid)), "PID not found in error")

		// Job killed OK?
		require.Nil(t, wf.Kill(jobName), "kill job failed")

		// Killing dead job fails?
		require.NotNil(t, wf.Kill(jobName), "kill dead job succeeded")

		// Job has exited and tidied up?
		assert.False(t, wf.IsRunning(jobName), "job still running")
		assert.False(t, util.PathExists(p), "PID file not deleted")

		// Invalid PID returns error?
		err = ioutil.WriteFile(p, []byte("bad PID"), 0600)
		require.Nil(t, err, "write PID file failed")

		assert.NotNil(t, wf.Kill(jobName), "invalid PID did not cause error")
	})
}

// invalid command fails
func TestWorkflow_RunInBackground_badJob(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		cmd := exec.Command("/does/not/exist")
		assert.NotNil(t, wf.RunInBackground("badJob", cmd), `run "/does/not/exist" succeeded`)
	})
}
