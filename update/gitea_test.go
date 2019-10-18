// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aw "github.com/deanishe/awgo"
)

// 6 valid releases, including one prerelease
// v1.0, v2.0, v6.0, v7.1.0-beta, v9.0 (Alfred 4+ only), v10.0-beta
var testGiteaDownloads = []Download{
	// Latest version for Alfred 4
	{
		URL:        "https://git.deanishe.net/attachments/8c1b2442-eba2-4740-91b3-c57dab219096",
		Filename:   "Dummy-10.0-beta.alfredworkflow",
		Version:    mustVersion("v10.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 4
	{
		URL:        "https://git.deanishe.net/attachments/acd4dc64-1c85-4d23-b053-711bb4f976c5",
		Filename:   "Dummy-9.0.alfred4workflow",
		Version:    mustVersion("v9.0"),
		Prerelease: false,
	},
	// Latest version for Alfred 3
	{
		URL:        "https://git.deanishe.net/attachments/36d70923-d65d-4670-a1c1-adb5d6980b0c",
		Filename:   "Dummy-7.1-beta.alfredworkflow",
		Version:    mustVersion("v7.1.0-beta"),
		Prerelease: true,
	},
	// Latest stable version for Alfred 3
	{
		URL:        "https://git.deanishe.net/attachments/13392981-721e-4880-b2a9-aad50225d0af",
		Filename:   "Dummy-6.0.alfred4workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://git.deanishe.net/attachments/eb86751a-7f31-49f0-be4c-1dd1e0557c9d",
		Filename:   "Dummy-6.0.alfred3workflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://git.deanishe.net/attachments/61aa34a1-1877-4a41-ae50-01c18c8e2598",
		Filename:   "Dummy-6.0.alfredworkflow",
		Version:    mustVersion("v6.0"),
		Prerelease: false,
	},
	{
		URL:        "https://git.deanishe.net/attachments/03a01b52-93bc-48f0-9b09-37ba212a03fd",
		Filename:   "Dummy-2.0.alfredworkflow",
		Version:    mustVersion("v2.0"),
		Prerelease: false,
	},
	{
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
	dls, err = src.Downloads()
	require.Nil(t, err, "parse empty JSON failed")
	assert.Equal(t, 0, len(dls), "releases in empty JSON")

	dls, err = parseGiteaReleases(data)
	require.Nil(t, err, "parse Gitea JSON failed")
	assert.Equal(t, testGiteaDownloads, dls, "unexpected downloads")
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
		td := td
		t.Run(td.repo, func(t *testing.T) {
			t.Parallel()
			src := &giteaSource{Repo: td.repo}
			assert.Equal(t, td.url, src.url(), "unexpected URL")
		})
	}
}

func TestGiteaUpdater(t *testing.T) {
	t.Parallel()
	withTempDir(func(dir string) {
		src := makeGiteaSource()
		dls, err := src.Downloads()
		require.Nil(t, err, "src.Downloads() failed")
		assert.Equal(t, testGiteaDownloads, dls, "unexpected downloads")

		// invalid versions
		_, err = NewUpdater(src, "", dir)
		assert.NotNil(t, err, "accepted empty version")
		_, err = NewUpdater(src, "stan", dir)
		assert.NotNil(t, err, "accepted invalid version")

		u, err := NewUpdater(src, "0.2.2", dir)
		require.Nil(t, err, "create updater failed")

		// Update releases
		err = u.CheckForUpdate()
		require.Nil(t, err, "retrieve releases failed")

		// Check info is cached
		u2, err := NewUpdater(src, "0.2.2", dir)
		require.Nil(t, err, "create updater failed")
		assert.Equal(t, u.CurrentVersion, u2.CurrentVersion, "differing versions")
		assert.True(t, u2.LastCheck.Equal(u.LastCheck), "differing LastCheck")
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
