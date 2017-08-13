
AwGo — A Go library for Alfred 3 workflows
==========================================

https://godoc.org/git.deanishe.net/deanishe/awgo

**Note**: This library is still in alpha. The API may change at any time.


Running/testing
---------------

The library, and therefore the unit tests, rely on being run in an Alfred-like environment, as they pull configuration options from environment variables (which are set by Alfred).

As such, you must `source` the `env.sh` script in the project root or run unit tests via the `run-tests.sh` script (which sources `env.sh` then calls `go test`).

