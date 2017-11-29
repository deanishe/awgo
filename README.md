
<div align="center">
    <img src="./Icon.png" alt="AwGo Logo" title="AwGo Logo">
</div>

AwGo — A Go library for Alfred workflows
========================================

Full-featured library to build lightning-fast workflows in a jiffy.

Features
--------

- Easy access to [Alfred context][context], such as data and cache directories.
- Fluent API for generating [Alfred JSON feedback][feedback] for Script Filters.
- Support for all applicable Alfred features up to v3.5.
- [Fuzzy sorting/filtering][fuzzy] of results.
- [Simple, but powerful, API][cache-api] for [caching/saving workflow data][cache].
- [Catches panics, logs stack trace and shows user an error message][run].
- Workflow [updates API][update] with built-in support for [GitHub releases][update-github].
- [Pre-configured logging][logging] for easier debugging, with a rotated log file.
- ["Magic" queries/actions][magic] for simplified development and user support.
- macOS [system icons][icons].


Installation & usage
--------------------

Install AwGo with:

```sh
go get -u github.com/deanishe/awgo
```

Typically, you'd call your program's main entry point via `Run()`. This
way, the library will rescue any panic, log the stack trace and show
an error message to the user in Alfred.

program.go:

```go
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
```

In the Script Filter's Script box (Language = /bin/bash with input as
argv):

```sh
./program "$1"
```


Documentation
-------------

Read the docs [on GoDoc][godoc].

Check out the [example workflows][examples-code] ([docs][examples-docs]), which show how to use AwGo. Use one as a template to get your own workflow up and running quickly.


Running/testing
---------------

The library, and therefore the unit tests, rely on being run in an Alfred-like environment, as they pull configuration options from environment variables (which are set by Alfred).

As such, you must `source` the `env.sh` script in the project root or run unit tests via the `run-tests.sh` script (which sources `env.sh` then calls `go test`).


Licensing & thanks
------------------

This library is released under the [MIT licence][licence].

The icon is based on the [Go Gopher][gopher] by [Renee French][renee].


[alfred]: https://www.alfredapp.com/
[licence]: ./LICENCE
[godoc]: https://godoc.org/github.com/deanishe/awgo
[gopher]: https://blog.golang.org/gopher
[renee]: http://reneefrench.blogspot.com
[context]: https://godoc.org/github.com/deanishe/awgo#Context
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
[examples-code]: https://github.com/deanishe/awgo/tree/master/examples
[examples-docs]: https://godoc.org/github.com/deanishe/awgo/examples
