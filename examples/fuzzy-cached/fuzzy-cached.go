//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*

fuzzy-cached demonstrates custom fuzzy-sortable objects and handling
larger datasets in AwGo, caching the data in a format that's
more quickly loaded.

It filters a list of the books from the Gutenberg project. The list
(a TSV file) is downloaded on first run, parsed and cached to disk
using gob.

The gob file loads ~5 times faster than the TSV.

There are >45K books in the list.

This runs in ~1s on my machine, which is *really* pushing the limits of
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

	"git.deanishe.net/deanishe/awgo"
	"github.com/docopt/docopt-go"
)

var (
	// maxResults is the maximum number of results to sent to Alfred
	maxResults = 200
	// minScore is the minimum score for a result
	minScore = 10.0
	// tsvURL is the source of the workflow's data
	tsvURL = "https://raw.githubusercontent.com/deanishe/alfred-index-demo/master/src/books.tsv"
	usage  = `fuzzy-cached [options] [<query>]

Usage:
	fuzzy-cached <query>
	fuzzy-cached -h|--version

Options:
	-h, --help  Show this message and exit.
	--version   Show version number and exit.
`
	sopts *aw.SortOptions
	wf    *aw.Workflow
)

func init() {
	sopts = aw.NewSortOptions()
	sopts.AdjacencyBonus = 5.0
	sopts.LeadingLetterPenalty = -0.1
	sopts.MaxLeadingLetterPenalty = -3.0
	sopts.UnmatchedLetterPenalty = -0.5
	wf = aw.NewWorkflow(&aw.Options{HelpURL: "http://www.deanishe.net/"})
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
type Books struct {
	Items []*Book
}

// Len implements sort.Interface
func (b *Books) Len() int { return len(b.Items) }

// Less implements sort.Interface
func (b *Books) Less(i, j int) bool { return b.Items[i].Title < b.Items[j].Title }

// Swap implements sort.Interface
func (b *Books) Swap(i, j int) { b.Items[i], b.Items[j] = b.Items[j], b.Items[i] }

// SortKey implements Sortable interface
func (b *Books) SortKey(i int) string {
	return fmt.Sprintf("%v %v", b.Items[i].Title, b.Items[i].Author)
}

// Filter removes non-matching Book objects.
func (b *Books) Filter(query string, max int) {
	items := b.Items[:0]
	s := aw.NewSorter(b, sopts)
	res := s.Sort(query)
	var n int
	for i, it := range b.Items {
		r := res[i]
		// Ignore items that are no match (i.e. not all characters in query
		// are in the item) or whose score is below minScore.
		if r.Match && r.Score >= minScore {
			n++
			items = append(items, it)
			log.Printf("%3d. score=%5.2f title=%s, author=%s",
				n, r.Score, it.Title, it.Author)
			if max > 0 && n == max {
				break
			}
		}
	}
	b.Items = items
}

// loadFromGob reads the book list from the cache.
func loadFromGob(path string) (*Books, error) {
	s := time.Now()
	b := &Books{}
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	dec := gob.NewDecoder(fp)
	err = dec.Decode(b)
	if err != nil {
		return nil, err
	}
	log.Printf("[gob] loaded %d books in %v", b.Len(), time.Now().Sub(s))
	return b, nil
}

// saveToGob serialises the books to disk.
func saveToGob(b *Books, path string) error {
	s := time.Now()
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	enc := gob.NewEncoder(fp)
	err = enc.Encode(b)
	if err != nil {
		return err
	}
	log.Printf("[gob] saved %d books in %v", b.Len(), time.Now().Sub(s))
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
func loadFromTSV(path string) (*Books, error) {
	s := time.Now()
	b := &Books{}
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
		b.Items = append(b.Items, &Book{id, author, title, url})
		// books = append(books, record...)
	}
	log.Printf("%d books loaded from %s", b.Len(), aw.ShortenPath(path))
	log.Printf("[tsv] loaded %d books in %v", b.Len(), time.Now().Sub(s))
	return b, nil
}

// loadBooks loads the Gutenberg books from the cache. If the cache
// file doesn't exist, the source data is downloaded and the cache
// generated.
func loadBooks() *Books {
	csvpath := filepath.Join(wf.DataDir(), "books.tsv")
	gobpath := filepath.Join(wf.DataDir(), "books.gob")
	if aw.PathExists(gobpath) {
		b, err := loadFromGob(gobpath)
		if err != nil {
			wf.FatalError(err)
		}
		return b
	}

	if !aw.PathExists(csvpath) {
		c := make(chan error)
		wf.Warn("Downloading book database…",
			"Try again in a few seconds.")
		go func(c chan error) {
			err := downloadTSV(csvpath)
			c <- err
		}(c)
		<-c // Wait for download to finish
	}
	b, err := loadFromTSV(csvpath)
	if err != nil {
		wf.FatalError(err)
	}
	err = saveToGob(b, gobpath)
	if err != nil {
		wf.FatalError(err)
	}
	return b
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

	b := loadBooks()
	total = b.Len()

	b.Filter(query, maxResults)

	// Feedback
	for _, book := range b.Items {
		wf.NewItem(book.Title).
			Subtitle(book.Author).
			Arg(book.URL).
			Valid(true)
	}

	// Filter books based on query
	log.Printf("%d/%d books match `%v`", b.Len(), total, query)

	wf.WarnEmpty("No books found", "Try a different query?")

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
