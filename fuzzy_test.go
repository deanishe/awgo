//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-10-30
//

package aw

import "testing"

var simpleHostnames = []string{
	"www.example.com",
	"one.example.com",
	"two.example.com",
	"www.google.com",
	"www.amazon.de",
	// Contains "two"
	"www.two.co.uk",
}

var firstTestData = []struct {
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

var rankTestData = []struct {
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

var matchNoMatchData = []struct {
	q string
	s string
	m bool
}{
	{"ruto", "Router", false},
	{"ruto", "ruTorrent", true},
	{"GoT", "Game of Thrones", true},
	{"GoT", "Game of Phones", false},
}

var feedbackTitles = []struct {
	q   string
	in  []string
	out []string
	m   []bool
}{
	{
		q:   "got",
		in:  []string{"game of thrones", "no match", "got milk?", "got"},
		out: []string{"got", "game of thrones", "got milk?", "no match"},
		m:   []bool{true, true, true, false},
	},
	{
		q:   "of",
		in:  []string{"out of time", "spelunking", "OmniFocus", "game of thrones"},
		out: []string{"OmniFocus", "out of time", "game of thrones", "spelunking"},
		m:   []bool{true, true, true, false},
	},
	{
		q:   "safa",
		in:  []string{"see all fellows' armpits", "Safari", "french canada", "spanish harlem"},
		out: []string{"Safari", "see all fellows' armpits", "spanish harlem", "french canada"},
		m:   []bool{true, true, false, false},
	},
}

var filterTitles = []struct {
	q   string
	in  []string
	out []string
}{
	{
		q:   "got",
		in:  []string{"game of thrones", "no match", "got milk?", "got"},
		out: []string{"got", "game of thrones", "got milk?"},
	},
	{
		q:   "of",
		in:  []string{"out of time", "spelunking", "OmniFocus", "game of thrones"},
		out: []string{"OmniFocus", "out of time", "game of thrones"},
	},
	{
		q:   "safa",
		in:  []string{"see all fellows' armpits", "Safari", "french canada", "spanish harlem"},
		out: []string{"Safari", "see all fellows' armpits"},
	},
}

// TestSortStrings tests that strings are sorted correctly.
func TestSortStrings(t *testing.T) {
	for _, td := range rankTestData {
		// t.Logf("query=%#v, in=%#v, expected=%#v", td.q, td.in, td.out)
		data := td.in[:]
		SortStrings(data, td.q)
		for i := 0; i < len(data); i++ {
			if data[i] != td.out[i] {
				t.Errorf("query=%#v, in=%#v, expected=%#v, actual=%#v", td.q, td.in, td.out, data)
			}
		}
	}
}

// TestMatchNoMatch tests queries and strings for match status.
func TestMatchNoMatch(t *testing.T) {
	for _, td := range matchNoMatchData {
		data := []string{td.s}
		r := SortStrings(data, td.q)
		if r[0].Match != td.m {
			t.Errorf("query=%#v, str=%#v => %v, expected=%v", td.q, td.s, r[0].Match, td.m)
		}
	}
}

// TestFirstMatch tests the expected matching result is first.
func TestFirstMatch(t *testing.T) {
	for _, td := range firstTestData {
		data := td.in[:]
		r := SortStrings(data, td.q)
		for i, s := range data {
			if r[i].Match {
				if s != td.first {
					t.Errorf("query=%#v => %#v, expected=%#v", td.q, s, td.first)
				}
				break
			}
		}

	}
}

// TestSortFeedback sorts Feedback.Items
func TestSortFeedback(t *testing.T) {
	o := NewSortOptions()
	for _, td := range feedbackTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		r := fb.Sort(td.q, o)
		for i, it := range fb.Items {
			if it.title != td.out[i] {
				t.Errorf("query=%#v, pos=%d, expected=%s, got=%s", td.q, i+1, td.out[i], it.title)
			}
			if r[i].Match != td.m[i] {
				t.Errorf("query=%#v, keywords=%#v, expected=%v, got=%v", td.q, it.title, td.m[i], r[i].Match)
			}
		}
	}
}

// TestFilterFeedback filters Feedback.Items
func TestFilterFeedback(t *testing.T) {
	o := NewSortOptions()
	for _, td := range filterTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		fb.Filter(td.q, o)
		if len(fb.Items) != len(td.out) {
			t.Errorf("query=%#v, expected %d results, got %d", td.q, len(td.out), len(fb.Items))
		}
		for i, it := range fb.Items {
			if it.title != td.out[i] {
				t.Errorf("query=%#v, pos=%d, expected=%s, got=%s", td.q, i+1, td.out[i], it.title)
			}
		}
	}
}
