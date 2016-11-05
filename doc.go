/*
Package aw provides utilities for building workflows for Alfred 3.
https://www.alfredapp.com/

Alfred 2 is not supported.

NOTE: This library is currently rather alpha. I'm new to Go, so
doubtless a lot will change as I figure out what I'm doing. The plan
is to implement something in idiomatic Go that is functionally similar
to my Alfred-Workflow library for Python:
http://www.deanishe.net/alfred-workflow/index.html

This library is released under the MIT licence, which you can read
online at https://opensource.org/licenses/MIT

To read this documentation on godoc.org, see
http://godoc.org/gogs.deanishe.net/deanishe/awgo


Features

The current main features are:

	- Easy access to Alfred context, such as data and cache directories.
	- Straightforward generation of Alfred JSON feedback.
	- Support for all applicable Alfred features up to v3.1.
	- Fuzzy sorting/filtering.
	- Catches panics, logs stack trace and shows user an error message.
	- Workflow updates API with built-in support for GitHub releases.
	- (Rotated) Log file for easier debugging.
	- "Magic" arguments/actions for simplified development and user support.
	- OS X system icons.


Upcoming features

These features are planned:

	TODO: Add support for Alfred v3.2 feedback-level variables
	TODO: Add support for Alfred v3.2 re-run feature
	TODO: Starting and managing background processes
	TODO: Caching and storing data
	TODO: Alfred/AppleScript helpers?
	TODO: Implement standard-compliant pre-release comparison in SemVer?


Usage

Typically, you'd call your program's main entry point via Run(). This
way, the library will rescue any panic, log the stack trace and show
an error message to the user in Alfred.

program.go:

	package main

	// Package is called aw
	import "gogs.deanishe.net/deanishe/awgo"
import "9fans.net/go/plan9"

	func run() {
		// Your workflow starts here
		it := aw.NewItem("First result!")
		aw.SendFeedback()
	}

	func main() {
		aw.Run(run)
	}

In the Script Filter's Script box (Language = /bin/bash with input as
argv):

	./program "$1"


The Item struct isn't intended to be used as the workflow's data model,
just as a way to encapsulate search results for Alfred. In particular,
its variables are only settable, not gettable.


Fuzzy sorting/filtering

Sort() and Match() implement fuzzy search, e.g. "of" will match "OmniFocus"
and "got" will match "Game of Thrones".

Match() compares a query and a string, while Sort() sorts an object that
implements the Sortable interface. Both return Result structs for each
compared string.

The Workflow and Feedback structs provide an additional Filter() method,
which fuzzy-sorts Items and removes any that do not match the query.

The Feedback struct implements Sortable, so you can sort/filter feedback
Items. See examples/fuzzy-simple for a basic example.

See examples/fuzzy-cached for a demonstration of implementing Sortable
on your own structs and customising the sort settings.

The algorithm is based on Forrest Smith's reverse engineering of Sublime
Text's search: https://blog.forrestthewoods.com/reverse-engineering-sublime-text-s-fuzzy-match-4cffeed33fdb

It additionally strips diacritics from sort keys if the query is ASCII.


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

If you want to include a warning with other results, use NewWarningItem().


Logging

AwGo uses the default log package. It is automatically configured to log
to STDERR (Alfred's debugger) and to a logfile in the workflow's cache
directory.

The log file is rotated when it exceeds 1 MiB in size. One previous
log is kept.

AwGo detects when Alfred's debugger is open (Workflow.Debug() returns
true) and in this case prepends filename:linenumber: to log messages.


Updates

The Updater/Releaser API provides the ability to check for newer versions
of your workflow. A GitHub Releaser that updates from GitHub releases is built in.
You can use your own backend by implementing the Releaser interface.

The only hard requirement is support for (mostly) semantic version numbers.
See http://semver.org for details.


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
