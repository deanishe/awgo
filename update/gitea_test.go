// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"net/url"
	"testing"

	aw "github.com/deanishe/awgo"
)

func TestParseGitea(t *testing.T) {
	t.Parallel()

	gr := &giteaReleaser{Repo: "deanishe/alfred-workflow-dummy", fetch: func(URL *url.URL) ([]byte, error) {
		return []byte(giteaReleasesEmptyJSON), nil
	}}
	rs, err := gr.Releases()
	// rels, err := parseGitHubReleases([]byte(ghReleasesEmptyJSON))
	if err != nil {
		t.Fatal("Error parsing empty JSON.")
	}
	if len(rs) != 0 {
		t.Fatal("Found releases in empty JSON.")
	}
	rs, err = parseGiteaReleases([]byte(giteaReleasesJSON))
	if err != nil {
		t.Fatal("Couldn't parse Gitea JSON.")
	}
	if len(rs) != 4 {
		t.Fatalf("Found %d Gitea releases, not 4.", len(rs))
	}
}

// makeGHReleaser creates a new Gitea Releaser and populates its release cache.
func makeGiteaReleaser() *giteaReleaser {
	gr := &giteaReleaser{Repo: "git.deanishe.net/deanishe/nonexistent"}
	// Avoid network
	rs, _ := parseGitHubReleases([]byte(giteaReleasesJSON))
	gr.releases = rs
	return gr
}

func TestGiteaURL(t *testing.T) {
	t.Parallel()

	data := []struct {
		repo string
		url  string
	}{
		// Invalid input
		{"", ""},
		{"https://git.deanishe.net/api/v1/repos/deanishe/nonexistent/releases", ""},
		{"git.deanishe.net/deanishe", ""},
		// Valid URLs
		{"git.deanishe.net/deanishe/nonexistent", "https://git.deanishe.net/api/v1/repos/deanishe/nonexistent/releases"},
		{"https://git.deanishe.net/deanishe/nonexistent", "https://git.deanishe.net/api/v1/repos/deanishe/nonexistent/releases"},
		{"http://git.deanishe.net/deanishe/nonexistent", "http://git.deanishe.net/api/v1/repos/deanishe/nonexistent/releases"},
	}

	for _, td := range data {
		gr := &giteaReleaser{Repo: td.repo}
		u := gr.url()
		if u == nil {
			if td.url != "" {
				t.Errorf("Bad API URL for %q. Expected=%q, Got=nil", td.repo, td.url)
			}
			continue
		}

		v := gr.url().String()
		if v != td.url {
			t.Errorf("Bad API URL. Expected=%v, Got=%v", td.url, v)
		}
	}
}

