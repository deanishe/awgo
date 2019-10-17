// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	aw "github.com/deanishe/awgo"
)

// Gitea is a Workflow Option. It sets a Workflow Updater for the specified Gitea repo.
// Repo name should be the URL of the repo, e.g. "git.deanishe.net/deanishe/alfred-ssh".
func Gitea(repo string) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := NewUpdater(&giteaSource{Repo: repo, fetch: getURL},
			wf.Version(),
			filepath.Join(wf.CacheDir(), "_aw/update"),
		)
		return aw.Update(u)(wf)
	}
}

type giteaSource struct {
	Repo  string
	dls   []Download
	fetch func(URL string) ([]byte, error)
}

// Downloads implements Source.
func (src *giteaSource) Downloads() ([]Download, error) {
	if src.dls == nil {
		src.dls = []Download{}
		js, err := src.fetch(src.url())
		if err != nil {
			return nil, err
		}
		if src.dls, err = parseGiteaReleases(js); err != nil {
			return nil, err
		}
	}
	log.Printf("%d download(s) in repo %s", len(src.dls), src.Repo)
	return src.dls, nil
}

func (src *giteaSource) url() string {
	if src.Repo == "" {
		return ""
	}
	u, err := url.Parse(src.Repo)
	if err != nil {
		return ""
	}
	// If no scheme is specified, assume HTTPS and re-parse URL.
	// This is necessary because URL.Host isn't present on URLs
	// without a scheme (hostname is added to path)
	if u.Scheme == "" {
		u.Scheme = "https"
		u, err = url.Parse(u.String())
		if err != nil {
			return ""
		}
	}
	if u.Host == "" {
		return ""
	}
	path := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(path) != 2 {
		return ""
	}

	u.Path = fmt.Sprintf("/api/v1/repos/%s/%s/releases", path[0], path[1])

	return u.String()
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
	Name       string `json:"name"`
	URL        string `json:"browser_download_url"`
	MinVersion SemVer `json:"-"`
}

// parseGiteaReleases parses Gitea releases JSON.
func parseGiteaReleases(js []byte) ([]Download, error) {
	var (
		rels []*giteaRelease
		dls  []Download
	)
	if err := json.Unmarshal(js, &rels); err != nil {
		return nil, err
	}
	for _, r := range rels {
		if len(r.Assets) == 0 {
			continue
		}
		v, err := NewSemVer(r.Tag)
		if err != nil {
			log.Printf("ignored release %s: not semantic: %v", r.Tag, err)
			continue
		}
		var all []Download
		for _, a := range r.Assets {
			m := rxWorkflowFile.FindStringSubmatch(a.Name)
			if len(m) != 2 {
				log.Printf("ignored release %s: no workflow files", r.Tag)
				continue
			}
			w := Download{
				URL:        a.URL,
				Filename:   a.Name,
				Version:    v,
				Prerelease: r.Prerelease,
			}
			all = append(all, w)
		}
		if err := validRelease(all); err != nil {
			log.Printf("ignored release %s: %v", r.Tag, err)
			continue
		}
		dls = append(dls, all...)
	}
	sort.Sort(sort.Reverse(byVersion(dls)))
	return dls, nil
}
