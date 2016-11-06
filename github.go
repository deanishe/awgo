//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-03
//

package aw

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
)

const (
	ghBaseURL = "https://api.github.com/repos/"
)

// GitHub updates from a GitHub repo's releases. Repo should be in the
// form "username/reponame", e.g. "deanishe/alfred-ssh". Releases
// are marked as pre-releases based on the "This is a pre-release"
// checkbox on the website, *not* the version number/tag.
type GitHub struct {
	Repo     string
	releases []*Release
}

// Releases implements Releaser.
func (gh *GitHub) Releases() ([]*Release, error) {
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

func (gh *GitHub) url() *url.URL {
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
			log.Printf("Invalid release: %s", err)
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
		return nil, fmt.Errorf("Invalid version/tag %q: %s", ghr.Tag, err)
	}
	rel.Version = v
	// Check files (assets)
	assets := []*ghAsset{}
	for _, gha := range ghr.Assets {
		if strings.HasSuffix(gha.Name, ".alfredworkflow") || strings.HasSuffix(gha.Name, ".alfred3workflow") {
			assets = append(assets, gha)
		}
	}
	if len(assets) > 1 {
		return nil, fmt.Errorf("Multiple (%d) workflow files in release.", len(assets))
	}
	if len(assets) == 0 {
		return nil, errors.New("No workflow files in release.")
	}
	rel.Filename = assets[0].Name
	u, err := url.Parse(assets[0].URL)
	if err != nil {
		return nil, err
	}
	rel.URL = u
	return rel, nil
}
