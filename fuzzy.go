//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-10-30
//

package aw

import (
	"log"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Default bonuses and penalties for fuzzy sorting. To customise
// sorting behaviour, create a SortOptions object with NewSortOptions().
// Pass this to NewSorter(), or if you're using a Workflow object,
// set Workflow.SortOptions or Options.SortOptions.
const (
	DefaultAdjacencyBonus          = 5.0  // Bonus for adjacent matches
	DefaultSeparatorBonus          = 10.0 // Bonus if the match is after a separator
	DefaultCamelBonus              = 10.0 // Bonus if match is uppercase and previous is lower
	DefaultLeadingLetterPenalty    = -3.0 // Penalty applied for every letter in string before first match
	DefaultMaxLeadingLetterPenalty = -9.0 // Maximum penalty for leading letters
	DefaultUnmatchedLetterPenalty  = -1.0 // Penalty for every letter that doesn't match
	DefaultStripDiacritics         = true // Strip diacritics from sort keys if query is plain ASCII
)

var stripper transform.Transformer

func init() {
	stripper = transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
}

// Sortable makes the implementer fuzzy-sortable. It is a superset
// of sort.Interface (i.e. your struct must also implement sort.Interface).
type Sortable interface {
	// Return the string that should be compared to the sort query
	SortKey(i int) string
	sort.Interface
}

// Result stores the result of a single fuzzy ranking.
type Result struct {
	// Match is whether or not the string matched the query,
	// i.e. if all characters in the query are present,
	// in order, in the string.
	Match bool
	// Query is the query that was matched against.
	Query string
	// Score is how well the string matched the query.
	// Higher is better.
	Score float64
	// SortKey is the string Query was compared to.
	SortKey string
}

// SortOptions sets bonuses and penalties for Sorter/Match.
type SortOptions struct {
	AdjacencyBonus          float64 // Bonus for adjacent matches
	SeparatorBonus          float64 // Bonus if the match is after a separator
	CamelBonus              float64 // Bonus if match is uppercase and previous is lower
	LeadingLetterPenalty    float64 // Penalty applied for every letter in string before first match
	MaxLeadingLetterPenalty float64 // Maximum penalty for leading letters
	UnmatchedLetterPenalty  float64 // Penalty for every letter that doesn't match
	StripDiacritics         bool    //  Strip diacritics from sort keys if query is plain ASCII
}

// NewSortOptions creates a SortOptions object with the default values.
func NewSortOptions() *SortOptions {
	return &SortOptions{
		AdjacencyBonus:          DefaultAdjacencyBonus,
		SeparatorBonus:          DefaultSeparatorBonus,
		CamelBonus:              DefaultCamelBonus,
		LeadingLetterPenalty:    DefaultLeadingLetterPenalty,
		MaxLeadingLetterPenalty: DefaultMaxLeadingLetterPenalty,
		UnmatchedLetterPenalty:  DefaultUnmatchedLetterPenalty,
		StripDiacritics:         DefaultStripDiacritics,
	}
}

// Sorter sorts Data based on the query passsed to Sorter.Sort().
type Sorter struct {
	// Data is an object implementing Sortable interface
	Data Sortable
	// Options contains the bonuses and penalties
	Options *SortOptions

	query           string    // Search query
	stripDiacritics bool      // Whether sort keys need folding
	results         []*Result // Results of the fuzzy sort
}

// NewSorter returns a new Sorter. If opts is nil, Sorter is initialised
// with the default sort parameters.
func NewSorter(data Sortable, opts *SortOptions) *Sorter {
	if opts == nil {
		opts = NewSortOptions()
	}
	return &Sorter{
		Data:    data,
		Options: opts,
		results: make([]*Result, data.Len()),
	}
}

// Len implements sort.Interface.
func (s *Sorter) Len() int { return s.Data.Len() }

// Less implements sort.Interface.
func (s *Sorter) Less(i, j int) bool {
	a, b := s.results[i].Score, s.results[j].Score
	if a == b {
		// Normal comparison because A comes before B.
		return s.Data.Less(i, j)
	}
	// Reverse comparison because higher score is better.
	return b < a
}

// Swap implements sort.Interface.
func (s *Sorter) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
	s.Data.Swap(i, j)
}

// Sort sorts data against query.
func (s *Sorter) Sort(query string) []*Result {
	s.results = make([]*Result, s.Data.Len())
	s.query = query
	if isASCII(query) && s.Options.StripDiacritics {
		s.stripDiacritics = true
	}
	// Generate matches for Data, then call sort.Sort()
	for i := 0; i < s.Data.Len(); i++ {
		key := s.Data.SortKey(i)
		s.results[i] = s.Match(key)
	}
	sort.Sort(s)
	return s.results
}

