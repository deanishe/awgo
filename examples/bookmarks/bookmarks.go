//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-09-17
//

package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/DHowett/go-plist"
	"github.com/deanishe/awgo/fuzzy"
)

var (
	// Where Safari stores its bookmarks
	bookmarksPath = os.ExpandEnv("$HOME/Library/Safari/Bookmarks.plist")
)

// Bookmarks is a slice of Bookmark structs that implements fuzzy.Interface.
type Bookmarks []*Bookmark

// Implement sort.Interface.
func (b Bookmarks) Len() int           { return len(b) }
func (b Bookmarks) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b Bookmarks) Less(i, j int) bool { return b[i].Title < b[j].Title }

// SortKey implements fuzzy.Interface.
func (b Bookmarks) SortKey(i int) string {
	return fmt.Sprintf("%s %s", b[i].Title, b[i].Domain)
}

// Filter filters bookmarks against query using a fuzzy.Sorter.
func (b Bookmarks) Filter(query string) Bookmarks {
	hits := b[:0]
	srt := fuzzy.New(b)
	res := srt.Sort(query)
	for i, r := range res {
		if !r.Match {
			continue
		}
		hits = append(hits, b[i])
	}
	return hits
}

// Bookmark is a Safari bookmark.
type Bookmark struct {
	Title  string // Bookmark title
	Domain string // Domain of URL
	URL    string // Bookmark URL
}

// entry is an node in Safari's Bookmarks.plist file. This struct
// matches all types of nodes in the file.
type entry struct {
	Title    string            `plist:"Title"`
	Type     string            `plist:"WebBookmarkType"`
	URL      string            `plist:"URLString"`
	UUID     string            `plist:"WebBookmarkUUID"`
	URIDict  map[string]string `plist:"URIDictionary"`
	Children []*entry          `plist:"Children"`
}

// loadBookmarks parses the Safari bookmarks file.
func loadBookmarks() (Bookmarks, error) {
	file, err := os.Open(bookmarksPath)
	if err != nil {
		return nil, err
	}

	root := entry{}
	entries := []*entry{}
	dec := plist.NewDecoder(file)
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	// Recursively parse tree. The children of root are the top-level folders.
	for _, e := range root.Children {
		if e.Title == "BookmarksBar" || e.Title == "BookmarksMenu" {
			entries = append(entries, extractLeafEntries(e)...)
		}
	}

	// Convert raw entries to bookmarks
	bkm := []*Bookmark{}
	// Ignore dupes
	seen := map[string]bool{}

	for _, e := range entries {
		// Use title & URL for dupe key, as you probably want bookmarks
		// with different titles to stil be shown, even if the URL is
		// a duplicate
		key := e.Title + e.URL
		if seen[key] == true {
			continue
		}

		// Convert entry to Bookmark
		var title string
		if e.Title != "" {
			title = e.Title
		} else {
			title = e.URIDict["title"]
		}
		u, err := url.Parse(e.URL)
		if err != nil {
			log.Printf("couldn't parse URL \"%s\" (%s): %v", e.URL, title, err)
			continue
		}

		seen[key] = true
		bkm = append(bkm, &Bookmark{Title: title, Domain: u.Host, URL: e.URL})
	}

	return bkm, nil
}

// extractLeafEntries recursively finds all bookmarks under root.
func extractLeafEntries(root *entry) []*entry {
	entries := []*entry{}
	for _, e := range root.Children {
		// Bookmarks have type "WebBookmarkTypeLeaf"
		if e.Type == "WebBookmarkTypeLeaf" {
			entries = append(entries, e)
		} else if len(e.Children) > 0 { // Recursively add children
			entries = append(entries, extractLeafEntries(e)...)
		}
	}
	return entries
}
