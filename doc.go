/*
Package aw provides utilities for building workflows for Alfred 3.
https://www.alfredapp.com/

NOTE: This library is currently rather alpha. I'm new to Go, so
doubtless a lot will change as I figure out what I'm doing.

This library is released under the MIT licence, which you can read
online at https://opensource.org/licenses/MIT

To read this documentation on godoc.org, see
http://godoc.org/gogs.deanishe.net/deanishe/awgo


Features

The current main features are:

	- Easy access to Alfred context, such as data and cache directories.
	- Simple generation of Alfred JSON feedback.
	- Fuzzy sorting.
	- Catches panics, logs stack trace and shows user an error message.
	- (Rotated) Log file for easier debugging.
	- OS X system icons.

	TODO: Starting background processes
	TODO: Caching and storing data
	TODO: Workflow update via GitHub releases
	TODO: Magic arguments
	TODO: Alfred/AppleScript helpers


Usage

Typically, you'd call your program's main entry point via Run(). This
way, the library will rescue any panic, log the stack trace and show
an error to the user.

program.go:

	package main

	// Package is called aw
	import "gogs.deanishe.net/deanishe/awgo"
import "github.com/mvdan/interfacer"

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


Fuzzy sorting

Sort() and Match() implement Alfred-like fuzzy search, e.g. "of" will
match "OmniFocus" and "got" will match "Game of Thrones".

Match() compares a query and a string, while Sort() sorts an object that
implements the Sortable interface. Both return Result structs for each
compared string.

The algorithm is based on Forrest Smith's reverse engineering of Sublime
Text's search: https://blog.forrestthewoods.com/reverse-engineering-sublime-text-s-fuzzy-match-4cffeed33fdb

The Feedback struct implements Sortable, so you can sort/filter feedback
Items. See examples/fuzzy-simple for a basic example.

See examples/fuzzy-cached for a demonstration of implementing Sortable
on your own structs and customising the sort settings.


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

Awgo uses the default log package. It is automatically configured to log
to STDERR (Alfred's debugger) and to a logfile in the workflow's cache
directory.

The log file is rotated when it exceeds 1 MiB in size. One previous
log is kept.

Awgo detects when Alfred's debugger is open (Workflow.Debug() returns
true) and in this case prepends filename:linenumber: to log messages.


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
