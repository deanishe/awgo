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

// Environment variables containing workflow and Alfred info.
//
// Read the values with os.Getenv(EnvVarName) or via Alfred:
//
//    // Returns a string
//    Alfred.Get(EnvVarName)
//    // Parse string into a bool
//    Alfred.GetBool(EnvVarDebug)
//
const (
	// Workflow info assigned in Alfred Preferences
	EnvVarName     = "alfred_workflow_name"     // Name of workflow
	EnvVarBundleID = "alfred_workflow_bundleid" // Bundle ID
	EnvVarVersion  = "alfred_workflow_version"  // Workflow version

	EnvVarUID = "alfred_workflow_uid" // Random UID assigned by Alfred

	// Workflow storage directories
	EnvVarCacheDir = "alfred_workflow_cache" // For temporary data
	EnvVarDataDir  = "alfred_workflow_data"  // For permanent data

	// Set to 1 when Alfred's debugger is open
	EnvVarDebug = "alfred_debug"

	// Theme info. Colours are in rgba format, e.g. "rgba(255,255,255,1.0)"
	EnvVarTheme            = "alfred_theme"                      // ID of user's selected theme
	EnvVarThemeBG          = "alfred_theme_background"           // Background colour
	EnvVarThemeSelectionBG = "alfred_theme_selection_background" // BG colour of selected item

	// Alfred info
	EnvVarAlfredVersion = "alfred_version"       // Alfred's version number
	EnvVarAlfredBuild   = "alfred_version_build" // Alfred's build number
	EnvVarPreferences   = "alfred_preferences"   // Path to "Alfred.alfredpreferences" file
	// Machine-specific hash. Machine preferences are stored in
	// Alfred.alfredpreferences/local/<hash>
	EnvVarLocalhash = "alfred_preferences_localhash"
)

/*
Alfred is a wrapper for Alfred's AppleScript API. With the API, you can
open Alfred in various modes and manipulate persistent workflow variables,
i.e. the values saved in info.plist.

Because calling Alfred is slow, the API uses a "Doer" interface, where
commands are collected and all sent together when Alfred.Do() is called:

	// Open Alfred
	a := NewAlfred()
	if err := a.Search("").Do(); err != nil {
		// handle error
	}

	// Browse /Applications
	a = NewAlfred()
	if err := a.Browse("/Applications").Do(); err != nil {
		// handle error
	}

	// Set multiple configuration values
	a = NewAlfred()

	a.SetConfig("USERNAME", "dave")
	a.SetConfig("CLEARANCE", "highest")
	a.SetConfig("PASSWORD", "hunter2")

	if err := a.Do(); err != nil {
		// handle error
	}

*/
type Alfred struct {
	Env
	bundleID string
	scripts  []string
	err      error
}

// NewAlfred creates a new Alfred from the environment.
//
// It accepts one optional Env argument. If an Env is passed, Alfred
// is initialised from that instead of the system environment.
func NewAlfred(env ...Env) *Alfred {

	var (
		a   *Alfred
		bid string
		e   Env
	)

	if len(env) > 0 {
		e = env[0]
	} else {
		e = sysEnv{}
	}

	if s, ok := e.Lookup("alfred_workflow_bundleid"); ok {
		bid = s
	}

	a = &Alfred{
		Env:      e,
		bundleID: bid,
		scripts:  []string{},
		err:      nil,
	}

	return a
}

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
// If not specified, it defaults to the current workflow's.
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
// If not specified, it defaults to the current workflow's.
func (a *Alfred) SetConfig(key, value string, export bool, bundleID ...string) *Alfred {

	bid := a.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"toValue":    value,
		"inWorkflow": bid,
		"exportable": export,
	}

	return a.addScriptOpts(scriptSetConfig, key, opts)
}

// setMulti is an internal wrapper around SetConfig and do. It implements
// the internal bindDest interface to make testing easier.
func (a *Alfred) setMulti(variables map[string]string, export bool) error {

	for k, v := range variables {
		a.SetConfig(k, v, export)
	}

	return a.Do()
}

// RemoveConfig removes a workflow variable from info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, it defaults to the current workflow's.
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

	return a.bundleID
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
