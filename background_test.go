//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-08
//

package aw

import (
	"os/exec"
	"testing"

	"github.com/deanishe/awgo/util"
)

// TestRunInBackground ensures background jobs work.
func TestRunInBackground(t *testing.T) {

	wf := New()

	cmd := exec.Command("sleep", "5")
	if wf.IsRunning("sleep") {
		t.Fatalf("Job 'sleep' is already running")
	}
	if err := wf.RunInBackground("sleep", cmd); err != nil {
		t.Fatalf("Error starting job 'sleep': %s", err)
	}
	if !wf.IsRunning("sleep") {
		t.Fatalf("Job 'sleep' is not running")
	}
	p := wf.pidFile("sleep")
	if !util.PathExists(p) {
		t.Fatalf("No PID file for 'sleep'")
	}
	// Duplicate jobs fail
	cmd = exec.Command("sleep", "5")
	err := wf.RunInBackground("sleep", cmd)
	if err == nil {
		t.Fatal("Starting duplicate 'sleep' job didn't error")
	}
	if _, ok := err.(ErrJobExists); !ok {
		t.Fatal("RunInBackground didn't return ErrAlreadyRunning")
	}
	if !IsJobExists(err) {
		t.Errorf("IsAlreadyRunning didn't identify ErrAlreadyRunning")
	}
	// Job killed OK
	if err := wf.Kill("sleep"); err != nil {
		t.Fatalf("Error killing 'sleep' job: %s", err)
	}
	if wf.IsRunning("sleep") {
		t.Fatal("'sleep' job still running")
	}
	if util.PathExists(p) {
		t.Fatal("'sleep' PID file not deleted")
	}
}
