// Copyright (c) 2014 gofuzz by nbjahan - https://github.com/nbjahan/gofuzz

// Based on https://github.com/sergi/go-diff
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fuzzy

import (
	"math"
	"strings"
)

// Matcher is a bitap searcher
type Matcher struct {
	// How far to search for a match (0 = exact location, 1000+ = broad match).
	// A match this many characters away from the expected location will add
	// 1.0 to the score (0.0 is a perfect match).
	Distance int
	// The number of bits in an int.
	MaxBits int
	// At what point is no match declared (0.0 = perfection, 1.0 = very loose).
	Threshold float64

	CaseSensitive bool
}

// NewMatcher creates a new Matcher object with default parameters.
func NewMatcher() *Matcher {
	// Defaults.
	return &Matcher{
		Threshold:     0.5,
		Distance:      1000,
		MaxBits:       32,
		CaseSensitive: false,
	}
}

// Search locates the best instance of 'pattern' in 'text' near 'loc'.
// Returns pos, score
// Returns -1 if no match found.
func (m *Matcher) Search(text, pattern string, loc int) (int, float64) {
	if !m.CaseSensitive {
		text = strings.ToLower(text)
		pattern = strings.ToLower(pattern)
	}
	loc = int(math.Max(0, math.Min(float64(loc), float64(len(text)))))
	if text == pattern {
		// Shortcut (potentially not guaranteed by the algorithm)
		return 0, 0
	} else if len(text) == 0 {
		// Nothing to match.
		return -1, 1
	} else if loc+len(pattern) <= len(text) && text[loc:loc+len(pattern)] == pattern {
		// Perfect match at the perfect spot!  (Includes case of nil pattern)
		return loc, 0
	}
	// Do a fuzzy compare.
	return m.searchBitap(text, pattern, loc)
}

// MatchBitap locates the best instance of 'pattern' in 'text' near 'loc' using the
// Bitap algorithm.  Returns -1 if no match found.
func (m *Matcher) searchBitap(text, pattern string, loc int) (pos int, score float64) {
	// Initialise the alphabet.
	s := m.alphabet(pattern)

	// Highest score beyond which we give up.
	var scoreThreshold = m.Threshold
	// Is there a nearby exact match? (speedup)
	bestLoc := strings.Index(text, pattern)
	if bestLoc != -1 {
		scoreThreshold = math.Min(m.bitapScore(0, bestLoc, loc,
			pattern), scoreThreshold)
		// What about in the other direction? (speedup)
		bestLoc = strings.LastIndex(text, pattern)
		if bestLoc != -1 {
			scoreThreshold = math.Min(m.bitapScore(0, bestLoc, loc,
				pattern), scoreThreshold)
		}
	}

	// Initialise the bit arrays.
	matchmask := 1 << uint((len(pattern) - 1))
	bestLoc = -1

	var binMin, binMid int
	binMax := len(pattern) + len(text)
	lastRD := []int{}
	for d := 0; d < len(pattern); d++ {
		// Scan for the best match; each iteration allows for one more error.
		// Run a binary search to determine how far from 'loc' we can stray at
		// this error level.
		binMin = 0
		binMid = binMax
		for binMin < binMid {
			if m.bitapScore(d, loc+binMid, loc, pattern) <= scoreThreshold {
				binMin = binMid
			} else {
				binMax = binMid
			}
			binMid = (binMax-binMin)/2 + binMin
		}
		// Use the result from this iteration as the maximum for the next.
		binMax = binMid
		start := int(math.Max(1, float64(loc-binMid+1)))
		finish := int(math.Min(float64(loc+binMid), float64(len(text))) + float64(len(pattern)))

		rd := make([]int, finish+2)
		rd[finish+1] = (1 << uint(d)) - 1

		for j := finish; j >= start; j-- {
			var charMatch int
			if len(text) <= j-1 {
				// Out of range.
				charMatch = 0
			} else if _, ok := s[text[j-1]]; !ok {
				charMatch = 0
			} else {
				charMatch = s[text[j-1]]
			}

			if d == 0 {
				// First pass: exact match.
				rd[j] = ((rd[j+1] << 1) | 1) & charMatch
			} else {
				// Subsequent passes: fuzzy match.
				rd[j] = ((rd[j+1]<<1)|1)&charMatch | (((lastRD[j+1] | lastRD[j]) << 1) | 1) | lastRD[j+1]
			}
			if (rd[j] & matchmask) != 0 {
				score = m.bitapScore(d, j-1, loc, pattern)
				// This match will almost certainly be better than any existing
				// match.  But check anyway.
				if score <= scoreThreshold {
					// Told you so.
					scoreThreshold = score
					bestLoc = j - 1
					if bestLoc > loc {
						// When passing loc, don't exceed our current distance from loc.
						start = int(math.Max(1, float64(2*loc-bestLoc)))
					} else {
						// Already passed loc, downhill from here on in.
						break
					}
				}
			}
		}
		if m.bitapScore(d+1, loc, loc, pattern) > scoreThreshold {
			// No hope for a (better) match at greater error levels.
			break
		}
		lastRD = rd
	}
	if bestLoc == -1 {
		score = 1
	}
	return bestLoc, score
}

// bitapScore computes and returns the score for a match with e errors and x location.
func (m *Matcher) bitapScore(e, x, loc int, pattern string) float64 {
	var accuracy = float64(e) / float64(len(pattern))
	proximity := math.Abs(float64(loc - x))
	if m.Distance == 0 {
		// Dodge divide by zero error.
		if proximity == 0 {
			return accuracy
		}
		return 1.0
	}
	return accuracy + (proximity / float64(m.Distance))
}

// alphabet initialises the alphabet for the Bitap algorithm.
func (m *Matcher) alphabet(pattern string) map[byte]int {
	s := map[byte]int{}
	charPattern := []byte(pattern)
	for _, c := range charPattern {
		_, ok := s[c]
		if !ok {
			s[c] = 0
		}
	}
	i := 0

	for _, c := range charPattern {
		value := s[c] | int(uint(1)<<uint((len(pattern)-i-1)))
		s[c] = value
		i++
	}
	return s
}
