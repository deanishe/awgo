//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package fuzzy

// Option configures a Sorter. Pass one or more Options to New() or
// Sorter.Configure(). An Option returns another Option to revert the
// configuration to the previous state.
type Option func(s *Sorter) Option

// AdjacencyBonus sets the bonus for adjacent matches.
func AdjacencyBonus(bonus float64) Option {
	return func(s *Sorter) Option {
		prev := s.AdjacencyBonus
		s.AdjacencyBonus = bonus
		return AdjacencyBonus(prev)
	}
}

// SeparatorBonus sets the bonus for matches directly after a separator.
func SeparatorBonus(bonus float64) Option {
	return func(s *Sorter) Option {
		prev := s.SeparatorBonus
		s.SeparatorBonus = bonus
		return SeparatorBonus(prev)
	}
}

// CamelBonus sets the bonus applied if match is uppercase and previous character is lowercase.
func CamelBonus(bonus float64) Option {
	return func(s *Sorter) Option {
		prev := s.CamelBonus
		s.CamelBonus = bonus
		return CamelBonus(prev)
	}
}

// LeadingLetterPenalty sets the penalty applied for every character before the first match.
func LeadingLetterPenalty(penalty float64) Option {
	return func(s *Sorter) Option {
		prev := s.LeadingLetterPenalty
		s.LeadingLetterPenalty = penalty
		return LeadingLetterPenalty(prev)
	}
}

// MaxLeadingLetterPenalty sets the maximum penalty for character preceding the first match.
func MaxLeadingLetterPenalty(penalty float64) Option {
	return func(s *Sorter) Option {
		prev := s.MaxLeadingLetterPenalty
		s.MaxLeadingLetterPenalty = penalty
		return MaxLeadingLetterPenalty(prev)
	}
}

// UnmatchedLetterPenalty sets the penalty for characters that do not match.
func UnmatchedLetterPenalty(penalty float64) Option {
	return func(s *Sorter) Option {
		prev := s.UnmatchedLetterPenalty
		s.UnmatchedLetterPenalty = penalty
		return UnmatchedLetterPenalty(prev)
	}
}

// StripDiacritics sets whether diacritics should be stripped.
func StripDiacritics(on bool) Option {
	return func(s *Sorter) Option {
		prev := s.StripDiacritics
		s.StripDiacritics = on
		return StripDiacritics(prev)
	}
}
