// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/deanishe/awgo/util"
)

// ErrJobExists is the error returned by RunInBackground if a job with
// the given name is already running.
type ErrJobExists struct {
	Name string // Name of the job
	Pid  int    // PID of the running job
}

// Error implements error interface.
func (err ErrJobExists) Error() string {
	return fmt.Sprintf(`job "%s" already running with PID %d`, err.Name, err.Pid)
}

// Is returns true if target is of type ErrJobExists.
func (err ErrJobExists) Is(target error) bool {
	_, ok := target.(ErrJobExists)
	return ok
}

// IsJobExists returns true if error is of type or wraps ErrJobExists.
func IsJobExists(err error) bool {
	return errors.Is(err, ErrJobExists{})
}

// RunInBackground executes cmd in the background. It returns an
// ErrJobExists error if a job of the same name is already running.
func (wf *Workflow) RunInBackground(jobName string, cmd *exec.Cmd) error {
	if wf.IsRunning(jobName) {
		pid, _ := wf.getPid(jobName)
		return ErrJobExists{jobName, pid}
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Prevent process from being killed when parent is
	cmd.SysProcAttr.Setpgid = true
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("execute command %v: %w", cmd, err)
	}

	return wf.savePid(jobName, cmd.Process.Pid)
}

// Kill stops a background job.
func (wf *Workflow) Kill(jobName string) error {
	pid, err := wf.getPid(jobName)
	if err != nil {
		return err
	}
	p := wf.pidFile(jobName)
	err = syscall.Kill(pid, syscall.SIGTERM)
	os.Remove(p)
	return err
}

// IsRunning returns true if a job with name jobName is currently running.
func (wf *Workflow) IsRunning(jobName string) bool {
	pid, err := wf.getPid(jobName)
	if err != nil {
		return false
	}
	if err = syscall.Kill(pid, 0); err != nil {
		// Delete stale PID file
		os.Remove(wf.pidFile(jobName))
		return false
	}
	return true
}

// Save PID to a job-specific file.
func (wf *Workflow) savePid(jobName string, pid int) error {
	return ioutil.WriteFile(wf.pidFile(jobName), []byte(strconv.Itoa(pid)), 0600)
}

// Return PID for job.
func (wf *Workflow) getPid(jobName string) (int, error) {
	data, err := ioutil.ReadFile(wf.pidFile(jobName))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// Path to PID file for job.
func (wf *Workflow) pidFile(jobName string) string {
	dir := util.MustExist(filepath.Join(wf.awCacheDir(), "jobs"))
	return filepath.Join(dir, jobName+".pid")
}
