// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package fuzzy

import "testing"

// TestSortStrings tests that strings are sorted correctly.
func TestSortStrings(t *testing.T) {

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
		data := []string{td.s}
		r := SortStrings(data, td.q)
		if r[0].Match != td.m {
			t.Errorf("query=%#v, str=%#v => %v, expected=%v", td.q, td.s, r[0].Match, td.m)
		}
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

// TestStripDiacritics
func TestStripDiacritics(t *testing.T) {

	// Non-ASCII query and data
	if r := Match("fün", "fün"); r.Match == false {
		t.Fatalf("fün != fün (diacritic stripping on): %+v", r)
	}
	// Non-ASCII data
	if r := Match("fün", "fun"); r.Match == false {
		t.Fatalf("fun != fün (diacritic stripping on): %+v", r)
	}
	// No diacritic stripping
	if r := Match("fün", "fün", StripDiacritics(false)); r.Match == false {
		t.Fatalf("fün != fün (diacritic stripping off): %+v", r)
	}
	if r := Match("fün", "fun", StripDiacritics(false)); r.Match == true {
		t.Fatalf("fun != fün (diacritic stripping off): %+v", r)
	}
}