// Match scores str against query using fuzzy matching and the specified sort options.
func (s *Sorter) Match(str string) *Result {
	if s.stripDiacritics {
		str = stripDiacritics(str)
	}

	var (
		match    = false
		score    = 0.0
		uStr     = []rune(str)
		uQuery   = []rune(s.query)
		strLen   = len(uStr)
		queryLen = len(uQuery)
	)
	var (
		queryIdx, strIdx                   int
		newScore, penalty, bestLetterScore float64
		queryChar, queryLower              string
		strChar, strLower, strUpper        string
		bestLetter, bestLower              string
		advanced, queryRepeat              bool
		nextMatch, rematch                 bool
		prevMatched, prevLower             bool
		prevSeparator                      = true
	)

	// Loop through each character in str
	for strIdx != strLen {
		strChar = string(uStr[strIdx])

		if queryIdx != queryLen {
			queryChar = string(uQuery[queryIdx])
		} else {
			queryChar = ""
		}

		queryLower = strings.ToLower(queryChar)
		strLower = strings.ToLower(strChar)
		strUpper = strings.ToUpper(strChar)

		if queryChar != "" && queryLower == strLower {
			nextMatch = true
		} else {
			nextMatch = false
		}
		if bestLetter != "" && bestLower == strLower {
			rematch = true
		} else {
			rematch = false
		}

		if nextMatch && bestLetter != "" {
			advanced = true
		} else {
			advanced = false
		}

		if bestLetter != "" && strChar != "" && bestLower == queryLower {
			queryRepeat = true
		} else {
			queryRepeat = false
		}

		if advanced || queryRepeat {
			score += bestLetterScore
			// matchedIdx = append(matchedIdx, bestLetterIdx)
			bestLetter = ""
			bestLower = ""
			bestLetterScore = 0.0
		}

		if nextMatch || rematch {
			newScore = 0.0

			// Apply penalty for letters before first match
			if queryIdx == 0 {
				penalty = float64(strIdx) * s.Options.LeadingLetterPenalty
				if penalty <= s.Options.MaxLeadingLetterPenalty {
					penalty = s.Options.MaxLeadingLetterPenalty
				}
				score += penalty
			}

			// Apply bonus for consecutive matches
			if prevMatched {
				newScore += s.Options.AdjacencyBonus
			}

			// Apply bonus for match after separator
			if prevSeparator {
				newScore += s.Options.SeparatorBonus
			}

			// Apply bonus across camel case boundaries
			if prevLower && strChar == strUpper && strLower != strUpper {
				newScore += s.Options.CamelBonus
			}

			// Update query index if next query letter was matched
			if nextMatch {
				queryIdx++
			}

			// Update best letter in key, which may be for a "next" letter
			// or a reMatch
			if newScore >= bestLetterScore {

				if bestLetter != "" {
					score += s.Options.UnmatchedLetterPenalty
				}

				bestLetter = strChar
				bestLower = strings.ToLower(bestLetter)
				bestLetterScore = newScore
			}

			prevMatched = true
		} else {
			score += s.Options.UnmatchedLetterPenalty
			prevMatched = false
		}

		// IsLetter check
		if strChar == strLower && strLower != strUpper {
			prevLower = true
		} else {
			prevLower = false
		}
		if strChar == "_" || strChar == " " {
			prevSeparator = true
		} else {
			prevSeparator = false
		}

		strIdx++
	}

	if bestLetter != "" {
		score += bestLetterScore
		// matchedIdx = append(matchedIdx, bestLetterIdx)
	}

	if queryIdx == queryLen {
		match = true
	}

	// log.Printf("query=%#v, str=%#v", match=%v, score=%v, query, str, match, score)
	return &Result{match, s.query, score, str}
}

// Sort sorts data against query. Convenience that creates and
// uses a Sorter with the default settings.
func Sort(data Sortable, query string) []*Result {
	s := NewSorter(data, nil)
	return s.Sort(query)
}

// stringSlice implements sort.Interface for []string.
// It is a helper for SortStrings.
type stringSlice struct {
	data []string
}

// Len etc. implement sort.Interface.
func (s stringSlice) Len() int           { return len(s.data) }
func (s stringSlice) Less(i, j int) bool { return s.data[i] < s.data[j] }
func (s stringSlice) Swap(i, j int)      { s.data[i], s.data[j] = s.data[j], s.data[i] }

// SortKey implements Sortable.
func (s stringSlice) SortKey(i int) string { return s.data[i] }

// Sort is a convenience method.
func (s stringSlice) Sort(query string) []*Result { return Sort(s, query) }

// SortStrings is a convenience function for fuzzy-sorting a slice of strings.
func SortStrings(data []string, query string) []*Result {
	s := stringSlice{data}
	return s.Sort(query)
}

// Match scores str against query using fuzzy matching and the specified sort options.
// WARNING: Match creates a new Sorter for every call. Don't use this on
// large datasets.
func Match(str, query string, o *SortOptions) *Result {
	data := stringSlice{[]string{str}}
	s := NewSorter(data, o)
	return s.Sort(query)[0]
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: non-spacing mark
}

func stripDiacritics(s string) string {
	stripped, _, err := transform.String(stripper, s)
	if err != nil {
		log.Printf("Couldn't strip diacritics from `%s`: %s", s, err)
		return s
	}
	return stripped
}

func isASCII(s string) bool {
	return stripDiacritics(s) == s
}
