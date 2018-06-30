//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-09-09
//

/*
Workflow bookmarks demonstrates implementing fuzzy.Sortable.

(This is not strictly necessary, as the Workflow/Feedback structs also
implement fuzzy.Sortable.)

It loads your Safari bookmarks from ~/Library/Safari/Bookmarks.plist into the
Bookmarks slice, which implements fuzzy.Sortable and a Filter() method,
which returns another Bookmarks slice containing all bookmarks that match
the given query.

See bookmarks.go for the implementation.

Alfred natively allows you to search your Safari bookmarks, but not with
fuzzy search.
*/
package main

import (
	"log"

	"github.com/deanishe/awgo"
)

var (
	helpURL    = "http://www.deanishe.net"
	maxResults = 200
	wf         *aw.Workflow

	// Icon for bookmark filetype
	icon = &aw.Icon{
		Value: "com.apple.safari.bookmark",
		Type:  aw.IconTypeFileType,
	}
)

func init() {
	wf = aw.New(aw.HelpURL(helpURL), aw.MaxResults(maxResults))
}

func run() {
	var query string

	// Use wf.Args() to enable Magic Actions
	if args := wf.Args(); len(args) > 0 {
		query = args[0]
	}

	log.Printf("[main] query=%s", query)

	// ----------------------------------------------------------------
	// Load bookmarks
	bookmarks, err := loadBookmarks()
	if err != nil {
		wf.FatalError(err)
	}

	log.Printf("%d total bookmark(s)", len(bookmarks))

	// ----------------------------------------------------------------
	// Filter bookmarks based on user query
	if query != "" {
		bookmarks = bookmarks.Filter(query)
	}

	// ----------------------------------------------------------------
	// Generate results for Alfred
	for _, b := range bookmarks {

		wf.NewItem(b.Title).
			Subtitle(b.URL).
			Arg(b.URL).
			UID(b.Title + b.URL).
			Icon(icon).
			Valid(true)
	}

	// Send results
	wf.WarnEmpty("No matching bookmarks", "Try a different query?")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
