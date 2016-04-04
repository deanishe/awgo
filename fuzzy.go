package workflow

import (
	"sort"
	"strings"
	"unicode"
)

// Fuzzy makes a slice fuzzy-sortable.
// The standard sort.Interface (i.e. Less()) is used as a fallback for
// data with the same score.
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
	if kw == fq.Query {
		return 1.0
	}
	if kwLC == fq.Query {
		return 0.99
	}
	if s := fq.scoreCapitals(kw); s > 0.0 {
		return s
	}
	if s := fq.scoreInitials(kw); s > 0.0 {
		return s
	}
	if s := fq.scorePrefix(kwLC); s > 0.0 {
		return s
	}
	if s := fq.scoreContains(kwLC); s > 0.0 {
		return s
	}
	if s := fq.scoreContainsAll(kwLC); s > 0.0 {
		return s
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
		return 0.98
	} else if strings.HasPrefix(str, q) {
		// TODO: Alter score based on relative length of match.
		return 0.95
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
		return 0.99
	} else if strings.HasPrefix(str, q) {
		return 0.94
	}
	return 0.0
}

// scorePrefix | Whether kw starts with fq.Query.
func (fq *fuzzyQuery) scorePrefix(kw string) float64 {
	q := strings.ToLower(fq.Query)
	if strings.HasPrefix(kw, q) {
		return 0.9
	}
	return 0.0
}

// scoreContains | Whether kw contains fq.Query.
func (fq *fuzzyQuery) scoreContains(kw string) float64 {
	q := strings.ToLower(fq.Query)
	if strings.Contains(kw, q) {
		return 0.7
	}
	return 0.0
}

// scoreContainsAll | Whether kw contains all characters in fq.Query in order.
func (fq *fuzzyQuery) scoreContainsAll(kw string) float64 {
	var i int
	q := strings.ToLower(fq.Query)
	for _, c := range q {
		i = strings.Index(kw, string(c))
		if i < 0 {
			return 0.0
		}
		kw = kw[i+1:]
	}
	return 0.3
}
