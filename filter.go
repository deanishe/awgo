package workflow

type FilterItem struct {
	SearchText string
	Payload    interface{}
	Score      float64
}

type FilterItems []FilterItem

func (s FilterItems) Len() int {
	return len(s)
}

func (s FilterItems) Less(i, j int) bool {
	return s[i].Score < s[i].Score
}

func (s FilterItems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
