// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	aw "github.com/deanishe/awgo"
)

// 6 valid releases, including one prerelease
// v1.0, v2.0, v6.0, v7.1.0-beta, v9.0 (Alfred 4+ only), v10.0-beta
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

	var (
		data = mustRead("testdata/gitea-releases.json")
		dls  []Download
		err  error
	)

	src := &giteaSource{
		Repo: "deanishe/alfred-workflow-dummy",
		fetch: func(URL string) ([]byte, error) {
			return ioutil.ReadFile("testdata/empty.json")
		}}
	if dls, err = src.Downloads(); err != nil {
		t.Fatal("parse empty JSON")
	}
	if len(dls) != 0 {
		t.Fatal("releases in empty JSON")
	}

	if dls, err = parseGiteaReleases(data); err != nil {
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

	dls, err := parseGiteaReleases(mustRead("testdata/gitea-releases.json"))
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
