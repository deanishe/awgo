//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*

fuzzy-cached demonstrates how to handle larger datasets in awgo, caching
the data in a format that's more quickly loaded.

It filters a list of the books from the Gutenberg project. The list
(a TSV file) is downloaded on first run, parsed and cached to disk
using gob.

The gob file loads ~5 times faster than the TSV.

There are >45K books in the list.

This runs in ~0.5s on my machine, which is really pushing the limits of
acceptable performance, imo.

A dataset of this size would be better off in an sqlite database, which
can *easily* handle this amount of data.

This demo is a complete Alfred 3 workflow.
*/
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docopt/docopt-go"
	"gogs.deanishe.net/deanishe/awgo"
)

var (
	// maxResults is the maximum number of results to sent to Alfred
	maxResults = 50
	// tsvURL is the source of the workflow's data
	tsvURL = "https://raw.githubusercontent.com/deanishe/alfred-index-demo/master/src/books.tsv"
	usage  = `fuzzy-big [options] [<query>]

Usage:
	fuzzy-big <query>
	fuzzy-big -h|--version

Options:
	-h, --help  Show this message and exit.
	--version   Show version number and exit.
`
	wf *workflow.Workflow
)

func init() {
	wf = workflow.NewWorkflow(&workflow.Options{
		MaxResults: maxResults,
	})
}

// Book is a single work on Gutenberg.org.
type Book struct {
	ID     int
	Author string
	Title  string
	// Page where you can download the book in multiple formats.
	URL string
}

// Books is a sequence of Book structs that implements the Fuzzy interface.
type Books []Book

// Len implements sort.Interface
func (b Books) Len() int { return len(b) }

// Less implements sort.Interface
func (b Books) Less(i, j int) bool { return b[i].Title < b[j].Title }

// Swap implements sort.Interface
func (b Books) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Keywords implements the Fuzzy interface
func (b Books) Keywords(i int) string {
	return fmt.Sprintf("%v %v", b[i].Title, b[i].Author)
}

// loadFromGob reads the book list from the cache.
func loadFromGob(path string) (Books, error) {
	s := time.Now()
	books := Books{}
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	dec := gob.NewDecoder(fp)
	err = dec.Decode(&books)
	if err != nil {
		return nil, err
	}
	log.Printf("[gob] loaded %d books in %v", len(books), time.Now().Sub(s))
	return books, nil
}

// saveToGob serialises the books to disk.
func saveToGob(books Books, path string) error {
	s := time.Now()
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	enc := gob.NewEncoder(fp)
	err = enc.Encode(books)
	if err != nil {
		return err
	}
	log.Printf("[gob] saved %d books in %v", len(books), time.Now().Sub(s))
	return nil
}

// downloadTSV fetches the data source TSV from GitHub and saves it
// in the workflow's data directory.
func downloadTSV(path string) error {
	s := time.Now()
	log.Printf("Fetching %s...", tsvURL)
	r, err := http.Get(tsvURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	i, err := io.Copy(fp, r.Body)
	if err != nil {
		return err
	}
	log.Printf("Saved %d bytes to %s (%v)", i, path, time.Now().Sub(s))
	return nil
}

// loadFromTSV loads the list of books from a TSV file.
func loadFromTSV(path string) (Books, error) {
	s := time.Now()
	books := Books{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bufio.NewReader(f))
	r.Comma, r.FieldsPerRecord = '\t', 4
	var id int
	var author, title, url string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// log.Printf("book=%v", record)
		id, err = strconv.Atoi(record[0])
		if err != nil {
			log.Printf("Bad record: %v : %v", record, err)
			continue
		}
		author, title, url = record[1], record[2], record[3]
		books = append(books, Book{id, author, title, url})
		// books = append(books, record...)
	}
	log.Printf("%d books loaded from %s", len(books), workflow.ShortenPath(path))
	log.Printf("[tsv] loaded %d books in %v", len(books), time.Now().Sub(s))
	return books, nil
}

// loadBooks loads the Gutenberg books from the cache. If the cache
// file doesn't exist, the source data is downloaded and the cache
// generated.
func loadBooks() Books {
	csvpath := filepath.Join(wf.DataDir(), "books.tsv")
	gobpath := filepath.Join(wf.DataDir(), "books.gob")
	if workflow.PathExists(gobpath) {
		books, err := loadFromGob(gobpath)
		if err != nil {
			wf.FatalError(err)
		}
		return books
	}

	if !workflow.PathExists(csvpath) {
		c := make(chan error)
		wf.Warn("Downloading book database…",
			"Try again in a few seconds.")
		go func(c chan error) {
			err := downloadTSV(csvpath)
			c <- err
		}(c)
		<-c // Wait for download to finish
	}
	books, err := loadFromTSV(csvpath)
	if err != nil {
		wf.FatalError(err)
	}
	err = saveToGob(books, gobpath)
	if err != nil {
		wf.FatalError(err)
	}
	return books
}

func run() {
	var query string
	var total int

	// Version is parsed from info.plist
	args, err := docopt.Parse(usage, nil, true, wf.Version(), false)
	if err != nil {
		log.Fatalf("Error parsing CLI options : %v", err)
	}

	// Docopt values are interface{} :(
	if s, ok := args["<query>"].(string); ok {
		query = s
	}

	books := loadBooks()
	total = len(books)

	// Feedback
	for _, book := range books {
		wf.NewItem(book.Title).
			Subtitle(book.Author).
			Arg(book.URL).
			SortKey(book.Title + " " + book.Author).
			Valid(true)
	}

	// Filter books based on query
	res := wf.Filter(query)
	log.Printf("%d/%d books match `%v`", len(res), total, query)

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
