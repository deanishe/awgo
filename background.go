//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-08
//

package aw

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// AlreadyRunning is the error returned by RunInBackground if a job with
// the given name is already running.
type AlreadyRunning struct {
	Name string
	Pid  int
}

// Error implements error interface.
func (a AlreadyRunning) Error() string {
	return fmt.Sprintf("Job '%s' already running with PID %d", a.Name, a.Pid)
}

// RunInBackground executes cmd in the background. It returns an
// AlreadyRunning error if a job of the same name is already running.
func RunInBackground(jobName string, cmd *exec.Cmd) error {
	if IsRunning(jobName) {
		pid, _ := getPid(jobName)
		return AlreadyRunning{jobName, pid}
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Prevent process from being killed when parent is
	cmd.SysProcAttr.Setpgid = true
	if err := cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	if err := savePid(jobName, pid); err != nil {
		return err
	}
	return nil
}

// Kill stops a background job.
func Kill(jobName string) error {
	pid, err := getPid(jobName)
	if err != nil {
		return err
	}
	p := pidFile(jobName)
	if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
		// Delete stale PID file
		os.Remove(p)
		return err
	}
	os.Remove(p)
	return nil
}

// IsRunning returns true if a job with name jobName is currently running.
func IsRunning(jobName string) bool {
	pid, err := getPid(jobName)
	if err != nil {
		return false
	}
	if err = syscall.Kill(pid, 0); err != nil {
		// Delete stale PID file
		os.Remove(pidFile(jobName))
		return false
	}
	return true
}

// Save PID a job-specific file.
func savePid(jobName string, pid int) error {
	p := pidFile(jobName)
	if err := ioutil.WriteFile(p, []byte(fmt.Sprintf("%d", pid)), 0600); err != nil {
		return err
	}
	return nil
}

// Return PID for job.
func getPid(jobName string) (int, error) {
	p := pidFile(jobName)
	data, err := ioutil.ReadFile(p)
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
func pidFile(jobName string) string {
	dir := EnsureExists(filepath.Join(awCacheDir(), "jobs"))
	return filepath.Join(dir, jobName+".pid")
}
