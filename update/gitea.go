// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	aw "github.com/deanishe/awgo"
)

// Gitea is a Workflow Option. It sets a Workflow Updater for the specified Gitea repo.
// Repo name should be of the form "username/repo", e.g. "git.deanishe.net/deanishe/alfred-ssh".
func Gitea(repo string) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := New(wf, &giteaReleaser{Repo: repo, fetch: getURL})
		return aw.Update(u)(wf)
	}
}

// giteaReleaser updates from a Gitea repo's releases. Repo should be in
// the form "domain.tld/username/reponame", e.g.
// "git.deanishe.net/deanishe/alfred-ssh". Releases
// are marked as pre-releases based on the "This is a pre-release"
// checkbox on the website, *not* the version number/tag.
type giteaReleaser struct {
	Repo     string     // Repo name in form domain.tld/username/repo
	releases []*Release // GitHub releases for Repo
	fetch    func(*url.URL) ([]byte, error)
}

// Releases implements Releaser. Returns a slice of available releases that
// contain an .alfredworkflow file.
func (gr *giteaReleaser) Releases() ([]*Release, error) {
	if gr.releases == nil {
		gr.releases = []*Release{}
		// rels := []*Release{}
		js, err := gr.fetch(gr.url())
		if err != nil {
			log.Printf("[ERROR] fetch Gitea releases: %s", err)
			return nil, err
		}
		// log.Printf("%d bytes of JSON", len(js))
		rs, err := parseGiteaReleases(js)
		if err != nil {
			log.Printf("[ERROR] parse Gitea releases: %s", err)
			return nil, err
		}
		gr.releases = rs
	}
	log.Printf("%d release(s) in repo %s", len(gr.releases), gr.Repo)
	return gr.releases, nil
}

func (gr *giteaReleaser) url() *url.URL {
	if gr.Repo == "" {
		return nil
	}
	u, err := url.Parse(gr.Repo)
	if err != nil {
		return nil
	}
	// If no scheme is specified, assume HTTPS and re-parse URL.
	// This is necessary because URL.Host isn't present on URLs
	// without a scheme (hostname is added to path)
	if u.Scheme == "" {
		u.Scheme = "https"
		u, err = url.Parse(u.String())
		if err != nil {
			return nil
		}
	}
	if u.Host == "" {
		return nil
	}
	path := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(path) != 2 {
		return nil
	}

	u.Path = fmt.Sprintf("/api/v1/repos/%s/%s/releases", path[0], path[1])

	return u
}

// giteaRelease is the data model for Gitea releases JSON.
type giteaRelease struct {
	Name       string        `json:"name"`
	Prerelease bool          `json:"prerelease"`
	Assets     []*giteaAsset `json:"assets"`
	Tag        string        `json:"tag_name"`
}

// giteaAsset is the data model for GitHub releases JSON.
type giteaAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// parseGiteaReleases parses Gitea releases JSON.
func parseGiteaReleases(js []byte) ([]*Release, error) {
	var (
		grs = []*giteaRelease{}
		rs  = []*Release{}
	)
	if err := json.Unmarshal(js, &grs); err != nil {
		return nil, err
	}
	for _, gr := range grs {
		r, err := giteaReleaseToRelease(gr)
		if err != nil {
			log.Printf("invalid release: %s", err)
		} else {
			rs = append(rs, r)
		}
	}
	return rs, nil
}

func giteaReleaseToRelease(gr *giteaRelease) (*Release, error) {
	r := &Release{Prerelease: gr.Prerelease}
	// Check version
	v, err := NewSemVer(gr.Tag)
	if err != nil {
		return nil, fmt.Errorf("invalid version/tag %q: %s", gr.Tag, err)
	}
	r.Version = v
	// Check files (assets)
	assets := []*giteaAsset{}
	assets3 := []*giteaAsset{} // .alfred3workflow files
	for _, a := range gr.Assets {
		if strings.HasSuffix(a.Name, ".alfredworkflow") {
			assets = append(assets, a)
		} else if strings.HasSuffix(a.Name, ".alfred3workflow") {
			assets3 = append(assets3, a)
		}
	}

	// Prefer .alfred3workflow files if present
	if len(assets3) > 0 {
		assets = assets3
	}

	// Reject bad releases
	if len(assets) > 1 {
		return nil, fmt.Errorf("multiple (%d) workflow files in release %s", len(assets), gr.Tag)
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("no workflow files in release %s", gr.Tag)
	}

	r.Filename = assets[0].Name
	u, err := url.Parse(assets[0].URL)
	if err != nil {
		return nil, err
	}
	r.URL = u
	return r, nil
}
