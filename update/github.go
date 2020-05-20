// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"sort"

	aw "github.com/deanishe/awgo"
)

// matches filename of a compiled Alfred workflow
var rxWorkflowFile = regexp.MustCompile(`\.alfred(\d+)?workflow$`)

// GitHub is a Workflow Option. It sets a Workflow Updater for the specified GitHub repo.
// Repo name should be of the form "username/repo", e.g. "deanishe/alfred-ssh".
func GitHub(repo string) aw.Option {
	return newOption(&source{
		URL:   "https://api.github.com/repos/" + repo,
		fetch: getURL,
	})
}

// create new Updater option from Source.
func newOption(src Source) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := NewUpdater(src, wf.Version(), filepath.Join(wf.CacheDir(), "_aw/update"))
		return aw.Update(u)(wf)
	}
}

type source struct {
	URL   string
	dls   []Download
	fetch func(URL string) ([]byte, error)
}

// Downloads implements Source.
func (src *source) Downloads() ([]Download, error) {
	if src.dls != nil {
		return src.dls, nil
	}

	src.dls = []Download{}
	js, err := src.fetch(src.URL)
	if err != nil {
		return nil, err
	}
	if src.dls, err = parseReleases(js); err != nil {
		return nil, err
	}

	return src.dls, nil
}

// parse GitHub/Gitea releases JSON.
func parseReleases(js []byte) ([]Download, error) {
	var (
		dls  = []Download{}
		rels = []struct {
			Name       string `json:"name"`
			Prerelease bool   `json:"prerelease"`
			Assets     []struct {
				Name             string `json:"name"`
				URL              string `json:"browser_download_url"`
				MinAlfredVersion SemVer `json:"-"`
			} `json:"assets"`
			Tag string `json:"tag_name"`
		}{}
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

// Reject releases that contain multiple files with the same extension.
func validRelease(dls []Download) error {
	if len(dls) == 0 {
		return errors.New("empty slice")
	}
	dupes := map[string]int{}
	for _, dl := range dls {
		x := filepath.Ext(dl.Filename)
		dupes[x]++
	}
	for x, n := range dupes {
		if n > 1 {
			return fmt.Errorf("multiple files with extension %q", x)
		}
	}
	return nil
}
