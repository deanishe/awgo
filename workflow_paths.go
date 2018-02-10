//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-09
//

package aw

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/deanishe/awgo/util"
)

// Dir returns the path to the workflow's root directory.
func Dir() string { return wf.Dir() }
func (wf *Workflow) Dir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

// CacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func CacheDir() string                { return wf.CacheDir() }
func (wf *Workflow) CacheDir() string { return util.MustExist(wf.Conf.Get(EnvVarCacheDir)) }

// OpenCache opens the workflow's cache directory in the default application (usually Finder).
func OpenCache() error { return wf.OpenCache() }
func (wf *Workflow) OpenCache() error {
	util.MustExist(wf.CacheDir())
	cmd := exec.Command("open", wf.CacheDir())
	return cmd.Run()
}

// ClearCache deletes all files from the workflow's cache directory.
func ClearCache() error { return wf.ClearCache() }
func (wf *Workflow) ClearCache() error {
	return util.ClearDirectory(wf.CacheDir())
}

// DataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func DataDir() string                { return wf.DataDir() }
func (wf *Workflow) DataDir() string { return util.MustExist(wf.Conf.Get(EnvVarDataDir)) }

// OpenData opens the workflow's data directory in the default application (usually Finder).
func OpenData() error { return wf.OpenData() }
func (wf *Workflow) OpenData() error {
	util.MustExist(wf.Conf.Get(EnvVarDataDir))
	cmd := exec.Command("open", wf.DataDir())
	return cmd.Run()
}

// ClearData deletes all files from the workflow's cache directory.
func ClearData() error { return wf.ClearData() }
func (wf *Workflow) ClearData() error {
	return util.ClearDirectory(wf.DataDir())
}

// Reset deletes all workflow data (cache and data directories).
func Reset() error { return wf.Reset() }
func (wf *Workflow) Reset() error {
	errs := []error{}
	if err := wf.ClearCache(); err != nil {
		errs = append(errs, err)
	}
	if err := wf.ClearData(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// LogFile returns the path to the workflow's log file.
func LogFile() string { return wf.LogFile() }
func (wf *Workflow) LogFile() string {
	return filepath.Join(wf.CacheDir(), fmt.Sprintf("%s.log", wf.BundleID()))
}

// OpenLog opens the workflow's logfile in the default application (usually Console.app).
func OpenLog() error { return wf.OpenLog() }
func (wf *Workflow) OpenLog() error {
	if !util.PathExists(wf.LogFile()) {
		log.Println("Creating log file...")
	}
	cmd := exec.Command("open", wf.LogFile())
	return cmd.Run()
}

func OpenHelp() error { return wf.OpenHelp() }
func (wf *Workflow) OpenHelp() error {
	if wf.HelpURL == "" {
		return errors.New("Help URL is not set")
	}
	cmd := exec.Command("open", wf.HelpURL)
	return cmd.Run()
}