func TestGiteaUpdater(t *testing.T) {
	t.Parallel()

	withVersioned("0.2.2", func(v *versioned) {

		gr := makeGiteaReleaser()

		// There are 4 valid releases (one prerelease)
		rs, err := gr.Releases()
		if err != nil {
			t.Fatalf("Error retrieving Gitea releases: %s", err)
		}
		if len(rs) != 4 {
			t.Fatalf("Found %d valid releases, not 4.", len(rs))
		}

		// v6.0 is available
		u, err := New(v, gr)
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

// Configure Workflow to update from a Gitea repo.
func ExampleGitea() {
	// Set source repo using Gitea Option
	wf := aw.New(Gitea("git.deanishe.net/deanishe/alfred-ssh"))
	// Is a check for a newer version due?
	fmt.Println(wf.UpdateCheckDue())
	// Output:
	// true
}

var (
	giteaReleasesEmptyJSON = `[]`
	// 4 valid releases, including one prerelease
	// v1.0, v2.0, v6.0 and v7.1.0-beta
	giteaReleasesJSON = `
[
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/6b3e403f-4151-4f59-8956-c4a848f36d4b",
        "created_at": "2019-04-06T19:03:54Z",
        "download_count": 0,
        "id": 18,
        "name": "Dummy-eight.alfredworkflow",
        "size": 36063,
        "uuid": "6b3e403f-4151-4f59-8956-c4a848f36d4b"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2018-12-07T16:03:23Z",
    "draft": false,
    "id": 61642,
    "name": "Invalid tag (non-semantic)",
    "prerelease": false,
    "published_at": "2018-12-07T16:03:23Z",
    "tag_name": "v8point0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v8point0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61642",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v8point0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/36d70923-d65d-4670-a1c1-adb5d6980b0c",
        "created_at": "2019-04-06T19:03:20Z",
        "download_count": 0,
        "id": 17,
        "name": "Dummy-7.1-beta.alfredworkflow",
        "size": 35726,
        "uuid": "36d70923-d65d-4670-a1c1-adb5d6980b0c"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-10-10T10:58:14Z",
    "draft": false,
    "id": 61643,
    "name": "Invalid release (pre-release status)",
    "prerelease": true,
    "published_at": "2014-10-10T10:58:14Z",
    "tag_name": "v7.1.0-beta",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.1.0-beta.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61643",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.1.0-beta.zip"
  },
  {
    "assets": [],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:25:55Z",
    "draft": false,
    "id": 61645,
    "name": "Invalid release (contains no files)",
    "prerelease": false,
    "published_at": "2014-09-14T19:25:55Z",
    "tag_name": "v7.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61645",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/683e09ce-5643-456b-82ab-9bd6d8d1bbb8",
        "created_at": "2019-04-06T19:01:30Z",
        "download_count": 0,
        "id": 14,
        "name": "Dummy-6.0.zip",
        "size": 36063,
        "uuid": "683e09ce-5643-456b-82ab-9bd6d8d1bbb8"
      },
      {
        "browser_download_url": "https://git.deanishe.net/attachments/eb86751a-7f31-49f0-be4c-1dd1e0557c9d",
        "created_at": "2019-04-06T19:01:30Z",
        "download_count": 0,
        "id": 15,
        "name": "Dummy-6.0.alfred3workflow",
        "size": 36063,
        "uuid": "eb86751a-7f31-49f0-be4c-1dd1e0557c9d"
      },
      {
        "browser_download_url": "https://git.deanishe.net/attachments/61aa34a1-1877-4a41-ae50-01c18c8e2598",
        "created_at": "2019-04-06T19:01:30Z",
        "download_count": 0,
        "id": 16,
        "name": "Dummy-6.0.alfredworkflow",
        "size": 36063,
        "uuid": "61aa34a1-1877-4a41-ae50-01c18c8e2598"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:24:55Z",
    "draft": false,
    "id": 61646,
    "name": "Latest valid release",
    "prerelease": false,
    "published_at": "2014-09-14T19:24:55Z",
    "tag_name": "v6.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v6.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61646",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v6.0.zip"
  },
  {
    "assets": [],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T19:22:44Z",
    "draft": false,
    "id": 61647,
    "name": "Invalid release (contains no files)",
    "prerelease": false,
    "published_at": "2014-09-14T19:22:44Z",
    "tag_name": "v5.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v5.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61647",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v5.0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/d13764be-c63a-4435-9104-e0df7e1b62c5",
        "created_at": "2019-04-06T19:00:21Z",
        "download_count": 0,
        "id": 12,
        "name": "Dummy-4.0.alfredworkflow",
        "size": 36063,
        "uuid": "d13764be-c63a-4435-9104-e0df7e1b62c5"
      },
      {
        "browser_download_url": "https://git.deanishe.net/attachments/75d6eadf-922e-4179-a179-af703e18f4f6",
        "created_at": "2019-04-06T19:00:21Z",
        "download_count": 0,
        "id": 13,
        "name": "Dummy-4.1.alfredworkflow",
        "size": 36063,
        "uuid": "75d6eadf-922e-4179-a179-af703e18f4f6"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:34:44Z",
    "draft": false,
    "id": 61648,
    "name": "Invalid release (contains 2 .alfredworkflow files)",
    "prerelease": false,
    "published_at": "2014-09-14T16:34:44Z",
    "tag_name": "v4.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v4.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61648",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v4.0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/d6e88cc4-1f2b-4cb2-9749-deb5f6a16e0e",
        "created_at": "2019-04-06T18:59:37Z",
        "download_count": 0,
        "id": 11,
        "name": "Dummy-3.0.zip",
        "size": 36063,
        "uuid": "d6e88cc4-1f2b-4cb2-9749-deb5f6a16e0e"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:34:16Z",
    "draft": false,
    "id": 61649,
    "name": "Invalid release (no .alfredworkflow file)",
    "prerelease": false,
    "published_at": "2014-09-14T16:34:16Z",
    "tag_name": "v3.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v3.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61649",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v3.0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/03a01b52-93bc-48f0-9b09-37ba212a03fd",
        "created_at": "2019-04-06T18:59:08Z",
        "download_count": 0,
        "id": 10,
        "name": "Dummy-2.0.alfredworkflow",
        "size": 36063,
        "uuid": "03a01b52-93bc-48f0-9b09-37ba212a03fd"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:33:36Z",
    "draft": false,
    "id": 61650,
    "name": "v2.0",
    "prerelease": false,
    "published_at": "2014-09-14T16:33:36Z",
    "tag_name": "v2.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v2.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61650",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v2.0.zip"
  },
  {
    "assets": [
      {
        "browser_download_url": "https://git.deanishe.net/attachments/d71ad702-cfce-46ba-aa26-2096d34ff97b",
        "created_at": "2019-04-06T18:58:02Z",
        "download_count": 0,
        "id": 9,
        "name": "Dummy-1.0.alfredworkflow",
        "size": 36063,
        "uuid": "d71ad702-cfce-46ba-aa26-2096d34ff97b"
      }
    ],
    "author": {
      "avatar_url": "https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon",
      "email": "deanishe@deanishe.net",
      "full_name": "",
      "id": 1,
      "language": "en-US",
      "login": "deanishe",
      "username": "deanishe"
    },
    "body": "",
    "created_at": "2014-09-14T16:33:06Z",
    "draft": false,
    "id": 61651,
    "name": "v1.0",
    "prerelease": false,
    "published_at": "2014-09-14T16:33:06Z",
    "tag_name": "v1.0",
    "tarball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v1.0.tar.gz",
    "target_commitish": "master",
    "url": "https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61651",
    "zipball_url": "https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v1.0.zip"
  }
]
`
)
