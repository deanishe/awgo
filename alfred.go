// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"path/filepath"

	"github.com/deanishe/awgo/util"
)

/*
Alfred wraps Alfred's AppleScript API, allowing you to open Alfred in
various modes or call External Triggers.

	a := NewAlfred()

	// Open Alfred
	if err := a.Search(""); err != nil {
		// handle error
	}

	// Browse /Applications
	if err := a.Browse("/Applications"); err != nil {
		// handle error
	}
*/
type Alfred struct {
	Env
}

// NewAlfred creates a new Alfred from the environment.
//
// It accepts one optional Env argument. If an Env is passed, Alfred
// is initialised from that instead of the system environment.
func NewAlfred(env ...Env) *Alfred {

	var e Env

	if len(env) > 0 {
		e = env[0]
	} else {
		e = sysEnv{}
	}

	return &Alfred{e}
}

// Search runs Alfred with the given query. Use an empty query to just open Alfred.
func (a *Alfred) Search(query string) error {
	_, err := util.RunJS(fmt.Sprintf(scriptSearch, util.QuoteJS(query)))
	return err
}

// Browse tells Alfred to open path in navigation mode.
func (a *Alfred) Browse(path string) error {

	var err error

	if path, err = filepath.Abs(path); err != nil {
		return err
	}

	_, err = util.RunJS(fmt.Sprintf(scriptBrowse, util.QuoteJS(path)))
	return err
}

// SetTheme tells Alfred to use the specified theme.
func (a *Alfred) SetTheme(name string) error {
	_, err := util.RunJS(fmt.Sprintf(scriptSetTheme, util.QuoteJS(name)))
	return err
}

// Action tells Alfred to show File Actions for path(s).
func (a *Alfred) Action(path ...string) error {

	if len(path) == 0 {
		return nil
	}

	var paths []string

	for _, p := range path {

		p, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("[action] couldn't make path absolute (%s): %v", p, err)
		}

		paths = append(paths, p)
	}

	_, err := util.RunJS(fmt.Sprintf(scriptAction, util.QuoteJS(paths)))
	return err
}

// RunTrigger runs an External Trigger in the given workflow. Query may be empty.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose trigger should be run.
// If not specified, it defaults to the current workflow's.
func (a *Alfred) RunTrigger(name, query string, bundleID ...string) error {

	var bid string
	if len(bundleID) > 0 {
		bid = bundleID[0]
	} else {
		bid, _ = a.Lookup(EnvVarBundleID)
	}

	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	if query != "" {
		opts["withArgument"] = query
	}

	_, err := util.RunJS(fmt.Sprintf(scriptTrigger, util.QuoteJS(name), util.QuoteJS(opts)))
	return err
}

// JXA scripts to call Alfred
var (
	// Simple scripts require one or no string
	scriptSearch   = "Application('Alfred 3').search(%s)"
	scriptAction   = "Application('Alfred 3').action(%s)"
	scriptBrowse   = "Application('Alfred 3').browse(%s)"
	scriptSetTheme = "Application('Alfred 3').setTheme(%s)"
	// These scripts require a string and an object of options
	scriptTrigger   = "Application('Alfred 3').runTrigger(%s, %s)"
	scriptSetConfig = "Application('Alfred 3').setConfiguration(%s, %s)"
	scriptRmConfig  = "Application('Alfred 3').removeConfiguration(%s, %s)"
)
