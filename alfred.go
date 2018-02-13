//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-11
//

package aw

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deanishe/awgo/util"
)

// Alfred is a wrapper for Alfred's AppleScript API. With the API, you can
// open Alfred in various modes and manipulate persistent workflow variables,
// i.e. the values saved in info.plist.
//
// Because calling Alfred is slow, the API uses a "Doer" interface, where
// commands are collected and all sent together when Alfred.Do() is called:
//
//     // Open Alfred
//     a := NewAlfred()
//     if err := a.Search("").Do(); err != nil {
//         // handle error
//     }
//
//     // Browse /Applications
//     a = NewAlfred()
//     if err := a.Browse("/Applications").Do(); err != nil {
//         // handle error
//     }
//
//     // Set multiple configuration values
//     a = NewAlfred()
//
//     a.SetConfig("USERNAME", "dave")
//     a.SetConfig("CLEARANCE", "highest")
//     a.SetConfig("PASSWORD", "hunter2")
//
//     if err := a.Do(); err != nil {
//         // handle error
//     }
//
// The BundleID is used as a default for methods that require a bundleID (e.g.
// RunTrigger).
type Alfred struct {
	// Default bundle ID for methods that require one
	BundleID string
	scripts  []string
	err      error
}

// NewAlfred creates a new Alfred using the bundle ID from the environment.
func NewAlfred() *Alfred { return &Alfred{BundleID(), []string{}, nil} }

// Do calls Alfred and runs the accumulated actions.
//
// If an error was encountered while preparing any commands, it will be
// returned here. It also returns an error if there are no commands to run,
// or if the call to Alfred fails.
//
// Succeed or fail, any accumulated scripts and errors are cleared when Do()
// is called.
func (a *Alfred) Do() error {

	var err error

	if a.err != nil {
		// reset
		err, a.err = a.err, nil
		a.scripts = []string{}

		return err
	}

	if len(a.scripts) == 0 {
		return errors.New("no commands to run")
	}

	script := strings.Join(a.scripts, "\n")
	// reset
	a.scripts = []string{}

	// log.Printf("-----------\n%s\n------------", script)

	_, err = util.RunJS(script)

	return err
}

// Search runs Alfred with the given query. Use an empty query to just open Alfred.
func (a *Alfred) Search(query string) *Alfred {
	return a.addScript(scriptSearch, query)
}

// Browse tells Alfred to open path in navigation mode.
func (a *Alfred) Browse(path string) *Alfred {

	path, err := filepath.Abs(path)
	if err != nil {
		a.err = err
		return a
	}

	return a.addScript(scriptBrowse, path)
}

// SetTheme tells Alfred to use the specified theme.
func (a *Alfred) SetTheme(name string) *Alfred {
	return a.addScript(scriptSetTheme, name)
}

// Action tells Alfred to show File Actions for path(s).
func (a *Alfred) Action(path ...string) *Alfred {

	if len(path) == 0 {
		return a
	}

	var paths []string

	for _, p := range path {

		p, err := filepath.Abs(p)
		if err != nil {
			a.err = fmt.Errorf("[action] couldn't make path absolute (%s): %v", p, err)
			continue
		}

		paths = append(paths, p)
	}

	script := fmt.Sprintf(scriptAction, util.QuoteJS(paths))

	a.scripts = append(a.scripts, script)

	return a
}

// RunTrigger runs an External Trigger in the given workflow. Query may be empty.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose trigger should be run.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a *Alfred) RunTrigger(name, query string, bundleID ...string) *Alfred {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	if query != "" {
		opts["withArgument"] = query
	}

	return a.addScriptOpts(scriptTrigger, name, opts)
}

// SetConfig saves a workflow variable to info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a *Alfred) SetConfig(key, value string, export bool, bundleID ...string) *Alfred {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"toValue":    value,
		"inWorkflow": bid,
		"exportable": export,
	}

	return a.addScriptOpts(scriptSetConfig, key, opts)
}

// RemoveConfig removes a workflow variable from info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a *Alfred) RemoveConfig(key string, bundleID ...string) *Alfred {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	return a.addScriptOpts(scriptRmConfig, key, opts)
}

// Add a JavaScript that takes a single argument.
func (a *Alfred) addScript(script, arg string) *Alfred {

	script = fmt.Sprintf(script, util.QuoteJS(arg))
	a.scripts = append(a.scripts, script)

	return a
}

// Run a JavaScript that takes two arguments, a string and an object.
func (a *Alfred) addScriptOpts(script, name string, opts map[string]interface{}) *Alfred {

	script = fmt.Sprintf(script, util.QuoteJS(name), util.QuoteJS(opts))
	a.scripts = append(a.scripts, script)

	return a
}

// Extract bundle ID from argument, Alfred.BundleID or environment (via BundleID()).
func (a *Alfred) getBundleID(bundleID ...string) string {

	if len(bundleID) > 0 {
		return bundleID[0]
	}

	if a.BundleID != "" {
		return a.BundleID
	}

	return BundleID()
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
