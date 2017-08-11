/*
Package update implements an API for fetching updates to workflows from remote servers.

The Updater/Releaser API provides the ability to check for newer versions
of your workflow. A GitHub Releaser that updates from GitHub releases is built in.
You can use your own backend by implementing the Releaser interface.

The only hard requirement is support for (mostly) semantic version numbers.
See http://semver.org for details.

See ../examples/update for one possible way to use this API.
*/
package update
