// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	aw "github.com/deanishe/awgo"
)

// 4 valid releases, including one prerelease
// v1.0, v2.0, v6.0, v9.0 (Alfred 4 only) and v10.0.1-beta
// v6.0 contains 3 valid workflow files with 3 extensions:
// .alfredworkflow, .alfred3workflow, .alfred4workflow.
// v9.0 contains only a .alfred4workflow file.
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

	var (
		data = mustRead("testdata/github-releases.json")
		dls  []Download
		err  error
	)

	src := &githubSource{
		Repo: "deanishe/alfred-workflow-dummy",
		fetch: func(URL string) ([]byte, error) {
			return ioutil.ReadFile("testdata/empty.json")
		},
	}
	if dls, err = src.Downloads(); err != nil {
		t.Fatal("parse empty JSON")
	}
	if len(dls) != 0 {
		t.Fatal("downloads in empty JSON")
	}

	if dls, err = parseGitHubReleases(data); err != nil {
		t.Fatal("parse GitHub JSON.")
	}

	if len(dls) != len(testGitHubDownloads) {
		t.Fatalf("Wrong download count. Expected=%d, Got=%d", len(testGitHubDownloads), len(dls))
	}

	for i, w := range dls {
		w2 := testGitHubDownloads[i]
		if !reflect.DeepEqual(w, w2) {
			t.Fatalf("Download mismatch at pos %d. Expected=%#v, Got=%#v", i, w2, w)
		}
	}
}

func makeGitHubSource() *githubSource {
	src := &githubSource{Repo: "deanishe/alfred-workflow-dummy"}
	dls, err := parseGitHubReleases(mustRead("testdata/github-releases.json"))
	if err != nil {
		panic(err)
	}
	src.dls = dls
	return src
}

func TestGitHubUpdater(t *testing.T) {
	t.Parallel()
	withTempDir(func(dir string) {
		src := makeGitHubSource()
		dls, err := src.Downloads()
		if err != nil {
			t.Fatal(err)
		}
		if len(dls) != len(testGitHubDownloads) {
			t.Errorf("Wrong no. of downloads. Expected=%v, Got=%v", len(testGitHubDownloads), len(dls))
		}

		// invalid versions
		if _, err := NewUpdater(src, "", dir); err == nil {
			t.Errorf("Accepted empty version")
		}
		if _, err := NewUpdater(src, "stan", dir); err == nil {
			t.Errorf("Accepted invalid version")
		}

		u, err := NewUpdater(src, "0.2.2", dir)
		if err != nil {
			t.Fatalf("create updater: %v", err)
		}

		// Update releases
		if err := u.CheckForUpdate(); err != nil {
			t.Fatalf("Couldn't retrieve releases: %s", err)
		}

		// Check info is cached
		u2, err := NewUpdater(src, "0.2.2", dir)
		if err != nil {
			t.Fatalf("create updater: %v", err)
		}
		if u2.CurrentVersion != u.CurrentVersion {
			t.Errorf("Differing versions. Expected=%v, Got=%v", u.CurrentVersion, u2.CurrentVersion)
		}
		if !u2.LastCheck.Equal(u.LastCheck) {
			t.Errorf("Differing LastCheck. Expected=%v, Got=%v", u.LastCheck, u2.LastCheck)
		}

		testUpdater("github", u, t)
	})
}

func testUpdater(name string, u *Updater, t *testing.T) {
	// v9.0 is latest stable version
	u.CurrentVersion = mustVersion("6")
	u.AlfredVersion = SemVer{}
	if !u.UpdateAvailable() {
		t.Errorf("%s: No update available for defaults", name)
	}

	// v6.0 is the latest stable version for Alfred 3
	u.AlfredVersion = mustVersion("3")
	if u.UpdateAvailable() {
		t.Errorf("%s: Unexpectedly found update", name)
	}
	// Prerelease v10.0-beta is newer
	u.Prereleases = true
	if !u.UpdateAvailable() {
		t.Errorf("%s: No update found", name)
	}

	// v9.0 is the latest stable version for Alfred 4
	u.Prereleases = false
	u.AlfredVersion = mustVersion("4")
	if !u.UpdateAvailable() {
		t.Errorf("%s: No stable update for Alfred 4 found", name)
	}
	u.CurrentVersion = mustVersion("9")
	if u.UpdateAvailable() {
		t.Errorf("%s: Unexpectedly found update for Alfred 4", name)
	}
	// v10.0-beta is the latest pre-release version
	u.Prereleases = true
	if !u.UpdateAvailable() {
		t.Errorf("%s: No pre-release update for Alfred 4 found", name)
	}
}

// TestUnconfiguredUpdater ensures an unconfigured workflow doesn't think it can update
func TestUnconfiguredUpdater(t *testing.T) {
	t.Parallel()

	wf := aw.New()
	if err := wf.ClearCache(); err != nil {
		t.Fatal(fmt.Sprintf("couldn't clear cache: %v", err))
	}
	if wf.UpdateCheckDue() != false {
		t.Fatal("Unconfigured workflow wants to update")
	}
	if wf.UpdateAvailable() != false {
		t.Fatal("Unconfigured workflow wants to update")
	}
	if err := wf.CheckForUpdate(); err == nil {
		t.Fatal("Unconfigured workflow didn't error on update check")
	}
	if err := wf.InstallUpdate(); err == nil {
		t.Fatal("Unconfigured workflow didn't error on update install")
	}

	// Once more with an updater
	wf = aw.New(GitHub("deanishe/alfred-ssh"))
	if wf.UpdateCheckDue() != true {
		t.Fatal("Workflow doesn't want to update")
	}
	if err := wf.ClearCache(); err != nil {
		t.Fatal(err)
	}
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
