//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-13
//

package aw

import "os"

// Context contains Alfred and workflow settings extracted from environment variables set by Alfred.
type Context struct {
	// Version of workflow (from workflow configuration sheet)
	WorkflowVersion string
	// Name of workflow (from configuration sheet)
	Name string
	// Bundle ID (from configuration sheet)
	BundleID string
	// UID assigned to workflow by Alfred
	UID string
	// true if Alfred's debugger is open
	Debug bool
	// Alfred's version string
	AlfredVersion string
	// Alfred's build
	AlfredBuild string
	// ID of user's selected theme
	Theme string
	// Theme's background colour in rgba format, e.g. "rgba(255,255,255,1.0)"
	ThemeBackground string
	// Theme's selection background colour in rgba format
	ThemeSelectionBackground string
	// Path to "Alfred.alfredpreferences" file
	Preferences string
	// Machine-specific hash. Machine preferences are stored in
	// Alfred.alfredpreferences/local/<hash>
	Localhash string
	// Path to workflow's cache directory. Use Workflow.CacheDir() instead
	CacheDir string
	// Path to workflow's data directory. Use Workflow.DataDir() instead
	DataDir string
}

// NewContext creates a new Context initialised from Alfred's environment variables.
func NewContext() *Context {
	return &Context{
		WorkflowVersion:          os.Getenv("alfred_workflow_version"),
		Name:                     os.Getenv("alfred_workflow_name"),
		BundleID:                 os.Getenv("alfred_workflow_bundleid"),
		UID:                      os.Getenv("alfred_workflow_uid"),
		Debug:                    os.Getenv("alfred_debug") == "1",
		AlfredVersion:            os.Getenv("alfred_version"),
		AlfredBuild:              os.Getenv("alfred_version_build"),
		Theme:                    os.Getenv("alfred_theme"),
		ThemeBackground:          os.Getenv("alfred_theme_background"),
		ThemeSelectionBackground: os.Getenv("alfred_theme_selection_background"),
		Preferences:              os.Getenv("alfred_preferences"),
		Localhash:                os.Getenv("alfred_preferences_localhash"),
		CacheDir:                 os.Getenv("alfred_workflow_cache"),
		DataDir:                  os.Getenv("alfred_workflow_data"),
	}
}
