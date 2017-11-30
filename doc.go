/*
Package aw is a utility library/framework for building workflows for Alfred.
https://www.alfredapp.com/

It provides APIs for interacting with Alfred (e.g. generating Script Filter
feedback and setting workflow variables) with a host of convenience functions,
plus support for common workflow idioms, such as caching data from
applications/web services and updating the cache in a background process to
keep your workflow super-responsive.

NOTE: AwGo is currently in development. The API *will* change as I learn to
write idiomatic Go, and should not be considered stable until v1.0.

This library is released under the MIT licence, which you can read
online at https://opensource.org/licenses/MIT

Read this documentation online at http://godoc.org/github.com/deanishe/awgo


Features

The main features are:

	- Easy access to Alfred context, such as data and cache directories.
	- Fluent API for generating Alfred JSON feedback for Script Filters.
	- Support for all applicable Alfred features up to v3.5.
	- Fuzzy sorting/filtering.
	- Simple, but powerful, API for caching/saving workflow data.
	- Catches panics, logs stack trace and shows user an error message.
	- Workflow updates API with built-in support for GitHub releases.
	- Pre-configured logging for easier debugging, with a rotated log file.
	- "Magic" queries/actions for simplified development and user support.
	- macOS system icons.


Usage

Typically, you'd call your program's main entry point via Run(). This
way, the library will rescue any panic, log the stack trace and show
an error message to the user in Alfred.

program.go:

	package main

	// Package is called aw
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

In the Script Filter's Script box (Language = /bin/bash with input as
argv):

	./program "$1"


The Item struct isn't intended to be used as the workflow's data model,
just as a way to encapsulate search results for Alfred. In particular,
its variables are only settable, not gettable. However, the Feedback struct,
to which Items belong, supports fuzzy.Interface, so in most situations, it's
not necessary to implement fuzzy.Interface yourself in order to use fuzzy
filtering.

Most package-level functions call the methods of the same name on the default
Workflow struct. If you want to use custom options, you can create a new
Workflow with New(), or reconfigure the default Workflow via the package-level
Configure() function.

Check out the examples/ subdirectory for some simple, but complete,
workflows which you can copy to get started.


Fuzzy filtering

Subpackage fuzzy provides a fuzzy search algorithm modelled on Sublime
Text's search. Implement fuzzy.Interface to make a slice fuzzy-sortable.

The Feedback struct implements this interface.

Feedback and Workflow provide an additional Filter() method,
which fuzzy-sorts the contained Items and removes any that do not match
the query.

See examples/fuzzy for a basic demonstration.

See examples/bookmarks for a demonstration of implementing fuzzy.Interface
on your own structs and customising the fuzzy sort settings.


Sending results to Alfred

Generally, you'll want to use NewItem() to create items, then
SendFeedback() to generate the JSON and send it to Alfred (i.e. print
it to STDOUT).

You can only call a sending method once: multiple calls would result in
invalid JSON, as there'd be multiple root objects, so any subsequent
calls to sending methods are logged and ignored. Sending methods are:

	SendFeedback()
	Fatal()
	Fatalf()
	FatalError()
	Warn()
	WarnEmpty()  // only sends if there are no items

The Workflow struct (more precisely, its Feedback struct) retains the
Item, so you don't need to. Just populate it and then call
SendFeedback() when all your results are ready.

There are additional helper methods for specific situations.

NewFileItem() returns an Item pre-populated from a filepath (title,
subtitle, icon, arg, etc.).

FatalError(), Fatal() and Fatalf() will immediately send a single result
to Alfred with an error message and then call log.Fatalf(), terminating
the workflow.

Warn() also immediately sends a single result to Alfred with a warning
message (and icon), but does not terminate the workflow. However,
because the JSON has already been sent to Alfred, you can't send any
more results after calling Warn().

WarnEmpty() calls Warn() if there are no (other) Items to send to Alfred.

If you want to include a warning with other results, use NewWarningItem().


Logging

AwGo uses the default log package. It is automatically configured to log
to STDERR (Alfred's debugger) and to a logfile in the workflow's cache
directory.

The log file is rotated when it exceeds 1 MiB in size. One previous
log is kept.

AwGo detects when Alfred's debugger is open (Workflow.Debug() returns
true) and in this case prepends filename:linenumber: to log messages.


Saving and caching data

Alfred provides data and cache directories for each workflow. The data
directory is for permanent data and the cache directory for temporary data.
You should use the CacheDir() and DataDir() methods to get the paths to
these directories, as the methods will ensure that the directories exist.

AwGo's Workflow struct has a simple API for saving data to these
directories. There are basic load/store methods for saving bytes or
(un)marshalling structs to/from JSON, plus LoadOrStore methods that return
cached data if they exist and are new enough, or refresh the cache via a
provided function, then return the data.

Workflow.Data points to the workflow's data directory, Workflow.Cache is
configured to point to the workflow's cache directory, and Workflow.Session
also uses the cache directory, but its cached data expire when the user
closes Alfred or runs a different workflow.

See the Cache and Session structs for the API.


Background jobs

AwGo provides a simple API to start/stop background processes via the
RunInBackground(), IsRunning() and Kill() functions. This is useful
for running checks for updates and other jobs that hit the network or
take a significant amount of time to complete, allowing you to keep
your Script Filters extremely responsive.

See examples/update for one possible way to use this API.


Performance

For smooth performance in Alfred, a Script Filter should ideally finish
in under 0.1 seconds. 0.3 seconds is about the upper limit for your
workflow not to feel sluggish.

As a rough guideline, loading and sorting/filtering ~20K is about the
limit before performance becomes noticeably hesitant.

If you have a larger dataset, consider using something like sqlite—which
can easily handle hundreds of thousands of items—for your
datastore.

*/
package aw
