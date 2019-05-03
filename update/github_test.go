// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"reflect"
	"testing"

	aw "github.com/deanishe/awgo"
)

var testGitHubDownloads = []Download{
	// Latest version for Alfred 4
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v10.0-beta/Dummy-10.0-beta.alfredworkflow",
		Filename:   "Dummy-10.0-beta.alfredworkflow",
		Version:    mustVersion("v10.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 4
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v9.0/Dummy-9.0.alfred4workflow",
		Filename:   "Dummy-9.0.alfred4workflow",
		Version:    mustVersion("v9.0"),
		Prerelease: false,
	},
	// Latest version for Alfred 3
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v7.1.0-beta/Dummy-7.1-beta.alfredworkflow",
		Filename:   "Dummy-7.1-beta.alfredworkflow",
		Version:    mustVersion("v7.1.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 3
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred4workflow",
		Filename:   "Dummy-6.0.alfred4workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred3workflow",
		Filename:   "Dummy-6.0.alfred3workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfredworkflow",
		Filename:   "Dummy-6.0.alfredworkflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v2.0/Dummy-2.0.alfredworkflow",
		Filename:   "Dummy-2.0.alfredworkflow",
		Version:    mustVersion("v2.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v1.0/Dummy-1.0.alfredworkflow",
		Filename:   "Dummy-1.0.alfredworkflow",
		Version:    mustVersion("v1.0"),
		Prerelease: false,
	},
}

func TestParseGitHub(t *testing.T) {
	t.Parallel()

	src := &githubSource{
		Repo: "deanishe/alfred-workflow-dummy",
		fetch: func(URL string) ([]byte, error) {
			return []byte(ghReleasesEmptyJSON), nil
		},
	}
	dls, err := src.Downloads()
	if err != nil {
		t.Fatal("parse empty JSON")
	}
	if len(dls) != 0 {
		t.Fatal("downloads in empty JSON")
	}

	dls, err = parseGitHubReleases([]byte(ghReleasesJSON))
	if err != nil {
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
	dls, err := parseGitHubReleases([]byte(ghReleasesJSON))
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

var (
	ghReleasesEmptyJSON = `[]`
	// 4 valid releases, including one prerelease
	// v1.0, v2.0, v6.0, v9.0 (Alfred 4 only) and v10.0.1-beta
	// v6.0 contains 3 valid workflow files with 3 extensions:
	// .alfredworkflow, .alfred3workflow, .alfred4workflow.
	// v9.0 contains only a .alfred4workflow file.
	ghReleasesJSON = `[
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132595",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132595/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132595/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v10.0-beta",
    "id": 17132595,
    "node_id": "MDc6UmVsZWFzZTE3MTMyNTk1",
    "tag_name": "v10.0-beta",
    "target_commitish": "master",
    "name": "Latest release (pre-release)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": true,
    "created_at": "2019-05-03T12:27:30Z",
    "published_at": "2019-05-03T12:28:36Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/12368368",
        "id": 12368368,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDEyMzY4MzY4",
        "name": "Dummy-10.0-beta.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2019-05-03T12:28:19Z",
        "updated_at": "2019-05-03T12:28:20Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v10.0-beta/Dummy-10.0-beta.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v10.0-beta",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v10.0-beta",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132521",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132521/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/17132521/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v9.0",
    "id": 17132521,
    "node_id": "MDc6UmVsZWFzZTE3MTMyNTIx",
    "tag_name": "v9.0",
    "target_commitish": "master",
    "name": "Latest release (Alfred 4)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2019-05-03T12:24:12Z",
    "published_at": "2019-05-03T12:25:11Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/12368327",
        "id": 12368327,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDEyMzY4MzI3",
        "name": "Dummy-9.0.alfred4workflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2019-05-03T12:25:01Z",
        "updated_at": "2019-05-03T12:25:02Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v9.0/Dummy-9.0.alfred4workflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v9.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v9.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v8point0",
    "id": 14412055,
    "node_id": "MDc6UmVsZWFzZTE0NDEyMDU1",
    "tag_name": "v8point0",
    "target_commitish": "master",
    "name": "Invalid tag (non-semantic)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2018-12-07T16:03:23Z",
    "published_at": "2018-12-07T16:04:30Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/10048629",
        "id": 10048629,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDEwMDQ4NjI5",
        "name": "Dummy-eight.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 2,
        "created_at": "2018-12-07T16:04:24Z",
        "updated_at": "2018-12-07T16:04:25Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v8point0/Dummy-eight.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v8point0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v8point0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.1.0-beta",
    "id": 617375,
    "node_id": "MDc6UmVsZWFzZTYxNzM3NQ==",
    "tag_name": "v7.1.0-beta",
    "target_commitish": "master",
    "name": "Invalid release (pre-release status)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": true,
    "created_at": "2014-10-10T10:58:14Z",
    "published_at": "2014-10-10T10:59:34Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/265007",
        "id": 265007,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI2NTAwNw==",
        "name": "Dummy-7.1-beta.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 35726,
        "download_count": 6,
        "created_at": "2014-10-10T10:59:10Z",
        "updated_at": "2014-10-10T10:59:12Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v7.1.0-beta/Dummy-7.1-beta.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.1.0-beta",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.1.0-beta",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.0",
    "id": 556526,
    "node_id": "MDc6UmVsZWFzZTU1NjUyNg==",
    "tag_name": "v7.0",
    "target_commitish": "master",
    "name": "Invalid release (contains no files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:25:55Z",
    "published_at": "2014-09-14T19:27:25Z",
    "assets": [

    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v6.0",
    "id": 556525,
    "node_id": "MDc6UmVsZWFzZTU1NjUyNQ==",
    "tag_name": "v6.0",
    "target_commitish": "master",
    "name": "Latest valid release",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:24:55Z",
    "published_at": "2014-09-14T19:27:09Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/4823231",
        "id": 4823231,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDQ4MjMyMzE=",
        "name": "Dummy-6.0.alfred3workflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2017-09-14T12:22:03Z",
        "updated_at": "2017-09-14T12:22:08Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred3workflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/12368192",
        "id": 12368192,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDEyMzY4MTky",
        "name": "Dummy-6.0.alfred4workflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2019-05-03T12:14:14Z",
        "updated_at": "2019-05-03T12:14:15Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred4workflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247310",
        "id": 247310,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMxMA==",
        "name": "Dummy-6.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 585,
        "created_at": "2014-09-23T18:59:00Z",
        "updated_at": "2014-09-23T18:59:01Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfredworkflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247311",
        "id": 247311,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMxMQ==",
        "name": "Dummy-6.0.zip",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/zip",
        "state": "uploaded",
        "size": 36063,
        "download_count": 2,
        "created_at": "2014-09-23T18:59:00Z",
        "updated_at": "2014-09-23T18:59:01Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.zip"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v6.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v6.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v5.0",
    "id": 556524,
    "node_id": "MDc6UmVsZWFzZTU1NjUyNA==",
    "tag_name": "v5.0",
    "target_commitish": "master",
    "name": "Invalid release (contains no files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:22:44Z",
    "published_at": "2014-09-14T19:26:30Z",
    "assets": [

    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v5.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v5.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v4.0",
    "id": 556356,
    "node_id": "MDc6UmVsZWFzZTU1NjM1Ng==",
    "tag_name": "v4.0",
    "target_commitish": "master",
    "name": "Invalid release (contains 2 .alfredworkflow files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:34:44Z",
    "published_at": "2014-09-14T16:36:34Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247308",
        "id": 247308,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwOA==",
        "name": "Dummy-4.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 694,
        "created_at": "2014-09-23T18:58:25Z",
        "updated_at": "2014-09-23T18:58:27Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.0.alfredworkflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247309",
        "id": 247309,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwOQ==",
        "name": "Dummy-4.1.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2014-09-23T18:58:26Z",
        "updated_at": "2014-09-23T18:58:27Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.1.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v4.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v4.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v3.0",
    "id": 556354,
    "node_id": "MDc6UmVsZWFzZTU1NjM1NA==",
    "tag_name": "v3.0",
    "target_commitish": "master",
    "name": "Invalid release (no .alfredworkflow file)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:34:16Z",
    "published_at": "2014-09-14T16:36:16Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247305",
        "id": 247305,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwNQ==",
        "name": "Dummy-3.0.zip",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/zip",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2014-09-23T18:57:53Z",
        "updated_at": "2014-09-23T18:57:54Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v3.0/Dummy-3.0.zip"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v3.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v3.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v2.0",
    "id": 556352,
    "node_id": "MDc6UmVsZWFzZTU1NjM1Mg==",
    "tag_name": "v2.0",
    "target_commitish": "master",
    "name": "",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:33:36Z",
    "published_at": "2014-09-14T16:35:47Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247300",
        "id": 247300,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwMA==",
        "name": "Dummy-2.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2014-09-23T18:57:19Z",
        "updated_at": "2014-09-23T18:57:21Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v2.0/Dummy-2.0.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v2.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v2.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v1.0",
    "id": 556350,
    "node_id": "MDc6UmVsZWFzZTU1NjM1MA==",
    "tag_name": "v1.0",
    "target_commitish": "master",
    "name": "",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:33:06Z",
    "published_at": "2014-09-14T16:35:25Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247299",
        "id": 247299,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzI5OQ==",
        "name": "Dummy-1.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2014-09-23T18:56:22Z",
        "updated_at": "2014-09-23T18:56:24Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v1.0/Dummy-1.0.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v1.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v1.0",
    "body": ""
  }
]`
)
