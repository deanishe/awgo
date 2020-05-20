// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"io/ioutil"
	"testing"

	aw "github.com/deanishe/awgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 6 valid releases, including one prerelease
// v1.0, v2.0, v6.0, v7.1.0-beta, v9.0 (Alfred 4+ only), v10.0-beta
var testGitHubDownloads = []Download{
	// Latest version for Alfred 4
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v10.0-beta/Dummy-10.0-beta.alfredworkflow",
		Filename:   "Dummy-10.0-beta.alfredworkflow",
		Version:    mustVersion("v10.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 4
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v9.0/Dummy-9.0.alfred4workflow",
		Filename:   "Dummy-9.0.alfred4workflow",
		Version:    mustVersion("v9.0"),
		Prerelease: false,
	},
	// Latest version for Alfred 3
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v7.1.0-beta/Dummy-7.1-beta.alfredworkflow",
		Filename:   "Dummy-7.1-beta.alfredworkflow",
		Version:    mustVersion("v7.1.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 3
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred4workflow",
		Filename:   "Dummy-6.0.alfred4workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred3workflow",
		Filename:   "Dummy-6.0.alfred3workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfredworkflow",
		Filename:   "Dummy-6.0.alfredworkflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v2.0/Dummy-2.0.alfredworkflow",
		Filename:   "Dummy-2.0.alfredworkflow",
		Version:    mustVersion("v2.0"),
		Prerelease: false,
	},
	{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v1.0/Dummy-1.0.alfredworkflow",
		Filename:   "Dummy-1.0.alfredworkflow",
		Version:    mustVersion("v1.0"),
		Prerelease: false,
	},
}

func TestParseGitHub(t *testing.T) {
	t.Parallel()
	testParseReleases("GitHub", "testdata/github-releases.json", testGitHubDownloads, t)
}

func testParseReleases(name, jsonPath string, downloads []Download, t *testing.T) {
	t.Run(name+"parse empty releases", func(t *testing.T) {
		t.Parallel()
		src := &source{
			fetch: func(URL string) ([]byte, error) {
				return ioutil.ReadFile("testdata/empty.json")
			},
		}
		dls, err := src.Downloads()
		require.Nil(t, err, "parse empty JSON")
		require.Equal(t, 0, len(dls), "downloads in empty JSON")
	})

	t.Run(name+" parse releases", func(t *testing.T) {
		t.Parallel()
		src := &source{
			fetch: func(URL string) ([]byte, error) {
				return ioutil.ReadFile(jsonPath)
			},
		}
		dls, err := src.Downloads()
		require.Nil(t, err, "parse %s JSON", name)
		require.Equal(t, len(downloads), len(dls), "wrong %s download count", name)
		require.Equal(t, downloads, dls, "%s downloads not equal", name)
	})
}

func TestGitHubUpdater(t *testing.T) {
	t.Parallel()
	src := &source{
		URL: "https://api.github.com/repos/deanishe/alfred-workflow-dummy",
		fetch: func(URL string) ([]byte, error) {
			return ioutil.ReadFile("testdata/github-releases.json")
		},
	}
	testSourceUpdater("GitHub", src, t)
}

func testSourceUpdater(name string, src *source, t *testing.T) {
	withTempDir(func(dir string) {
		dls, err := src.Downloads()
		require.Nil(t, err, "parse %s JSON", name)
		assert.Equal(t, len(testGitHubDownloads), len(dls), "wrong no. of %s downloads", name)

		t.Run(name+" invalid versions", func(t *testing.T) {
			for _, s := range []string{"", "stan"} {
				_, err := NewUpdater(src, s, dir)
				assert.NotNil(t, err, "accepted invalid version %q", s)
			}
		})

		var u *Updater
		t.Run(name+" updater", func(t *testing.T) {
			u, err = NewUpdater(src, "0.2.2", dir)
			require.Nil(t, err, "create updater")
			require.Nil(t, u.CheckForUpdate(), "retrieve releases")
		})

		t.Run(name+" updater info cached", func(t *testing.T) {
			// Check info is cached
			u2, err := NewUpdater(src, "0.2.2", dir)
			require.Nil(t, err, "create updater")

			assert.Equal(t, u2.CurrentVersion, u.CurrentVersion, "differing versions")
			assert.True(t, u2.LastCheck.Equal(u.LastCheck), "differing LastCheck")
		})

		t.Run(name+" updater", func(t *testing.T) {
			testUpdater(name, u, t)
		})
	})
}

func testUpdater(name string, u *Updater, t *testing.T) {
	u.CurrentVersion = mustVersion("6")

	tests := []struct {
		name           string
		currentVersion string
		alfredVersion  string
		prereleases    bool
		x              bool
	}{
		// v9.0 is latest stable version
		{"sanity check", "6", "", false, true},
		// v6.0 is the latest stable version for Alfred 3
		{"update for Alfred 3 not available", "6", "3", false, false},
		// Prerelease v10.0-beta is newer
		{"pre-release for Alfred 3 available", "6", "3", true, true},
		// v9.0 is the latest stable version for Alfred 4
		{"stable update for Alfred 4 available", "6", "4", false, true},
		{"no update for Alfred 4 available", "9", "4", false, false},
		// v10.0-beta is the latest pre-release version
		{"pre-release update for Alfred 4 available", "9", "4", true, true},
	}

	for _, td := range tests {
		t.Run(name+" "+td.name, func(t *testing.T) {
			u.CurrentVersion = mustVersion(td.currentVersion)
			u.AlfredVersion = mustVersion(td.alfredVersion)
			u.Prereleases = td.prereleases
			assert.Equal(t, td.x, u.UpdateAvailable(), td.name+" failed")
		})
	}
}

// TestUnconfiguredUpdater ensures an unconfigured workflow doesn't think it can update
func TestUnconfiguredUpdater(t *testing.T) {
	t.Parallel()

	wf := aw.New()
	assert.Nil(t, wf.ClearCache(), "failed to clear cache")
	assert.False(t, wf.UpdateCheckDue(), "unconfigured workflow wants to update")
	assert.False(t, wf.UpdateAvailable(), "unconfigured workflow has available update")
	assert.NotNil(t, wf.CheckForUpdate(), "unconfigured workflow didn't error on update")
	assert.NotNil(t, wf.InstallUpdate(), "unconfigured workflow didn't error on update install")

	// Once more with an updater
	wf = aw.New(GitHub("deanishe/alfred-ssh"))
	assert.True(t, wf.UpdateCheckDue(), "workflow doesn't want to update")
	assert.Nil(t, wf.ClearCache(), "failed to clear cache")
}

// Configure Workflow to update from a GitHub repo.
func ExampleGitHub() {
	// Set source repo using GitHub Option
	wf := aw.New(GitHub("deanishe/alfred-ssh"))
	// Is a check for a newer version due?
	fmt.Println(wf.UpdateCheckDue())
	// Output:
	// true
}
