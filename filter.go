package workflow

import (
	"log"
	"sort"
	"strings"
)

type Filterable interface {
	Keywords() string
}

// Filter returns the subset of items that match query.
func Filter(query string, items []Filterable, minScore float64) []Filterable {
	rankers := make(Rankers, len(items))
	// results := make([]Filterable, len(items))
	var results []Filterable
	var r Ranker
	for i, it := range items {
		r = Ranker{it, 0.0}
		r.Rank(query)
		log.Printf("ranker=%v", r)
		rankers[i] = r
	}
	sort.Sort(sort.Reverse(rankers))
	for _, r := range rankers {
		log.Printf("%v %v", r, r.Item.Keywords())
		if r.Score >= minScore {
			results = append(results, r.Item)
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
func (r Ranker) Rank(query string) {
	// TODO: Rank results better
	kw := r.Item.Keywords()
	kwLC := strings.ToLower(kw)
	query = strings.ToLower(query)
	log.Printf("kw=%v, kwLC=%v, query=%v", kw, kwLC, query)
	if strings.HasPrefix(kwLC, query) {
		r.Score = 100.0
		log.Printf("%f %v", r.Score, kw)
	} else {
		r.Score = -1.0
		log.Printf("%f No Match : %v", r.Score, kw)
	}
}

// Rankers is a sortable slice of Ranker structs.
type Rankers []Ranker

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
