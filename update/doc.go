// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

/*
Package update implements an API for fetching workflow updates from remote servers.

It is the "backend" for aw.Workflow's update API, and provides concrete updaters for
GitHub and Gitea releases, and Alfred metadata.json files (as aw.Options). Updater
implements aw.Updater and you can create a custom Updater to use with
aw.Workflow/aw.Update() by passing a custom implementation of Source to NewUpdater().

The only hard requirement is support for (mostly) semantic version numbers. See
SemVer documentation and http://semver.org for details.

Updater is also Alfred-version-aware, and ignores incompatible workflow version,
e.g. workflow files with the extension ".alfred4workflow" are ignored
when Updater is run in Alfred 3.

See ../_examples/update for one possible way to using the updater API.
*/
package update
