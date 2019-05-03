// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	aw "github.com/deanishe/awgo"
)

var testGiteaDownloads = []Download{
	// Latest version for Alfred 4
	Download{
		URL:        "https://git.deanishe.net/attachments/8c1b2442-eba2-4740-91b3-c57dab219096",
		Filename:   "Dummy-10.0-beta.alfredworkflow",
		Version:    mustVersion("v10.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 4
	Download{
		URL:        "https://git.deanishe.net/attachments/acd4dc64-1c85-4d23-b053-711bb4f976c5",
		Filename:   "Dummy-9.0.alfred4workflow",
		Version:    mustVersion("v9.0"),
		Prerelease: false,
	},
	// Latest version for Alfred 3
	Download{
		URL:        "https://git.deanishe.net/attachments/36d70923-d65d-4670-a1c1-adb5d6980b0c",
		Filename:   "Dummy-7.1-beta.alfredworkflow",
		Version:    mustVersion("v7.1.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 3
	Download{
		URL:        "https://git.deanishe.net/attachments/13392981-721e-4880-b2a9-aad50225d0af",
		Filename:   "Dummy-6.0.alfred4workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://git.deanishe.net/attachments/eb86751a-7f31-49f0-be4c-1dd1e0557c9d",
		Filename:   "Dummy-6.0.alfred3workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://git.deanishe.net/attachments/61aa34a1-1877-4a41-ae50-01c18c8e2598",
		Filename:   "Dummy-6.0.alfredworkflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://git.deanishe.net/attachments/03a01b52-93bc-48f0-9b09-37ba212a03fd",
		Filename:   "Dummy-2.0.alfredworkflow",
		Version:    mustVersion("v2.0"),
		Prerelease: false,
	},
	Download{
		URL:        "https://git.deanishe.net/attachments/d71ad702-cfce-46ba-aa26-2096d34ff97b",
		Filename:   "Dummy-1.0.alfredworkflow",
		Version:    mustVersion("v1.0"),
		Prerelease: false,
	},
}

func TestParseGitea(t *testing.T) {
	t.Parallel()

	src := &giteaSource{
		Repo: "deanishe/alfred-workflow-dummy",
		fetch: func(URL string) ([]byte, error) {
			return []byte(giteaReleasesEmptyJSON), nil
		}}
	dls, err := src.Downloads()
	if err != nil {
		t.Fatal("parse empty JSON")
	}
	if len(dls) != 0 {
		t.Fatal("releases in empty JSON")
	}

	dls, err = parseGiteaReleases([]byte(giteaReleasesJSON))
	if err != nil {
		t.Fatal("parse Gitea JSON.")
	}
	if len(dls) != len(testGiteaDownloads) {
		t.Fatalf("Wrong download count. Expected=%d, Got=%d", len(testGiteaDownloads), len(dls))
	}

	for i, w := range dls {
		w2 := testGiteaDownloads[i]
		if !reflect.DeepEqual(w, w2) {
			t.Fatalf("Download mismatch at pos %d. Expected=%#v, Got=%#v", i, w2, w)
		}
	}
}

func makeGiteaSource() *giteaSource {
	src := &giteaSource{Repo: "git.deanishe.net/deanishe/nonexistent"}
	dls, err := parseGiteaReleases([]byte(giteaReleasesJSON))
	if err != nil {
		panic(err)
	}
	src.dls = dls
	return src
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
		src := &giteaSource{Repo: td.repo}
		u := src.url()
		if u == "" {
			if td.url != "" {
				t.Errorf("Bad API URL for %q. Expected=%q, Got=nil", td.repo, td.url)
			}
			continue
		}

		v := src.url()
		if v != td.url {
			t.Errorf("Bad API URL. Expected=%v, Got=%v", td.url, v)
		}
	}
}

func TestGiteaUpdater(t *testing.T) {
	t.Parallel()
	withTempDir(func(dir string) {
		src := makeGiteaSource()
		dls, err := src.Downloads()
		if err != nil {
			t.Fatal(err)
		}
		if len(dls) != len(testGiteaDownloads) {
			t.Errorf("Wrong no. of downloads. Expected=%v, Got=%v", len(testGiteaDownloads), len(dls))
			for i, dl := range dls {
				log.Printf("download %d: %s", i, dl.Filename)
			}
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

		testUpdater("gitea", u, t)
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
	// 6 valid releases, including one prerelease
	// v1.0, v2.0, v6.0, v7.1.0-beta, v9.0 (Alfred 4+ only), v10.0-beta
	giteaReleasesJSON = `[{"id":64612,"tag_name":"v10.0-beta","target_commitish":"master","name":"Latest release (pre-release)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/64612","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v10.0-beta.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v10.0-beta.zip","draft":false,"prerelease":true,"created_at":"2019-05-03T12:27:30Z","published_at":"2019-05-03T12:27:30Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":20,"name":"Dummy-10.0-beta.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-05-03T14:46:11Z","uuid":"8c1b2442-eba2-4740-91b3-c57dab219096","browser_download_url":"https://git.deanishe.net/attachments/8c1b2442-eba2-4740-91b3-c57dab219096"}]},{"id":64613,"tag_name":"v9.0","target_commitish":"master","name":"Latest release (Alfred 4)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/64613","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v9.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v9.0.zip","draft":false,"prerelease":false,"created_at":"2019-05-03T12:24:12Z","published_at":"2019-05-03T12:24:12Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":19,"name":"Dummy-9.0.alfred4workflow","size":36063,"download_count":0,"created_at":"2019-05-03T14:45:38Z","uuid":"acd4dc64-1c85-4d23-b053-711bb4f976c5","browser_download_url":"https://git.deanishe.net/attachments/acd4dc64-1c85-4d23-b053-711bb4f976c5"}]},{"id":61642,"tag_name":"v8point0","target_commitish":"master","name":"Invalid tag (non-semantic)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61642","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v8point0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v8point0.zip","draft":false,"prerelease":false,"created_at":"2018-12-07T16:03:23Z","published_at":"2018-12-07T16:03:23Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":18,"name":"Dummy-eight.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T19:03:54Z","uuid":"6b3e403f-4151-4f59-8956-c4a848f36d4b","browser_download_url":"https://git.deanishe.net/attachments/6b3e403f-4151-4f59-8956-c4a848f36d4b"}]},{"id":61643,"tag_name":"v7.1.0-beta","target_commitish":"master","name":"Invalid release (pre-release status)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61643","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.1.0-beta.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.1.0-beta.zip","draft":false,"prerelease":true,"created_at":"2014-10-10T10:58:14Z","published_at":"2014-10-10T10:58:14Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":17,"name":"Dummy-7.1-beta.alfredworkflow","size":35726,"download_count":0,"created_at":"2019-04-06T19:03:20Z","uuid":"36d70923-d65d-4670-a1c1-adb5d6980b0c","browser_download_url":"https://git.deanishe.net/attachments/36d70923-d65d-4670-a1c1-adb5d6980b0c"}]},{"id":61645,"tag_name":"v7.0","target_commitish":"master","name":"Invalid release (contains no files)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61645","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v7.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T19:25:55Z","published_at":"2014-09-14T19:25:55Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[]},{"id":61646,"tag_name":"v6.0","target_commitish":"master","name":"Latest valid release","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61646","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v6.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v6.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T19:24:41Z","published_at":"2014-09-14T19:24:41Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":14,"name":"Dummy-6.0.zip","size":36063,"download_count":0,"created_at":"2019-04-06T19:01:30Z","uuid":"683e09ce-5643-456b-82ab-9bd6d8d1bbb8","browser_download_url":"https://git.deanishe.net/attachments/683e09ce-5643-456b-82ab-9bd6d8d1bbb8"},{"id":15,"name":"Dummy-6.0.alfred3workflow","size":36063,"download_count":0,"created_at":"2019-04-06T19:01:30Z","uuid":"eb86751a-7f31-49f0-be4c-1dd1e0557c9d","browser_download_url":"https://git.deanishe.net/attachments/eb86751a-7f31-49f0-be4c-1dd1e0557c9d"},{"id":16,"name":"Dummy-6.0.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T19:01:30Z","uuid":"61aa34a1-1877-4a41-ae50-01c18c8e2598","browser_download_url":"https://git.deanishe.net/attachments/61aa34a1-1877-4a41-ae50-01c18c8e2598"},{"id":21,"name":"Dummy-6.0.alfred4workflow","size":36063,"download_count":0,"created_at":"2019-05-03T19:28:16Z","uuid":"13392981-721e-4880-b2a9-aad50225d0af","browser_download_url":"https://git.deanishe.net/attachments/13392981-721e-4880-b2a9-aad50225d0af"}]},{"id":61647,"tag_name":"v5.0","target_commitish":"master","name":"Invalid release (contains no files)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61647","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v5.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v5.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T19:22:44Z","published_at":"2014-09-14T19:22:44Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[]},{"id":61648,"tag_name":"v4.0","target_commitish":"master","name":"Invalid release (contains 2 .alfredworkflow files)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61648","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v4.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v4.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T16:34:44Z","published_at":"2014-09-14T16:34:44Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":12,"name":"Dummy-4.0.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T19:00:21Z","uuid":"d13764be-c63a-4435-9104-e0df7e1b62c5","browser_download_url":"https://git.deanishe.net/attachments/d13764be-c63a-4435-9104-e0df7e1b62c5"},{"id":13,"name":"Dummy-4.1.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T19:00:21Z","uuid":"75d6eadf-922e-4179-a179-af703e18f4f6","browser_download_url":"https://git.deanishe.net/attachments/75d6eadf-922e-4179-a179-af703e18f4f6"}]},{"id":61649,"tag_name":"v3.0","target_commitish":"master","name":"Invalid release (no .alfredworkflow file)","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61649","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v3.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v3.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T16:34:16Z","published_at":"2014-09-14T16:34:16Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":11,"name":"Dummy-3.0.zip","size":36063,"download_count":0,"created_at":"2019-04-06T18:59:37Z","uuid":"d6e88cc4-1f2b-4cb2-9749-deb5f6a16e0e","browser_download_url":"https://git.deanishe.net/attachments/d6e88cc4-1f2b-4cb2-9749-deb5f6a16e0e"}]},{"id":61650,"tag_name":"v2.0","target_commitish":"master","name":"v2.0","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61650","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v2.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v2.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T16:33:36Z","published_at":"2014-09-14T16:33:36Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":10,"name":"Dummy-2.0.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T18:59:08Z","uuid":"03a01b52-93bc-48f0-9b09-37ba212a03fd","browser_download_url":"https://git.deanishe.net/attachments/03a01b52-93bc-48f0-9b09-37ba212a03fd"}]},{"id":61651,"tag_name":"v1.0","target_commitish":"master","name":"v1.0","body":"","url":"https://git.deanishe.net/api/v1/deanishe/alfred-workflow-dummy/releases/61651","tarball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v1.0.tar.gz","zipball_url":"https://git.deanishe.net/deanishe/alfred-workflow-dummy/archive/v1.0.zip","draft":false,"prerelease":false,"created_at":"2014-09-14T16:33:06Z","published_at":"2014-09-14T16:33:06Z","author":{"id":1,"login":"deanishe","full_name":"","email":"deanishe@deanishe.net","avatar_url":"https://secure.gravatar.com/avatar/f8a47e9dc5393dabf96054d4abb76478?d=identicon","language":"en-US","is_admin":true,"username":"deanishe"},"assets":[{"id":9,"name":"Dummy-1.0.alfredworkflow","size":36063,"download_count":0,"created_at":"2019-04-06T18:58:02Z","uuid":"d71ad702-cfce-46ba-aa26-2096d34ff97b","browser_download_url":"https://git.deanishe.net/attachments/d71ad702-cfce-46ba-aa26-2096d34ff97b"}]}]`
)
