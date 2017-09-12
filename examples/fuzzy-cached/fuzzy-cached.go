//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*
Command fuzzy-cached is a complete Alfred 3 workflow.

It demonstrates custom fuzzy-sortable structs and handling larger datasets
in AwGo, caching the data in a format that's more quickly loaded.

It filters a list of the books from the Gutenberg project. The list
(a TSV file) is downloaded on first run, parsed and cached to disk
using gob.

The gob file loads ~5 times faster than the TSV.

There are >45K books in the list.

This runs in ~1s on my machine, which is *really* pushing the limits of
acceptable performance, imo.

A dataset of this size would be better off in an sqlite database, which
can easily handle this amount of data.
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
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"git.deanishe.net/deanishe/awgo"
	"git.deanishe.net/deanishe/awgo/fuzzy"
	"git.deanishe.net/deanishe/awgo/util"
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
	fuzzy-cached --download
	fuzzy-cached -h|--version

Options:
    --download  Download list of books to cache.
	-h, --help  Show this message and exit.
	--version   Show version number and exit.
`
	sopts []fuzzy.Option
	wf    *aw.Workflow
)

func init() {
	sopts = []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.HelpURL("http://www.deanishe.net/"))
}

// Book is a single work on Gutenberg.org.
type Book struct {
	ID     int
	Author string
	Title  string
	URL    string // Page where you can download the book in multiple formats.
}

// Books is a slice of Book structs that implements fuzzy.Interface.
type Books []*Book

// Len implements sort.Interface
func (b Books) Len() int { return len(b) }

// Less implements sort.Interface
func (b Books) Less(i, j int) bool { return b[i].Title < b[j].Title }

// Swap implements sort.Interface
func (b Books) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// SortKey implements fuzzy.Interface.
func (b Books) SortKey(i int) string {
	return fmt.Sprintf("%s %s", b[i].Author, b[i].Title)
}

// Filter removes non-matching Book objects.
func (b Books) Filter(query string, max int) Books {
	hits := b[:0]

	s := fuzzy.New(b, sopts...)
	t := time.Now()
	res := s.Sort(query)
	log.Printf("[fuzzy] sorted %d books in %v", b.Len(), util.ReadableDuration(time.Now().Sub(t)))

	var n int
	for i, it := range b {
		r := res[i]
		// Ignore items that are no match (i.e. not all characters in query
		// are in the item) or whose score is below minScore.
		if r.Match && r.Score >= minScore {
			n++
			hits = append(hits, it)
			log.Printf("[fuzzy] %3d. score=%5.2f title=%s, author=%s",
				n, r.Score, it.Title, it.Author)
			if max > 0 && n == max {
				break
			}
		}
	}
	return hits
}

// loadFromGob reads the book list from the cache.
func loadFromGob(path string) (Books, error) {
	s := time.Now()
	b := Books{}
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	dec := gob.NewDecoder(fp)
	err = dec.Decode(&b)
	if err != nil {
		return nil, err
	}
	log.Printf("[gob] loaded %d books in %v", b.Len(), util.ReadableDuration(time.Now().Sub(s)))
	return b, nil
}

// saveToGob serialises the books to disk.
func saveToGob(b Books, path string) error {
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
	log.Printf("[download] fetching %s...", tsvURL)
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
	log.Printf("[download] saved %d bytes to %s (%v)", i, path, util.ReadableDuration(time.Now().Sub(s)))
	return nil
}

// loadFromTSV loads the list of books from a TSV file.
func loadFromTSV(path string) (Books, error) {
	s := time.Now()
	b := Books{}
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
			log.Printf("[tsv] bad record: %v : %v", record, err)
			continue
		}
		author, title, url = record[1], record[2], record[3]
		b = append(b, &Book{id, author, title, url})
	}
	log.Printf("[tsv] loaded %d loaded from %s", b.Len(), util.ShortenPath(path))
	log.Printf("[tsv] loaded %d books in %v", b.Len(), time.Now().Sub(s))
	return b, nil
}

// loadCachedBooks loads the books from cache if possible else returns nil.
func loadCachedBooks() Books {
	gobpath := filepath.Join(wf.DataDir(), "books.gob")
	if !util.PathExists(gobpath) {
		log.Printf("[cache] books not yet cached")
		return nil
	}
	b, err := loadFromGob(gobpath)
	if err != nil {
		wf.FatalError(err)
	}
	return b
}

// cacheBooks downloads books and generates cache.
func cacheBooks() error {
	csvpath := filepath.Join(wf.DataDir(), "books.tsv")
	gobpath := filepath.Join(wf.DataDir(), "books.gob")

	if !util.PathExists(csvpath) {
		if err := downloadTSV(csvpath); err != nil {
			return fmt.Errorf("couldn't download books: %v", err)
		}
	}

	b, err := loadFromTSV(csvpath)
	if err != nil {
		return fmt.Errorf("couldn't loads books from TSV file (%s): %v", csvpath, err)
	}
	err = saveToGob(b, gobpath)
	if err != nil {
		return fmt.Errorf("couldn't save books to GOB file (%s): %v", gobpath, err)
	}
	return nil
}

func run() {
	var query string

	// Version is parsed from info.plist
	args, err := docopt.Parse(usage, wf.Args(), true, wf.Version(), false)
	if err != nil {
		log.Fatalf("Error parsing CLI options : %v", err)
	}

	if v, ok := args["--download"].(bool); ok {
		if v == true {
			wf.TextErrors = true
			log.Printf("[main] downloading book list...")
			if err := cacheBooks(); err != nil {
				wf.FatalError(err)
			}
			log.Printf("[main] downloaded book list")
			return
		}
	}

	// Docopt values are interface{} :(
	if s, ok := args["<query>"].(string); ok {
		query = s
	}
	log.Printf("[main] query=%s", query)

	// Try to load books
	b := loadCachedBooks()
	if b == nil { // Books not yet cached
		// Send an info message to user and tell Alfred to run the
		// workflow again in 0.5 seconds
		wf.Rerun(0.5)
		wf.NewItem("Downloading books…").
			Icon(aw.IconInfo)
		wf.SendFeedback()
		// Call this program in background to cache books
		if err := aw.RunInBackground("download", exec.Command("./fuzzy-cached", "--download")); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// Filter books based on query
	hits := b.Filter(query, maxResults)

	// Feedback
	for _, book := range hits {
		wf.NewItem(book.Title).
			Subtitle(book.Author).
			Arg(book.URL).
			Valid(true)
	}

	log.Printf("[main] %d/%d books match \"%s\"", hits.Len(), b.Len(), query)

	wf.WarnEmpty("No books found", "Try a different query?")

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
