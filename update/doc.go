// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

/*
Package update implements an API for fetching updates to workflows from remote servers.

The Updater/Releaser API provides the ability to check for newer versions
of your workflow. Support for updating from GitHub releases is built in.
See GitHub example.

You can use your own backend by implementing the Releaser interface.

The only hard requirement is support for (mostly) semantic version numbers.
See SemVer documentation and http://semver.org for details.

This package is the "backend". You should set an Updater on an aw.Workflow
struct (using e.g. the GitHub aw.Option) and use the Workflow methods
CheckForUpdate(), UpdateAvailable() and InstallUpdate() to interact with
the Updater.

See ../_examples/update for one possible way to use this API.
*/
package update
