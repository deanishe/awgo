// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package fuzzy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSortStrings tests that strings are sorted correctly.
func TestSortStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q   string
		in  []string
		out []string
	}{
		{
			q:   "got",
			in:  []string{"go and throw", "baby got back", "game of thrones"},
			out: []string{"game of thrones", "go and throw", "baby got back"},
		},
		{
			q:   "ruto",
			in:  []string{"Router", "Wolf // ruTorrent"},
			out: []string{"Wolf // ruTorrent", "Router"},
		},
	}

	for _, td := range tests {
		td := td
		t.Run(td.q, func(t *testing.T) {
			SortStrings(td.in, td.q)
			assert.Equal(t, td.out, td.in, "unexpected sort results")
		})
	}
}

// TestMatchNoMatch tests queries and strings for match status.
func TestMatchNoMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q string
		s string
		m bool
	}{
		{"ruto", "Router", false},
		{"ruto", "ruTorrent", true},
		{"GoT", "Game of Thrones", true},
		{"GoT", "Game of Phones", false},
	}

	for _, td := range tests {
		td := td
		t.Run(td.s, func(t *testing.T) {
			data := []string{td.s}
			r := SortStrings(data, td.q)
			assert.Equal(t, td.m, r[0].Match, "unexpected match")
		})
	}
}

// TestFirstMatch tests the expected matching result is first.
func TestFirstMatch(t *testing.T) {
	simpleHostnames := []string{
		"www.example.com",
		"one.example.com",
		"two.example.com",
		"www.google.com",
		"www.amazon.de",
		// Contains "two"
		"www.two.co.uk",
	}

	tests := []struct {
		q     string
		in    []string
		first string
	}{
		{"one", simpleHostnames, "one.example.com"},
		{"two", simpleHostnames, "two.example.com"},
		{"oec", simpleHostnames, "one.example.com"},
		{"am", simpleHostnames, "www.amazon.de"},
		{"example", simpleHostnames, "one.example.com"},
		{"wex", simpleHostnames, "www.example.com"},
		{"tuk", simpleHostnames, "www.two.co.uk"},
	}

	for _, td := range tests {
		td := td
		t.Run(td.q, func(t *testing.T) {
			r := SortStrings(td.in, td.q)
			for i, s := range td.in {
				if r[i].Match {
					assert.Equal(t, td.first, s, "unexpected first result")
					break
				}
			}
		})
	}
}

// TestStripDiacritics
func TestStripDiacritics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s, q     string
		strip, x bool
	}{
		// non-ASCII query and data
		{"fün", "fün", true, true},
		// non-ASCII data
		{"fün", "fun", true, true},
		// no stripping
		{"fün", "fün", false, true},
		{"fün", "fun", false, false},
	}

	for _, td := range tests {
		td := td
		t.Run(fmt.Sprintf("%q=%q", td.q, td.s), func(t *testing.T) {
			assert.Equal(t, td.x, Match(td.s, td.q, StripDiacritics(td.strip)).Match, "unexpected match")
		})
	}
}
