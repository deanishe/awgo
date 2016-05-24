//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package workflow

import (
	"sort"
	"strings"
	"unicode"
)

// Weightings for matching rules. Each match method returns
// a score between 0.0 and 100.0, which is multiplied by the
// corresponding weighting.
//
// Setting a weighting to 0.0 disables that match rule.
var (
	// Exact, case-sensitive match.
	WeightingExact = 1.0
	// Exact, case-insensitive match.
	WeightingExactCaseless = 0.98
	// Capital letters in keywords match query.
	WeightingCaps = 0.95
	// Initials of "words" in keywords match query.
	// Keywords are split on non-letter characters.
	WeightingInitials = 0.9
	// Keywords starts with query (case-insensitive).
	WeightingPrefix = 0.8
	// Query is a substring of keywords (case-insensitive).
	WeightingContains = 0.7
	// All characters in query appear in order in keywords (case-insensitive).
	WeightingOrderedSubset = 0.5
)

// Fuzzy makes a slice fuzzy-sortable.
// The standard sort.Interface (i.e. Less()) is used as a fallback for
// data with the same score.
//
// TODO: Simply to a single method plus an intermediate wrapper struct?
type Fuzzy interface {
	// Terms against which data will be sorted.
	Keywords(i int) string

	sort.Interface
}

// SortFuzzy fuzzy-sorts data based on query and returns the scores.
// Highest scores/best matches first.
//
// SortFuzzy *does not* truncate data. To only show matching results,
// you must check scores:
//
//    for i, score := range workflow.SortFuzzy(data, query) {
//        if score == 0.0 {
//            data = data[:i]
//            break
//        }
//    }
//
func SortFuzzy(data Fuzzy, query string) []float64 {
	fq := fuzzyQuery{query, data, nil}
	fq.initialize()
	sort.Sort(fq)
	// for i := 0; i < fq.Data.Len(); i++ {
	// 	log.Printf("%03d. %v", i, fq.Data.Keywords(i))
	// }
	return fq.scores
}

// fuzzyQuery fuzzy-sorts Data based on Query.
type fuzzyQuery struct {
	Query  string
	Data   Fuzzy
	scores []float64
}

// initialize fetches the keywords and calculates the scores for fq.Data.
func (fq *fuzzyQuery) initialize() {
	// log.Printf("Initialising fuzzy search for `%s` ...", fq.Query)
	fq.scores = make([]float64, fq.Data.Len())
	var score float64
	for i := 0; i < fq.Data.Len(); i++ {
		score = fq.CalculateScore(fq.Data.Keywords(i))
		fq.scores[i] = score
		// log.Printf("score=%f, keywords=%s", score, fq.Data.Keywords(i))
	}
}

// Len implements sort.Interface.
func (fq fuzzyQuery) Len() int {
	return fq.Data.Len()
}

// Swap implements sort.Interface.
func (fq fuzzyQuery) Swap(i, j int) {
	fq.scores[i], fq.scores[j] = fq.scores[j], fq.scores[i]
	fq.Data.Swap(i, j)
}

// Less implements sort.Interface. Comparison is based on fuzzy score
// and reversed, so higher scores (i.e. better matches) are first.
func (fq fuzzyQuery) Less(i, j int) bool {
	a, b := fq.scores[i], fq.scores[j]
	if a == b {
		// Normal comparison because A comes before B.
		return fq.Data.Less(i, j)
	}
	// Reverse comparison because higher comes before lower.
	return b < a
}

// CalculateScore rates kw against fq.Query.
func (fq *fuzzyQuery) CalculateScore(kw string) float64 {
	kwLC := strings.ToLower(kw)
	if WeightingExact > 0.0 && kw == fq.Query {
		return 100.0 * WeightingExact
	}
	if WeightingExactCaseless > 0.0 && kwLC == fq.Query {
		return 100.0 * WeightingExactCaseless
	}
	if WeightingCaps > 0.0 {
		if s := fq.scoreCapitals(kw); s > 0.0 {
			return s * WeightingCaps
		}
	}
	if WeightingInitials > 0.0 {
		if s := fq.scoreInitials(kw); s > 0.0 {
			return s * WeightingInitials
		}
	}
	if WeightingPrefix > 0.0 {
		if s := fq.scorePrefix(kwLC); s > 0.0 {
			return s * WeightingPrefix
		}
	}
	if WeightingContains > 0.0 {
		if s := fq.scoreContains(kwLC); s > 0.0 {
			return s * WeightingContains
		}
	}
	if WeightingOrderedSubset > 0.0 {
		if s := fq.scoreContainsAll(kwLC); s > 0.0 {
			return s * WeightingOrderedSubset
		}
	}
	return 0.0
}

// scoreCapitals | Whether fq.Query is a prefix of the combined capital
// letters in kw, such that, e.g. "of" matches "OmniFocus".
func (fq *fuzzyQuery) scoreCapitals(kw string) float64 {
	var caps []rune
	for _, c := range kw {
		if unicode.IsUpper(c) {
			caps = append(caps, unicode.ToLower(c))
		}
	}
	if len(caps) == 0 {
		return 0.0
	}
	str := string(caps)
	q := strings.ToLower(fq.Query)
	if strings.EqualFold(str, q) {
		return 100.0
	} else if strings.HasPrefix(str, q) {
		// TODO: Alter score based on relative length of match.
		return 100.0 - float64(len(str)-len(str))
		// j := float64(len(q)) / float64(len(str))
		// log.Printf("%v prefix of %v : %v", q, str, j)
		// return 0.95
	}
	return 0.0
}

// scoreInitals | Whether fq.Query matches first letters of words in kw.
// kw is split on non-word characters, and the initials are the first
// characters of each of those elements.
func (fq *fuzzyQuery) scoreInitials(kw string) float64 {
	var initials []rune
	var isLetter, wasLetter bool
	for i, c := range kw {
		isLetter = unicode.IsLetter(c)
		if i == 0 {
			if isLetter {
				initials = append(initials, c)
			}
		} else if isLetter && !wasLetter {
			initials = append(initials, c)
		}
		wasLetter = isLetter
	}
	str := strings.ToLower(string(initials))
	q := strings.ToLower(fq.Query)
	if strings.EqualFold(str, q) {
		return 100.0
	} else if strings.HasPrefix(str, q) {
		return 100.0 - float64(len(str)-len(q))
		// return 94.0
	}
	return 0.0
}

// scorePrefix | Whether kw starts with fq.Query.
func (fq *fuzzyQuery) scorePrefix(kw string) float64 {
	q := strings.ToLower(fq.Query)
	if strings.HasPrefix(kw, q) {
		return 100.0 - float64(len(kw)-len(q))
		// return 0.9
	}
	return 0.0
}

// scoreContains | Whether kw contains fq.Query.
func (fq *fuzzyQuery) scoreContains(kw string) float64 {
	q := strings.ToLower(fq.Query)
	if strings.Contains(kw, q) {
		i := len(kw) - len(q)
		j := strings.Index(kw, q)
		return float64(100 - i - j)
	}
	return 0.0
}

// scoreContainsAll | Whether kw contains all characters in fq.Query in order.
func (fq *fuzzyQuery) scoreContainsAll(kw string) float64 {
	var i, j int
	q := strings.ToLower(fq.Query)
	for _, c := range q {
		i = strings.Index(kw, string(c))
		if i < 0 {
			return 0.0
		}
		j += i
		kw = kw[i+1:]
	}
	return float64(100 - j)
}
