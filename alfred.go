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

	"github.com/deanishe/awgo/util"
)

// Alfred is a wrapper for Alfred's AppleScript API.
//
// The methods open Alfred in various states, with various input,
// and also allow you to manipulate persistent workflow variables,
// i.e. the values saved in info.plist.
//
// The BundleID is used as a default for methods that require a
// bundleID (e.g. RunTrigger).
type Alfred struct {
	// Default bundle ID for methods that require one
	BundleID string
}

// NewAlfred creates a new Alfred using the bundle ID from the environment.
func NewAlfred() Alfred { return Alfred{BundleID()} }

// Search runs Alfred with the given query. Use an empty query to just open Alfred.
func (a Alfred) Search(query string) error { return a.runScript(scriptSearch, query) }

// Browse tells Alfred to open path in navigation mode.
func (a Alfred) Browse(path string) error {

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return a.runScript(scriptBrowse, path)
}

// SetTheme tells Alfred to use the specified theme.
func (a Alfred) SetTheme(name string) error { return a.runScript(scriptSetTheme, name) }

// Action tells Alfred to show File Actions for path(s).
func (a Alfred) Action(path ...string) error {

	if len(path) == 0 {
		return errors.New("Action requires at least one path")
	}

	var paths []string
	for _, p := range path {
		p, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("couldn't make path absolute (%s): %v", p, err)
		}
		paths = append(paths, p)
	}

	script := fmt.Sprintf(scriptAction, util.QuoteJS(paths))

	_, err := util.RunJS(script)

	return err
}

// RunTrigger runs an External Trigger in the given workflow. Query may be empty.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose trigger should be run.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a Alfred) RunTrigger(name, query string, bundleID ...string) error {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	if query != "" {
		opts["withArgument"] = query
	}

	return a.runScriptOpts(scriptTrigger, name, opts)
}

// SetConfig saves a workflow variable to info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a Alfred) SetConfig(key, value string, export bool, bundleID ...string) error {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"toValue":    value,
		"inWorkflow": bid,
		"exportable": export,
	}

	return a.runScriptOpts(scriptSetConfig, key, opts)
}

// RemoveConfig removes a workflow variable from info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, the ID defaults to Alfred.BundleID or the bundle ID
// from the environment.
func (a Alfred) RemoveConfig(key string, bundleID ...string) error {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	return a.runScriptOpts(scriptRmConfig, key, opts)
}

// Run a JavaScript that takes a single argument.
func (a Alfred) runScript(script, arg string) error {

	script = fmt.Sprintf(script, util.QuoteJS(arg))

	_, err := util.RunJS(script)

	return err
}

// Run a JavaScript that takes two arguments, a string and an object.
func (a Alfred) runScriptOpts(script, name string, opts map[string]interface{}) error {

	script = fmt.Sprintf(script, util.QuoteJS(name), util.QuoteJS(opts))

	_, err := util.RunJS(script)

	return err
}

// Extract bundle ID from argument, Alfred.BundleID or environment (via BundleID()).
func (a Alfred) getBundleID(bundleID ...string) string {

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
