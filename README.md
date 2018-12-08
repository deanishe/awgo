
<div align="center">
    <img src="./Icon.png" alt="AwGo Logo" title="AwGo Logo">
</div>

AwGo — A Go library for Alfred workflows
========================================

[![Build Status][travis-icon]][travis-link]
[![Go Report Card][goreport-icon]][goreport-link]
[![Codacy Badge][codacy-icon]][codacy-link]
[![Coverage Status][coveralls-icon]][coveralls-link]
[![GoDoc][godoc-icon]][godoc-link]

Full-featured library to build lightning-fast workflows in a jiffy.

Features
--------

- Easy access to [Alfred and workflow settings][config].
- Fluent API for generating [Alfred JSON feedback][feedback] for Script Filters.
- Support for all applicable Alfred features up to v3.6.
- [Fuzzy sorting/filtering][fuzzy] of results.
- [Simple, but powerful, API][cache-api] for [caching/saving workflow data][cache].
- [Catches panics, logs stack trace and shows user an error message][run].
- Workflow [update API][update] with built-in support for [GitHub releases][update-github].
- [Pre-configured logging][logging] for easier debugging, with a rotated log file.
- ["Magic" queries/actions][magic] for simplified development and user support.
- macOS [system icons][icons].


Installation & usage
--------------------

Install AwGo with:

```sh
go get -u github.com/deanishe/awgo
```

Typically, you'd call your program's main entry point via `Workflow.Run()`.
This way, the library will rescue any panic, log the stack trace and show
an error message to the user in Alfred.

program.go:

```go
package main

// Package is called aw
import "github.com/deanishe/awgo"

// Workflow is the main API
var wf *aw.Workflow

func init() {
    // Create a new Workflow using default settings.
    // Critical settings are provided by Alfred via environment variables,
    // so this *will* die in flames if not run in an Alfred-like environment.
    wf = aw.New()
}

// Your workflow starts here
func run() {
    // Add a "Script Filter" result
    wf.NewItem("First result!")
    // Send results to Alfred
    wf.SendFeedback()
}

func main() {
    // Wrap your entry point with Run() to catch and log panics and
    // show an error in Alfred instead of silently dying
    wf.Run(run)
}
```

In the Script Filter's Script box (Language = /bin/bash with input as
argv):

```sh
./program "$1"
```

Documentation
-------------

Read the docs [on GoDoc][godoc].

Check out the [example workflows][examples-code] ([docs][examples-docs]), which
show how to use AwGo. Use one as a template to get your own workflow up and
running quickly.


Requirements
------------

The library (and therefore the unit tests) rely on being run in a minimally
Alfred-like environment, as they pull configuration options from the environment
variables set by Alfred.

This means that if you want to run AwGo-based code outside Alfred, e.g. in your
shell, you must set at least the following environment variables to meaningful
values, or the library will panic:

- `alfred_workflow_bundleid`
- `alfred_workflow_cache`
- `alfred_workflow_data`

And if you're using the update API, also:

- `alfred_workflow_version`


Development
-----------

To create a sufficiently Alfred-like environment, you can `source` the `env.sh`
script in the project root or run unit tests via the `run-tests.sh` script
(which also sets up an appropriate environment before calling `go test`).


Licensing & thanks
------------------

This library is released under the [MIT licence][licence].

The icon is based on the [Go Gopher][gopher] by [Renee French][renee].


[alfred]: https://www.alfredapp.com/
[licence]: ./LICENCE
[godoc]: https://godoc.org/github.com/deanishe/awgo
[gopher]: https://blog.golang.org/gopher
[renee]: http://reneefrench.blogspot.com
[config]: https://godoc.org/github.com/deanishe/awgo#Config
[feedback]: https://godoc.org/github.com/deanishe/awgo#Feedback.NewItem
[fuzzy]: https://godoc.org/github.com/deanishe/awgo/fuzzy
[cache]: https://godoc.org/github.com/deanishe/awgo#hdr-Saving_and_caching_data
[cache-api]: https://godoc.org/github.com/deanishe/awgo#Cache
[run]: https://godoc.org/github.com/deanishe/awgo#Run
[update]: https://godoc.org/github.com/deanishe/awgo/update
[update-github]: https://godoc.org/github.com/deanishe/awgo/update#GitHub
[logging]: https://godoc.org/github.com/deanishe/awgo#hdr-Logging
[magic]: https://godoc.org/github.com/deanishe/awgo#MagicAction
[icons]: https://godoc.org/github.com/deanishe/awgo#Icon
[examples-code]: https://github.com/deanishe/awgo/tree/master/_examples
[examples-docs]: https://godoc.org/github.com/deanishe/awgo/_examples
[travis-link]: https://travis-ci.org/deanishe/awgo
[travis-icon]: https://travis-ci.org/deanishe/awgo.svg?branch=master
[goreport-link]: https://goreportcard.com/report/github.com/deanishe/awgo
[goreport-icon]: https://goreportcard.com/badge/github.com/deanishe/awgo
[codacy-icon]: https://api.codacy.com/project/badge/Grade/e785f7b0e830468da6fa2856d62e59ab
[codacy-link]: https://www.codacy.com/app/deanishe/awgo?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=deanishe/awgo&amp;utm_campaign=Badge_Grade
[coveralls-icon]: https://coveralls.io/repos/github/deanishe/awgo/badge.svg?branch=master
[coveralls-link]: https://coveralls.io/github/deanishe/awgo?branch=master
[godoc-icon]: https://godoc.org/github.com/deanishe/awgo?status.svg
[godoc-link]: https://godoc.org/github.com/deanishe/awgo
