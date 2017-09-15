//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-03
//

package update

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	aw "github.com/deanishe/awgo"
)

const (
	ghBaseURL = "https://api.github.com/repos/"
)

// GitHub is a Workflow Option. It sets a Workflow Updater for the specified GitHub repo.
// Repo name should be of the form "username/repo", e.g. "deanishe/alfred-ssh".
func GitHub(repo string) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := New(wf, &GitHubReleaser{Repo: repo})
		return aw.Update(u)(wf)
	}
}

// GitHubReleaser updates from a GitHub repo's releases. Repo should be in
// the form "username/reponame", e.g. "deanishe/alfred-ssh". Releases
// are marked as pre-releases based on the "This is a pre-release"
// checkbox on the website, *not* the version number/tag.
type GitHubReleaser struct {
	Repo     string     // Repo name in form username/repo
	releases []*Release // GitHub releases for Repo
}

// Releases implements Releaser. Returns a slice of available releases that
// contain an .alfredworkflow file.
func (gh *GitHubReleaser) Releases() ([]*Release, error) {
	if gh.releases == nil {
		gh.releases = []*Release{}
		// rels := []*Release{}
		js, err := getURL(gh.url())
		if err != nil {
			log.Printf("Error fetching GitHub releases: %s", err)
			return nil, err
		}
		// log.Printf("%d bytes of JSON", len(js))
		rels, err := parseGitHubReleases(js)
		if err != nil {
			log.Printf("Error parsing GitHub releases: %s", err)
			return nil, err
		}
		gh.releases = rels
	}
	log.Printf("%d release(s) in repo %s", len(gh.releases), gh.Repo)
	return gh.releases, nil
}

func (gh *GitHubReleaser) url() *url.URL {
	u, _ := url.Parse(fmt.Sprintf("%s%s/releases", ghBaseURL, gh.Repo))
	return u
}

// ghRelease is the data model for GitHub releases JSON.
type ghRelease struct {
	Name       string     `json:"name"`
	Prerelease bool       `json:"prerelease"`
	Assets     []*ghAsset `json:"assets"`
	Tag        string     `json:"tag_name"`
}

// ghAsset is the data model for GitHub releases JSON.
type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// parseGitHubReleases parses GitHub releases JSON.
func parseGitHubReleases(js []byte) ([]*Release, error) {
	ghrels := []*ghRelease{}
	rels := []*Release{}
	if err := json.Unmarshal(js, &ghrels); err != nil {
		return nil, err
	}
	for _, ghr := range ghrels {
		r, err := ghReleaseToRelease(ghr)
		if err != nil {
			log.Printf("invalid release: %s", err)
		} else {
			rels = append(rels, r)
		}
	}
	return rels, nil
}

func ghReleaseToRelease(ghr *ghRelease) (*Release, error) {
	rel := &Release{Prerelease: ghr.Prerelease}
	// Check version
	v, err := NewSemVer(ghr.Tag)
	if err != nil {
		return nil, fmt.Errorf("invalid version/tag %q: %s", ghr.Tag, err)
	}
	rel.Version = v
	// Check files (assets)
	assets := []*ghAsset{}
	assets3 := []*ghAsset{} // .alfred3workflow files
	for _, gha := range ghr.Assets {
		if strings.HasSuffix(gha.Name, ".alfredworkflow") {
			assets = append(assets, gha)
		} else if strings.HasSuffix(gha.Name, ".alfred3workflow") {
			assets3 = append(assets3, gha)
		}
	}

	// Prefer .alfred3workflow files if present
	if len(assets3) > 0 {
		assets = assets3
	}

	// Reject bad releases
	if len(assets) > 1 {
		return nil, fmt.Errorf("multiple (%d) workflow files in release %s", len(assets), ghr.Tag)
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("no workflow files in release %s", ghr.Tag)
	}

	rel.Filename = assets[0].Name
	u, err := url.Parse(assets[0].URL)
	if err != nil {
		return nil, err
	}
	rel.URL = u
	return rel, nil
}
