// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

/*
Package aw is a "plug-and-play" workflow development library/framework for Alfred 3 & 4
(https://www.alfredapp.com/).

It provides everything you need to create a polished and blazing-fast Alfred
frontend for your project.


Features

As of AwGo 0.17, all applicable features of Alfred 4.0 are supported.

The main features are:

	- Full support for Alfred 3 & 4
	- Bi-directional interface to workflow's configuration
	- Fluent API for generating Script Filter JSON
	- Fuzzy filtering
	- Simple, powerful API for caching/saving workflow data
	- Keychain API to securely store (and sync) sensitive data
	- API to call Alfred's AppleScript methods from Go code
	- Helpers to easily run scripts and script code
	- Workflow update API with built-in support for GitHub & Gitea
	- Pre-configured logging for easier debugging, with a rotating log file
	- Catches panics, logs stack trace and shows user an error message
	- "Magic" queries/actions for simplified development and user support
	- Some default icons based on macOS system icons


Usage

AwGo is an opinionated framework that expects to be used in a certain way in
order to eliminate boilerplate. It *will* panic if not run in a valid,
minimally Alfred-like environment. At a minimum the following environment
variables should be set to meaningful values:

	// Absolutely required. No ifs or buts.
	alfred_workflow_bundleid

	// Cache & data dirs can be set to anything, but for best
	// results, point them at the same directories as Alfred uses
	// Alfred 3:  ~/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data/<bundle ID>/
	// Alfred 4+: ~/Library/Caches/com.runningwithcrayons.Alfred/Workflow Data/<bundle ID>/
	alfred_workflow_cache

	// Alfred 3:  ~/Library/Application Support/Alfred 3/Workflow Data/<bundle ID>/
	// Alfred 4+: ~/Library/Application Support/Alfred/Workflow Data/<bundle ID>/
	alfred_workflow_data

	// If you're using the Updater API, a semantic-ish workflow version
	// must be set otherwise the Updater will panic
    alfred_workflow_version

	// If you're using the Alfred API and running Alfred 3, you need to
	// set `alfred_version` as AwGo defaults to calling Alfred 4+
	alfred_version=3


NOTE: AwGo is currently in development. The API *will* change and should
not be considered stable until v1.0. Until then, be sure to pin a version
using go modules or similar.

Be sure to also check out the _examples/ subdirectory, which contains
some simple, but complete, workflows that demonstrate the features
of AwGo and useful workflow idioms.

Typically, you'd call your program's main entry point via Workflow.Run().
This way, the library will rescue any panic, log the stack trace and show an
error message to the user in Alfred.

	// script_filter.go

	package main

	// Import name is "aw"
	import "github.com/deanishe/awgo"

	// aw.Workflow is the main API
	var wf *aw.Workflow

	func init() {
		// Create a new *Workflow using default configuration
		// (workflow settings are read from the environment variables
		// set by Alfred)
		wf = aw.New()
	}

	func main() {
		// Wrap your entry point with Run() to catch and log panics and
		// show an error in Alfred instead of silently dying
		wf.Run(run)
	}

	func run() {
		// Create a new item
		wf.NewItem("Hello World!")
		// And send the results to Alfred
		wf.SendFeedback()
	}


In the Script box (Language = "/bin/bash"):

	./script_filter


Script Filters

To generate results for Alfred to show in a Script Filter, use the feedback
API of Workflow:

	// Create new items
	NewItem()
	NewFileItem()
	NewWarningItem()

	// Sorting/filtering results
	Filter()

	// Send feedback to Alfred
	SendFeedback()

	// Warning/error calls that drop all other Items on the floor
	// and send feedback immediately
	Warn()
	WarnEmpty()
	Fatal()      // exits program
	Fatalf()     // exits program
	FatalError() // exits program

You can set workflow variables (via feedback) with Workflow.Var, Item.Var
and Modifier.Var.

See Workflow.SendFeedback for more documentation.


Run Script actions

Alfred requires a different JSON format if you wish to set workflow variables.

Use the ArgVars (named for its equivalent element in Alfred) struct to
generate output from Run Script actions.

Be sure to set TextErrors to true to prevent Workflow from generating
Alfred JSON if it catches a panic:

	wf.Configure(aw.TextErrors(true))

See ArgVars for more information.


Configuration

New() creates a *Workflow using the default values and workflow settings
read from environment variables set by Alfred.

You can change defaults by passing one or more Options to New(). If
you do not want to use Alfred's environment variables, or they aren't set
(i.e. you're not running the code in Alfred), use NewFromEnv() with a custom
Env implementation.

A Workflow can be re-configured later using its Configure() method.

See the documentation for Option for more information on configuring a Workflow.


Updates

AwGo can check for and install new versions of your workflow.
Subpackage update provides an implementation of the Updater interface and
sources to load updates from GitHub or Gitea releases, or from the URL of
an Alfred `metadata.json` file.

See subpackage update and _examples/update.


Fuzzy filtering

AwGo can filter Script Filter feedback using a Sublime Text-like fuzzy
matching algorithm.

Workflow.Filter() sorts feedback Items against the provided query, removing
those that do not match.

Sorting is performed by subpackage fuzzy via the fuzzy.Sortable interface.

See _examples/fuzzy for a basic demonstration, and _examples/bookmarks for a
demonstration of implementing fuzzy.Sortable on your own structs and customising
the fuzzy sort settings.


Logging

AwGo automatically configures the default log package to write to STDERR
(Alfred's debugger) and a log file in the workflow's cache directory.

The log file is necessary because background processes aren't connected
to Alfred, so their output is only visible in the log. It is rotated when
it exceeds 1 MiB in size. One previous log is kept.

AwGo detects when Alfred's debugger is open (Workflow.Debug() returns true)
and in this case prepends filename:linenumber: to log messages.


Workflow settings

The Config struct (which is included in Workflow as Workflow.Config) provides an
interface to the workflow's settings from the Workflow Environment Variables panel.
https://www.alfredapp.com/help/workflows/advanced/variables/#environment

Alfred exports these settings as environment variables, and you can read them
ad-hoc with the Config.Get*() methods, and save values back to Alfred with
Config.Set().

Using Config.To() and Config.From(), you can "bind" your own structs to the
settings in Alfred:

	// Options will be populated from workflow/environment variables
	type Options struct {
		Server   string `env:"HOSTNAME"`
		Port     int    // use default: PORT
		User     string `env:"USERNAME"`
		Password string `env:"-"` // ignore
	}

	cfg := NewConfig()
	opts := &Options{}

	// Populate Options from the corresponding environment variables.
	if err := cfg.To(opts); err != nil {
		// handle error
	}

	// Save Options back to Alfred.
	if err := cfg.From(opts); err != nil {
		// handle error
	}

See the documentation for Config.To and Config.From for more information,
and _examples/settings for a demo workflow based on the API.


Alfred actions

The Alfred struct provides methods for the rest of Alfred's AppleScript
API. Amongst other things, you can use it to tell Alfred to open, to search
for a query, or to browse/action files & directories.

See documentation of the Alfred struct for more information.


Storing data

AwGo provides a basic, but useful, API for loading and saving data.
In addition to reading/writing bytes and marshalling/unmarshalling to/from
JSON, the API can auto-refresh expired cache data.

See Cache and Session for the API documentation.

Workflow has three caches tied to different directories:

    Workflow.Data     // Cache pointing to workflow's data directory
    Workflow.Cache    // Cache pointing to workflow's cache directory
    Workflow.Session  // Session pointing to cache directory tied to session ID

These all share (almost) the same API. The difference is in when the data go
away.

Data saved with Session are deleted after the user closes Alfred or starts
using a different workflow. The Cache directory is in a system cache
directory, so may be deleted by the system or "system maintenance" tools.

The Data directory lives with Alfred's application data and would not
normally be deleted.


Scripts and background jobs

Subpackage util provides several functions for running script files and
snippets of AppleScript/JavaScript code. See util for documentation and
examples.

AwGo offers a simple API to start/stop background processes via Workflow's
RunInBackground(), IsRunning() and Kill() methods. This is useful for
running checks for updates and other jobs that hit the network or take a
significant amount of time to complete, allowing you to keep your Script
Filters extremely responsive.

See _examples/update and _examples/workflows for demonstrations of this API.


Links

Docs:     https://godoc.org/github.com/deanishe/awgo

Source:   https://github.com/deanishe/awgo

Issues:   https://github.com/deanishe/awgo/issues

Licence:  https://github.com/deanishe/awgo/blob/master/LICENCE
*/
package aw
