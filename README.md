
<div align="center">
    <img src="https://raw.githubusercontent.com/deanishe/awgo/master/Icon.png" alt="AwGo Logo" title="AwGo Logo">
</div>


AwGo — A Go library for Alfred workflows
========================================

[![Build Status][ghaction-badge]][ghaction-link]
[![Go Report Card][goreport-badge]][goreport-link]
[![Coverage Status][coveralls-badge]][coveralls-link]
[![GoDoc][godoc-badge]][godoc-link]

<!--
[![Build Status][azure-badge]][azure-link]
[![Build Status][travis-badge]][travis-link]
[![Codacy Quality][codacy-badge-quality]][codacy-link]
[![Codacy coverage][codacy-badge-coverage]][codacy-link]
[![Codecov coverage][codecov-badge]][codecov-link]
-->


Full-featured library to build lightning-fast workflows in a jiffy.

- [Features](#features)
- [Installation & usage](#installation--usage)
- [Documentation](#documentation)
- [Requirements](#requirements)
- [Development](#development)
- [Licensing & thanks](#licensing--thanks)


Features
--------

- Full support for Alfred 3 & 4
- Bi-directional interface to [workflow's config][config]
- Fluent API for generating [Script Filter JSON][feedback]
- [Fuzzy sorting/filtering][fuzzy]
- [Simple, powerful API][cache-api] for [caching/saving workflow data][cache]
- Keychain API to [securely store (and sync) sensitive data][keychain]
- Helpers to [easily run scripts and script code][scripts]
- Workflow [update API][update] with built-in support for [GitHub][update-github] & [Gitea][update-gitea]
- [Pre-configured logging][logging] for easier debugging, with a rotated log file
- [Catches panics, logs stack trace and shows user an error message][run]
- ["Magic" queries/actions][magic] for simplified development and user support
- macOS [system icons][icons]


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
- `alfred_version` (not needed for Alfred 4+)


Development
-----------

To create a sufficiently Alfred-like environment, you can `source` the `env.sh`
script in the project root or run unit tests via the `run-tests.sh` script
(which also sets up an appropriate environment before calling `go test`).


Licensing & thanks
------------------

This library is released under the [MIT licence][licence]. It was built with
[neovim][neovim] and [GoLand by JetBrains][jetbrains].

The icon is based on the [Go Gopher][gopher] by [Renee French][renee].


[alfred]: https://www.alfredapp.com/
[licence]: ./LICENCE
[godoc]: https://godoc.org/github.com/deanishe/awgo
[gopher]: https://blog.golang.org/gopher
[renee]: http://reneefrench.blogspot.com
[config]: https://godoc.org/github.com/deanishe/awgo#Config
[feedback]: https://godoc.org/github.com/deanishe/awgo#Feedback.NewItem
[fuzzy]: https://godoc.org/github.com/deanishe/awgo/fuzzy
[cache]: https://godoc.org/github.com/deanishe/awgo#hdr-Storing_data
[cache-api]: https://godoc.org/github.com/deanishe/awgo#Cache
[run]: https://godoc.org/github.com/deanishe/awgo#Run
[keychain]: https://godoc.org/github.com/deanishe/awgo/keychain
[scripts]: https://godoc.org/github.com/deanishe/awgo/util#hdr-Scripting
[update]: https://godoc.org/github.com/deanishe/awgo/update
[update-github]: https://godoc.org/github.com/deanishe/awgo/update#GitHub
[update-gitea]: https://godoc.org/github.com/deanishe/awgo/update#Gitea
[logging]: https://godoc.org/github.com/deanishe/awgo#hdr-Logging
[magic]: https://godoc.org/github.com/deanishe/awgo#MagicAction
[icons]: https://godoc.org/github.com/deanishe/awgo#Icon
[examples-code]: https://github.com/deanishe/awgo/tree/master/_examples
[examples-docs]: https://godoc.org/github.com/deanishe/awgo/_examples
[jetbrains]: https://www.jetbrains.com/?from=deanishe/awgo
[neovim]: https://neovim.io/

[godoc-badge]: https://godoc.org/github.com/deanishe/awgo?status.svg
[godoc-link]: https://godoc.org/github.com/deanishe/awgo
[goreport-link]: https://goreportcard.com/report/github.com/deanishe/awgo
[goreport-badge]: https://goreportcard.com/badge/github.com/deanishe/awgo
[azure-badge]: https://img.shields.io/azure-devops/build/deanishe/6cd8e4fe-7366-4485-aea6-e9d75e7757b2/1
[azure-link]: https://dev.azure.com/deanishe/AwGo/_build
[ghaction-badge]: https://github.com/deanishe/awgo/workflows/Test/badge.svg
[ghaction-link]: https://github.com/deanishe/awgo/actions?query=workflow%3ATest
[coveralls-badge]: https://coveralls.io/repos/github/deanishe/awgo/badge.svg?branch=master&v3
[coveralls-link]: https://coveralls.io/github/deanishe/awgo?branch=master

<!--
[coverage-badge]: https://img.shields.io/codacy/coverage/e785f7b0e830468da6fa2856d62e59ab?color=brightgreen
[codacy-link]: https://www.codacy.com/app/deanishe/awgo
[travis-badge]: https://img.shields.io/travis/deanishe/awgo
[travis-link]: https://travis-ci.org/deanishe/awgo
[codacy-badge-quality]: https://api.codacy.com/project/badge/Grade/e785f7b0e830468da6fa2856d62e59ab
[codacy-badge-coverage]: https://api.codacy.com/project/badge/Coverage/e785f7b0e830468da6fa2856d62e59ab
[travis-badge]: https://travis-ci.org/deanishe/awgo.svg?branch=master
[codecov-badge]: https://codecov.io/gh/deanishe/awgo/branch/master/graph/badge.svg
[codecov-link]: https://codecov.io/gh/deanishe/awgo
[coveralls-badge]: https://img.shields.io/coveralls/github/deanishe/awgo/master
-->
