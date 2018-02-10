//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-09-09
//

/*
Workflow bookmarks demonstrates implementing fuzzy.Interface on your own structs.

(This is not strictly necessary, as the Workflow/Feedback structs also
implement fuzzy.Interface.)

It loads your Safari bookmarks from ~/Library/Safari/Bookmarks.plist into the
Bookmarks struct, which implements fuzzy.Interface and a Filter() method,
which returns another Bookmarks struct containing all bookmarks that match
the given query.

See bookmarks.go for the implementation.

Alfred also allows you to search your Safari bookmarks, but not with fuzzy
search.
*/
package main

import (
	"fmt"
	"log"

	"github.com/deanishe/awgo"
	"github.com/docopt/docopt-go"
)

var (
	helpURL    = "http://www.deanishe.net"
	maxResults = 200
	wf         *aw.Workflow
	icon       = &aw.Icon{ // Icon for bookmark filetype
		Value: "com.apple.safari.bookmark",
		Type:  aw.IconTypeFileType,
	}
)

var (
	usage = `bookmarks [options] [<query>]

Usage:
    bookmarks <query>
    bookmarks -h|--version

Options:
    -h, --help  Show this message and exit
    --version   Show version number and exit
`
)

func init() {
	wf = aw.New(aw.HelpURL(helpURL), aw.MaxResults(maxResults))
}

func run() {
	var query string

	// Parse command-line arguments
	args, err := docopt.Parse(usage, wf.Args(), true, wf.Version(), false)
	if err != nil {
		wf.Fatal(fmt.Sprintf("couldn't parse CLI flags: %v", err))
	}

	if s, ok := args["<query>"].(string); ok {
		query = s
	}
	log.Printf("[main] query=%s", query)

	// Load bookmarks
	bookmarks, err := loadBookmarks()
	if err != nil {
		wf.FatalError(err)
	}
	log.Printf("%d total bookmark(s)", len(bookmarks))

	// Filter bookmarks based on user query
	if query != "" {
		bookmarks = bookmarks.Filter(query)
	}

	// Generate results for Alfred
	for _, b := range bookmarks {
		wf.NewItem(b.Title).
			Subtitle(b.URL).
			Arg(b.URL).
			UID(b.Title + b.URL).
			Icon(icon).
			Valid(true)
	}

	wf.WarnEmpty("No matching bookmarks", "Try a different query?")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
