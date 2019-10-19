// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	aw "github.com/deanishe/awgo"
)

// Metadata is a Workflow Option. It sets a Workflow Updater based on
// a `metadata.json` file exported from Alfred 4+.
//
// URL is the location of the `metadata.json` file. Note: You *must*
// set `downloadurl` in the `metadata.json` file to the URL
// of your .alfredworkflow (or .alfred4workflow etc.) file.
func Metadata(url string) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := NewUpdater(&metadataSource{url: url, fetch: getURL},
			wf.Version(),
			filepath.Join(wf.CacheDir(), "_aw/update"),
		)
		return aw.Update(u)(wf)
	}
}

type metadataSource struct {
	url   string
	dl    *Download
	fetch func(URL string) ([]byte, error)
}

// Downloads implements Source.
func (src *metadataSource) Downloads() ([]Download, error) {
	if src.dl == nil {
		var (
			js  []byte
			dl  Download
			err error
		)
		if js, err = src.fetch(src.url); err != nil {
			return nil, err
		}
		if dl, err = parseMetadata(js); err != nil {
			return nil, err
		}
		src.dl = &dl
	}
	return []Download{*src.dl}, nil
}

// data model for metadata.json JSON.
type metadataRelease struct {
	Data struct {
		URL     string `json:"downloadurl"`
		Version string `json:"version"`
	} `json:"alfredworkflow"`
}

func parseMetadata(data []byte) (Download, error) {
	var (
		dl  Download
		rel *metadataRelease
		u   *url.URL
		v   SemVer
		err error
	)
	if err = json.Unmarshal(data, &rel); err != nil {
		return dl, err
	}
	if rel.Data.Version == "" {
		return dl, errors.New("empty version")
	}
	if rel.Data.URL == "" {
		return dl, errors.New("empty url")
	}
	if v, err = NewSemVer(rel.Data.Version); err != nil {
		return dl, err
	}
	dl.Version = v
	dl.URL = rel.Data.URL
	if u, err = url.Parse(rel.Data.URL); err != nil {
		return dl, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return dl, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	dl.Filename = filepath.Base(u.Path)
	m := rxWorkflowFile.FindStringSubmatch(dl.Filename)
	if len(m) != 2 {
		return dl, errors.New("not a workflow file")
	}

	return dl, nil
}
