package workflow

import (
	"log"
	"sort"
	"strings"
	"unicode"
)

// Filterable interface denotes compatibility with function Filter.
type Filterable interface {
	Keywords() string
}

// Filter returns the subset of items that match query.
// The comparison is fuzzy.
func Filter(query string, items []Filterable, minScore float64) []Filterable {
	rankers := make(Rankers, len(items))
	// results := make([]Filterable, len(items))
	var results []Filterable
	var r *Ranker
	for i, it := range items {
		r = &Ranker{it, 0.0}
		r.Rank(query)
		// log.Printf("ranker=%v", r)
		rankers[i] = r
	}
	sort.Sort(sort.Reverse(rankers))
	i := 0
	for _, r := range rankers {
		// log.Printf("%v %v", r, r.Item.Keywords())
		if r.Score >= minScore {
			i += 1
			results = append(results, r.Item)
			log.Printf("%3d. (%f) %v", i, r.Score, r.Item)
		}
	}
	return results
}

// Ranker scores a Filterable based on its Keywords()
type Ranker struct {
	Item  Filterable
	Score float64
}

// Rank computes the score for Item.
func (r *Ranker) Rank(query string) {
	// TODO: Rank results better. Return normalised scores per rank
	// function and combine them with weightings.
	kw := r.Item.Keywords()
	kwLC := strings.ToLower(kw)
	query = strings.ToLower(query)
	// log.Printf("kw=%v, kwLC=%v, query=%v", kw, kwLC, query)
	// TODO: Score on contains word
	if strings.EqualFold(kw, query) {
		r.Score = 1000.0
	} else if score := r.rankCapitals(kw, query); score > 0.0 {
		r.Score = score
	} else if score := r.rankInitials(kw, query); score > 0.0 {
		r.Score = score
	} else if score := r.rankHasPrefix(kwLC, query); score > 0.0 {
		r.Score = score
		// } else if strings.Contains(kwLC, query) {
	} else if score := r.rankContains(kwLC, query); score > 0.0 {
		r.Score = score
	} else if score := r.rankContainsAll(kwLC, query); score > 0.0 {
		r.Score = score
	} else { // No match
		r.Score = -1.0
	}
	// log.Printf("[%f] %v", r.Score, kw)
}

// rankCapitals | Whether query is a prefix of the combined capital
// letters in kw, such that, e.g. "of" matches "OmniFocus".
func (r *Ranker) rankCapitals(kw, query string) float64 {
	var caps []rune
	// var caps string
	// caps := make([]rune, len(query))
	for _, c := range kw {
		if unicode.IsUpper(c) {
			caps = append(caps, unicode.ToLower(c))
			// caps = fmt.Sprintf("%s%s", caps, c)
		}
	}
	if len(caps) == 0 {
		return 0.0
	}
	test := string(caps)
	// log.Printf("[CAPS] %v  ->  %v", kw, test)
	if strings.EqualFold(query, test) {
		return 200.0
	} else if strings.HasPrefix(test, query) {
		return 190.0
	}
	return 0.0
}

// rankInitials | Whether query matches first letters of words in kw.
func (r *Ranker) rankInitials(kw, query string) float64 {
	var initials []rune
	var isLetter, wasJustLetter bool
	for i, c := range kw {
		isLetter = unicode.IsLetter(c)
		if i == 0 {
			if isLetter {
				initials = append(initials, c)
			}
		} else if isLetter && !wasJustLetter {
			initials = append(initials, c)
		}
		wasJustLetter = isLetter
	}
	test := string(initials)
	// log.Printf("[INITS] %s  ->  %s", kw, test)
	if strings.EqualFold(query, test) {
		return 180.0
	} else if strings.HasPrefix(test, query) {
		return 170.0
	}
	return 0.0
}

// rankHasPrefix | Whether kw starts with query.
func (r *Ranker) rankHasPrefix(kw, query string) float64 {
	if strings.HasPrefix(kw, query) {
		return 100.0
	}
	return 0.0
}

// rankContains | Whether kw contains query.
func (r *Ranker) rankContains(kw, query string) float64 {
	if strings.Contains(kw, query) {
		return 70.0
	}
	return 0.0
}

// rankContainsAll | Whether all characters in query appear in order in kw.
func (r *Ranker) rankContainsAll(kw, query string) float64 {
	var s string
	var i int
	s = kw
	for _, c := range query {
		i = strings.Index(s, string(c))
		if i < 0 {
			return 0.0
		}
		s = s[i+1:]
	}
	return 20.0
}

// Rankers is a sortable slice of Ranker structs.
type Rankers []*Ranker

// Len is a sort.Interface method.
func (slice Rankers) Len() int {
	return len(slice)
}

// Less is a sort.Interface method.
func (slice Rankers) Less(i, j int) bool {
	return slice[i].Score < slice[j].Score
}

// Swap is a sort.Interface method.
func (slice Rankers) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// func (s FilterItems) Len() int {
// 	return len(s)
// }

// func (s FilterItems) Less(i, j int) bool {
// 	return s[i].Score < s[i].Score
// }

// func (s FilterItems) Swap(i, j int) {
// 	s[i], s[j] = s[j], s[i]
// }
