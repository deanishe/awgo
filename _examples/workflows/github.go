// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	apiURL     = "https://api.github.com/search/repositories?per_page=100"
	apiQuery   = "topic:alfred-workflow"
	apiHeaders = map[string]string{
		"Accept":     "application/vnd.github.mercy-preview+json",
		"User-Agent": "github.com/deanishe/awgo",
	}
	client *http.Client
)

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 60 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}
}

// Repo is a GitHub repo from the GitHub API.
type Repo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Owner       *repoOwner `json:"owner"`
	URL         string     `json:"html_url"`
	Stars       int64      `json:"stargazers_count"`
	Topics      []string   `json:"topics"`
	Lang        string     `json:"language"`
}

// FullName returns standard "owner/repo" format.
func (r *Repo) FullName() string {
	return fmt.Sprintf("%s/%s", r.Username(), r.Name)
}

// Username is GitHub user login.
func (r *Repo) Username() string { return r.Owner.Login }

// repoOwner is a helper struct for unmarshalling API JSON response.
type repoOwner struct {
	Login string `json:"login"`
}

// apiResponse is the top-level helper struct for unmarshalling API JSON response.
type apiResponse struct {
	Repos []*Repo `json:"items"`
	Total int     `json:"total_count"`
}

// fetchRepos fetches all repos with topic "alfred-workflow" from GitHub.
//
// It iterates through all pages of results, returning all matching repos.
func fetchRepos() ([]*Repo, error) {
	repos := []*Repo{}
	var (
		pageCount int
		pageNum   = 1
	)

	for {
		if pageCount != 0 && pageNum > pageCount {
			break
		}
		log.Printf("fetching page %d of %d ...", pageNum, pageCount)

		// Generate URL for next page of results
		URL, _ := url.Parse(apiURL)
		q := URL.Query()
		q.Set("page", fmt.Sprintf("%d", pageNum))
		q.Set("q", apiQuery)
		URL.RawQuery = q.Encode()

		log.Printf("fetching %s ...", URL)
		req, err := http.NewRequest("GET", URL.String(), nil)
		if err != nil {
			return nil, err
		}
		// Add headers to request (label feature isn't part of the standard API yet)
		for k, v := range apiHeaders {
			req.Header.Add(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		log.Printf("[%d] %s", resp.StatusCode, URL)
		if resp.StatusCode > 299 {
			return nil, errors.New(resp.Status)
		}

		// Parse response
		data, _ := ioutil.ReadAll(resp.Body)
		r := apiResponse{}
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, err
		}
		repos = append(repos, r.Repos...)

		// Populate pageCount if unset
		if pageCount == 0 {
			pageCount = r.Total / 100
			if math.Mod(float64(r.Total), 100.0) > 0.0 {
				pageCount++
			}
		}
		pageNum++

	}
	return repos, nil
}
