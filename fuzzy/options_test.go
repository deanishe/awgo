package fuzzy

import "testing"

// Test option round-tripping.
func TestOptions(t *testing.T) {
	var (
		adj        = 1.1
		sepBonus   = 1.2
		camBonus   = 1.3
		leadPen    = -1.4
		maxLeadPen = -15.0
		unmatchPen = -1.6
		strip      = true
	)

	s := &Sorter{}
	prev := s.Configure(AdjacencyBonus(adj))
	if s.AdjacencyBonus != adj {
		t.Errorf("Bad AdjacencyBonus. Expected=%v, Got=%v", adj, s.AdjacencyBonus)
	}
	s.Configure(prev)
	if s.AdjacencyBonus != 0.0 {
		t.Errorf("Bad AdjacencyBonus. Expected=%v, Got=%v", 0.0, s.AdjacencyBonus)
	}

	prev = s.Configure(SeparatorBonus(sepBonus))
	if s.SeparatorBonus != sepBonus {
		t.Errorf("Bad SeparatorBonus. Expected=%v, Got=%v", sepBonus, s.SeparatorBonus)
	}
	s.Configure(prev)
	if s.SeparatorBonus != 0.0 {
		t.Errorf("Bad SeparatorBonus. Expected=%v, Got=%v", 0.0, s.SeparatorBonus)
	}

	prev = s.Configure(CamelBonus(camBonus))
	if s.CamelBonus != camBonus {
		t.Errorf("Bad CamelBonus. Expected=%v, Got=%v", camBonus, s.CamelBonus)
	}
	s.Configure(prev)
	if s.CamelBonus != 0.0 {
		t.Errorf("Bad CamelBonus. Expected=%v, Got=%v", 0.0, s.CamelBonus)
	}

	prev = s.Configure(LeadingLetterPenalty(leadPen))
	if s.LeadingLetterPenalty != leadPen {
		t.Errorf("Bad LeadingLetterPenalty. Expected=%v, Got=%v", leadPen, s.LeadingLetterPenalty)
	}
	s.Configure(prev)
	if s.LeadingLetterPenalty != 0.0 {
		t.Errorf("Bad LeadingLetterPenalty. Expected=%v, Got=%v", 0.0, s.LeadingLetterPenalty)
	}

	prev = s.Configure(MaxLeadingLetterPenalty(maxLeadPen))
	if s.MaxLeadingLetterPenalty != maxLeadPen {
		t.Errorf("Bad MaxLeadingLetterPenalty. Expected=%v, Got=%v", maxLeadPen, s.MaxLeadingLetterPenalty)
	}
	s.Configure(prev)
	if s.MaxLeadingLetterPenalty != 0.0 {
		t.Errorf("Bad MaxLeadingLetterPenalty. Expected=%v, Got=%v", 0.0, s.MaxLeadingLetterPenalty)
	}

	prev = s.Configure(UnmatchedLetterPenalty(unmatchPen))
	if s.UnmatchedLetterPenalty != unmatchPen {
		t.Errorf("Bad UnmatchedLetterPenalty. Expected=%v, Got=%v", unmatchPen, s.UnmatchedLetterPenalty)
	}
	s.Configure(prev)
	if s.UnmatchedLetterPenalty != 0.0 {
		t.Errorf("Bad UnmatchedLetterPenalty. Expected=%v, Got=%v", 0.0, s.UnmatchedLetterPenalty)
	}

	prev = s.Configure(StripDiacritics(strip))
	if s.StripDiacritics != strip {
		t.Errorf("Bad StripDiacritics. Expected=%v, Got=%v", strip, s.StripDiacritics)
	}
	s.Configure(prev)
	if s.StripDiacritics != false {
		t.Errorf("Bad StripDiacritics. Expected=%v, Got=%v", false, s.StripDiacritics)
	}

}
