// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"net/url"
	"testing"

	aw "github.com/deanishe/awgo"
)

func TestParseGH(t *testing.T) {
	t.Parallel()

	ghr := &gitHubReleaser{Repo: "deanishe/alfred-workflow-dummy", fetch: func(URL *url.URL) ([]byte, error) {
		return []byte(ghReleasesEmptyJSON), nil
	}}
	rels, err := ghr.Releases()
	// rels, err := parseGitHubReleases([]byte(ghReleasesEmptyJSON))
	if err != nil {
		t.Fatal("Error parsing empty JSON.")
	}
	if len(rels) != 0 {
		t.Fatal("Found releases in empty JSON.")
	}
	rels, err = parseGitHubReleases([]byte(ghReleasesJSON))
	if err != nil {
		t.Fatal("Couldn't parse GitHub JSON.")
	}
	if len(rels) != 4 {
		t.Fatalf("Found %d GitHub releases, not 4.", len(rels))
	}
}

// makeGHReleaser creates a new GitHub Releaser and populates its release cache.
func makeGHReleaser() *gitHubReleaser {
	gh := &gitHubReleaser{Repo: "deanishe/nonexistent"}
	// Avoid network
	rels, _ := parseGitHubReleases([]byte(ghReleasesJSON))
	gh.releases = rels
	return gh
}

func TestGHUpdater(t *testing.T) {
	t.Parallel()

	withVersioned("0.2.2", func(v *versioned) {

		gh := makeGHReleaser()

		// There are 4 valid releases (one prerelease)
		rels, err := gh.Releases()
		if err != nil {
			t.Fatalf("Error retrieving GH releases: %s", err)
		}
		if len(rels) != 4 {
			t.Fatalf("Found %d valid releases, not 4.", len(rels))
		}

		u, err := New(v, gh)
		if err != nil {
			t.Fatalf("Error creating updater: %s", err)
		}
		u.CurrentVersion = mustVersion("2")

		// Update releases
		if err := u.CheckForUpdate(); err != nil {
			t.Fatalf("Couldn't retrieve releases: %s", err)
		}

		if !u.UpdateAvailable() {
			t.Fatal("No update found")
		}
		// v6.0 is the latest stable version
		u.CurrentVersion = mustVersion("6")
		if u.UpdateAvailable() {
			t.Fatal("Unexpectedly found update")
		}
		// Prerelease v7.1.0-beta is newer
		u.Prereleases = true
		if !u.UpdateAvailable() {
			t.Fatal("No update found")
		}
	})
}

