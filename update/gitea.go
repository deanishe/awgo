// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"net/url"
	"strings"

	aw "github.com/deanishe/awgo"
)

// Gitea is a Workflow Option. It sets a Workflow Updater for the specified Gitea repo.
// Repo name should be the URL of the repo, e.g. "git.deanishe.net/deanishe/alfred-ssh".
func Gitea(repo string) aw.Option {
	return newOption(&source{URL: giteaURL(repo), fetch: getURL})
}

func giteaURL(repo string) string {
	if repo == "" {
		return ""
	}
	u, err := url.Parse(repo)
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
