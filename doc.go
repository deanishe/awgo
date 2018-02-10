//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

/*

Package aw is a utility library/framework for Alfred 3 workflows
https://www.alfredapp.com/

It provides APIs for interacting with Alfred (e.g. Script Filter feedback) and
the workflow environment (variables, caches, settings).

NOTE: AwGo is currently in development. The API *will* change as I learn to
write idiomatic Go, and should not be considered stable until v1.0.


Links

Docs:     https://godoc.org/github.com/deanishe/awgo

Source:   https://github.com/deanishe/awgo

Issues:   https://github.com/deanishe/awgo/issues

Licence:  https://github.com/deanishe/awgo/blob/master/LICENCE


Features

As of AwGo 0.14, all applicable features of Alfred 3.6 are supported.

The main features are:

	- Simple access to workflow settings.
	- Fluent API for generating Alfred JSON.
	- Fuzzy filtering.
	- Simple, but powerful, API for caching/saving workflow data.
	- Default icons based on macOS system icons.
	- Workflow update API with built-in support for GitHub releases.
	- Pre-configured logging for easier debugging, with a rotated log file.
	- Catches panics, logs stack trace and shows user an error message.
	- "Magic" queries/actions for simplified development and user support.


Usage

Typically, you'd call your program's main entry point via Run(). This way, the
library will rescue any panic, log the stack trace and show an error message to
the user in Alfred.

	# script_filter.go

	package main

	// Import name is "aw"
	import "github.com/deanishe/awgo"

	// Your workflow starts here
	func run() {
		// Add a "Script Filter" result
		aw.NewItem("First result!")
		// Send results to Alfred
		aw.SendFeedback()
	}

	func main() {
		// Wrap your entry point with Run() to catch and log panics and
		// show an error in Alfred instead of silently dying
		aw.Run(run)
	}

In the Script Filter's Script box (Language = "/bin/bash" with "input as argv"):

	./script_filter "$1"


Most package-level functions call the methods of the same name on the default
Workflow struct. If you want to use custom options, you can create a new
Workflow with New(), or reconfigure the default Workflow via the package-level
Configure() function.

Check out the _examples/ subdirectory for some simple, but complete, workflows
which you can copy to get started.

See the documentation for Option for more information on configuring a Workflow.


Fuzzy filtering

AwGo can filter Script Filter feedback using a Sublime Text-like fuzzy
matching algorithm.

Filter() sorts feedback Items against the provided query, removing those that
do not match.

Sorting is performed by subpackage fuzzy via the fuzzy.Sortable interface.

See _examples/fuzzy for a basic demonstration.

See _examples/bookmarks for a demonstration of implementing fuzzy.Sortable on
your own structs and customising the fuzzy sort settings.


Generating feedback

Workflows return data to Alfred via STDOUT. Alfred interprets some data
as JSON and AwGo provides an API for generating this.

JSON feedback for Script Filters is generated mostly via NewItem(), and
then sent to Alfred with SendFeedback().

JSON output to set workflow variables from a Run Script action is
generated with ArgVars.

WARNING: Only send JSON to Alfred once, and don't write anything else to
STDOUT. Otherwise the JSON would be invalid. The Feedback sending methods
ignore subsequent calls, but ArgVars cannot prevent double output.

See SendFeedback for more documentation.


Logging

AwGo uses the default log package. It is automatically configured to log to
STDERR (Alfred's debugger) and to a logfile in the workflow's cache directory.

The log file is rotated when it exceeds 1 MiB in size. One previous log is kept.

AwGo detects when Alfred's debugger is open (Workflow.Debug() returns true) and
in this case prepends filename:linenumber: to log messages.


Saving and caching data

Alfred provides data and cache directories for each workflow. The data directory
is for permanent data and the cache directory for temporary data.  You should
use the CacheDir() and DataDir() methods to get the paths to these directories,
as the methods will ensure that the directories exist.

AwGo's Workflow struct has a simple API for saving data to these directories.
There are basic load/store methods for saving bytes or (un)marshalling structs
to/from JSON, plus LoadOrStore methods that return cached data if they exist and
are new enough, or refresh the cache via a provided function, then return the
data.

Workflow.Data points to the workflow's data directory, Workflow.Cache is
configured to point to the workflow's cache directory, and Workflow.Session also
uses the cache directory, but its cached data expire when the user closes Alfred
or runs a different workflow.

See the Cache and Session structs for the API.


Background jobs

AwGo provides a simple API to start/stop background processes via the
RunInBackground(), IsRunning() and Kill() functions. This is useful for running
checks for updates and other jobs that hit the network or take a significant
amount of time to complete, allowing you to keep your Script Filters extremely
responsive.

See _examples/update for one possible way to use this API.

*/
package aw