// TestUpdates ensures an unconfigured workflow doesn't think it can update
func TestUpdates(t *testing.T) {
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
	// v1.0, v2.0, v6.0 and v7.1.0-beta
	ghReleasesJSON = `
[
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v8point0/Dummy-eight.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2018-12-07T16:04:24Z",
        "download_count": 0,
        "id": 10048629,
        "label": null,
        "name": "Dummy-eight.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDEwMDQ4NjI5",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2018-12-07T16:04:25Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/10048629"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2018-12-07T16:03:23Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v8point0",
    "id": 14412055,
    "name": "Invalid tag (non-semantic)",
    "node_id": "MDc6UmVsZWFzZTE0NDEyMDU1",
    "prerelease": false,
    "published_at": "2018-12-07T16:04:30Z",
    "tag_name": "v8point0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v8point0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/14412055",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v8point0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v7.1.0-beta/Dummy-7.1-beta.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-10-10T10:59:10Z",
        "download_count": 5,
        "id": 265007,
        "label": null,
        "name": "Dummy-7.1-beta.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI2NTAwNw==",
        "size": 35726,
        "state": "uploaded",
        "updated_at": "2014-10-10T10:59:12Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/265007"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-10-10T10:58:14Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.1.0-beta",
    "id": 617375,
    "name": "Invalid release (pre-release status)",
    "node_id": "MDc6UmVsZWFzZTYxNzM3NQ==",
    "prerelease": true,
    "published_at": "2014-10-10T10:59:34Z",
    "tag_name": "v7.1.0-beta",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.1.0-beta",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.1.0-beta"
  },
  {
    "assets": [],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:25:55Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.0",
    "id": 556526,
    "name": "Invalid release (contains no files)",
    "node_id": "MDc6UmVsZWFzZTU1NjUyNg==",
    "prerelease": false,
    "published_at": "2014-09-14T19:27:25Z",
    "tag_name": "v7.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred3workflow",
        "content_type": "application/octet-stream",
        "created_at": "2017-09-14T12:22:03Z",
        "download_count": 0,
        "id": 4823231,
        "label": null,
        "name": "Dummy-6.0.alfred3workflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDQ4MjMyMzE=",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2017-09-14T12:22:08Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/4823231"
      },
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-09-23T18:59:00Z",
        "download_count": 584,
        "id": 247310,
        "label": null,
        "name": "Dummy-6.0.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMxMA==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:59:01Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247310"
      },
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.zip",
        "content_type": "application/zip",
        "created_at": "2014-09-23T18:59:00Z",
        "download_count": 1,
        "id": 247311,
        "label": null,
        "name": "Dummy-6.0.zip",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMxMQ==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:59:01Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247311"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:24:55Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v6.0",
    "id": 556525,
    "name": "Latest valid release",
    "node_id": "MDc6UmVsZWFzZTU1NjUyNQ==",
    "prerelease": false,
    "published_at": "2014-09-14T19:27:09Z",
    "tag_name": "v6.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v6.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v6.0"
  },
  {
    "assets": [],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:22:44Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v5.0",
    "id": 556524,
    "name": "Invalid release (contains no files)",
    "node_id": "MDc6UmVsZWFzZTU1NjUyNA==",
    "prerelease": false,
    "published_at": "2014-09-14T19:26:30Z",
    "tag_name": "v5.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v5.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v5.0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.0.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-09-23T18:58:25Z",
        "download_count": 693,
        "id": 247308,
        "label": null,
        "name": "Dummy-4.0.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwOA==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:58:27Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247308"
      },
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.1.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-09-23T18:58:26Z",
        "download_count": 0,
        "id": 247309,
        "label": null,
        "name": "Dummy-4.1.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwOQ==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:58:27Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247309"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:34:44Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v4.0",
    "id": 556356,
    "name": "Invalid release (contains 2 .alfredworkflow files)",
    "node_id": "MDc6UmVsZWFzZTU1NjM1Ng==",
    "prerelease": false,
    "published_at": "2014-09-14T16:36:34Z",
    "tag_name": "v4.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v4.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v4.0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v3.0/Dummy-3.0.zip",
        "content_type": "application/zip",
        "created_at": "2014-09-23T18:57:53Z",
        "download_count": 0,
        "id": 247305,
        "label": null,
        "name": "Dummy-3.0.zip",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwNQ==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:57:54Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247305"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:34:16Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v3.0",
    "id": 556354,
    "name": "Invalid release (no .alfredworkflow file)",
    "node_id": "MDc6UmVsZWFzZTU1NjM1NA==",
    "prerelease": false,
    "published_at": "2014-09-14T16:36:16Z",
    "tag_name": "v3.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v3.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v3.0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v2.0/Dummy-2.0.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-09-23T18:57:19Z",
        "download_count": 0,
        "id": 247300,
        "label": null,
        "name": "Dummy-2.0.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzMwMA==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:57:21Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247300"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:33:36Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v2.0",
    "id": 556352,
    "name": "",
    "node_id": "MDc6UmVsZWFzZTU1NjM1Mg==",
    "prerelease": false,
    "published_at": "2014-09-14T16:35:47Z",
    "tag_name": "v2.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v2.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v2.0"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v1.0/Dummy-1.0.alfredworkflow",
        "content_type": "application/octet-stream",
        "created_at": "2014-09-23T18:56:22Z",
        "download_count": 0,
        "id": 247299,
        "label": null,
        "name": "Dummy-1.0.alfredworkflow",
        "node_id": "MDEyOlJlbGVhc2VBc3NldDI0NzI5OQ==",
        "size": 36063,
        "state": "uploaded",
        "updated_at": "2014-09-23T18:56:24Z",
        "uploader": {
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/deanishe",
          "id": 747913,
          "login": "deanishe",
          "node_id": "MDQ6VXNlcjc0NzkxMw==",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "type": "User",
          "url": "https://api.github.com/users/deanishe"
        },
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247299"
      }
    ],
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets",
    "author": {
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/deanishe",
      "id": 747913,
      "login": "deanishe",
      "node_id": "MDQ6VXNlcjc0NzkxMw==",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:33:06Z",
    "draft": false,
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v1.0",
    "id": 556350,
    "name": "",
    "node_id": "MDc6UmVsZWFzZTU1NjM1MA==",
    "prerelease": false,
    "published_at": "2014-09-14T16:35:25Z",
    "tag_name": "v1.0",
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v1.0",
    "target_commitish": "master",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets{?name,label}",
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v1.0"
  }
]`
)
