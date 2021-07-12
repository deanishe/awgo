// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deanishe/awgo/util"
)

// JXA scripts to call Alfred.
const (
	scriptSearch = "Application(%s).search(%s);"
	scriptAction = "Application(%s).action(%s);"
	// support "asType" option added in Alfred 4.5
	scriptActionType = "Application(%s).action(%s, %s);"
	scriptBrowse     = "Application(%s).browse(%s);"
	scriptSetTheme   = "Application(%s).setTheme(%s);"
	scriptTrigger    = "Application(%s).runTrigger(%s, %s);"
	scriptSetConfig  = "Application(%s).setConfiguration(%s, %s);"
	scriptRmConfig   = "Application(%s).removeConfiguration(%s, %s);"
	scriptReload     = "Application(%s).reloadWorkflow(%s);"
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
	// For testing. Set to true to save JXA script to lastScript
	// instead of running it.
	noRunScripts bool
	lastScript   string
}

// NewAlfred creates a new Alfred from the environment.
//
// It accepts one optional Env argument. If an Env is passed, Alfred
// is initialised from that instead of the system environment.
func NewAlfred(env ...Env) *Alfred {
	var e Env = sysEnv{}
	if len(env) > 0 {
		e = env[0]
	}

	return &Alfred{Env: e}
}

// Search runs Alfred with the given query. Use an empty query to just open Alfred.
func (a *Alfred) Search(query string) error {
	return a.runScript(scriptSearch, query)
}

// Browse tells Alfred to open path in navigation mode.
func (a *Alfred) Browse(path string) error {
	var err error

	if path, err = filepath.Abs(path); err != nil {
		return err
	}

	return a.runScript(scriptBrowse, path)
}

// SetTheme tells Alfred to use the specified theme.
func (a *Alfred) SetTheme(name string) error {
	return a.runScript(scriptSetTheme, name)
}

// Action tells Alfred to show Universal Actions for value(s). This calls Alfred.ActionAsType
// with an empty type.
func (a *Alfred) Action(value ...string) error { return a.ActionAsType("", value...) }

// ActionAsType tells Alfred to show Universal Actions for value(s). Type typ
// may be one of "file", "url" or "text", or an empty string to tell Alfred
// to guess the type.
//
// Added in Alfred 4.5
func (a *Alfred) ActionAsType(typ string, value ...string) error {
	if len(value) == 0 {
		return nil
	}

	if typ == TypeFile {
		for i, s := range value {
			p, err := filepath.Abs(s)
			if err != nil {
				return fmt.Errorf("make absolute path %q: %w", s, err)
			}
			value[i] = p
		}
	}

	switch typ {
	case "":
		return a.runScript(scriptAction, value)
	case "file", "url", "text":
		opts := map[string]interface{}{
			"asType": typ,
		}
		return a.runScript(scriptActionType, value, opts)
	default:
		return fmt.Errorf("unknown type: %s", typ)
	}
}

// RunTrigger runs an External Trigger in the given workflow. Query may be empty.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose trigger should be run.
// If not specified, it defaults to the current workflow's.
func (a *Alfred) RunTrigger(name, query string, bundleID ...string) error {
	bid, _ := a.Lookup(EnvVarBundleID)
	if len(bundleID) > 0 {
		bid = bundleID[0]
	}

	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	if query != "" {
		opts["withArgument"] = query
	}

	return a.runScript(scriptTrigger, name, opts)
}

// ReloadWorkflow tells Alfred to reload a workflow from disk.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow to reload. If not specified, it defaults to the current workflow's.
func (a *Alfred) ReloadWorkflow(bundleID ...string) error {
	bid, _ := a.Lookup(EnvVarBundleID)
	if len(bundleID) > 0 {
		bid = bundleID[0]
	}

	return a.runScript(scriptReload, bid)
}

func (a *Alfred) runScript(script string, arg ...interface{}) error {
	quoted := []interface{}{util.QuoteJS(scriptAppName())}
	for _, v := range arg {
		quoted = append(quoted, util.QuoteJS(v))
	}
	script = fmt.Sprintf(script, quoted...)

	if a.noRunScripts {
		a.lastScript = script
		return nil
	}

	_, err := util.RunJS(script)
	return err
}

// Name of JXA Application for running Alfred
func scriptAppName() string {
	// Alfred 3
	if strings.HasPrefix(os.Getenv(EnvVarAlfredVersion), "3") {
		return "Alfred 3"
	}
	// Alfred 4+
	return "com.runningwithcrayons.Alfred"
}
