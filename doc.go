/*
Package workflow provides utilities for building workflows for Alfred 2.
https://www.alfredapp.com/

You need Alfred's Powerpack to be able to use workflows.

NOTE: This software is very alpha and not even half-finished.

To read this documentation on godoc.org, see
http://godoc.org/gogs.deanishe.net/deanishe/awgo.git

This library provides an API for communicating with Alfred and several
convenience methods for common workflow tasks.

It's primary purpose is to make writing Script Filters easier.

USAGE

program.go:

	package main

	import "gogs.deanishe.net/deanishe/awgo"

	func run() {
		// Your workflow starts here
	}

	func main() {
		workflow.Run(run)
	}

In the Script Filter's Script box (Language = /bin/bash):

	./program "{query}"

*/
package workflow
