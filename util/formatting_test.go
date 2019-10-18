// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in string
		x  string
	}{
		{"", ""},
		{" ", "-"},
		{"Test", "Test"},
		{"Test Space", "Test-Space"},
		{"Test  Multiple  Spaces", "Test-Multiple-Spaces"},
		{" Trim Space ", "-Trim-Space-"},
		{"  Trim Spaces  ", "-Trim-Spaces-"},
		{"Dots.Are.OK", "Dots.Are.OK"},
		{"ÄSCÏI ònly", "ASCII-only"},
		{"Filesystem/safe", "Filesystem-safe"},
		{"Filesystem: safe", "Filesystem-safe"},
	}

	for _, td := range tests {
		td := td
		t.Run(td.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, Slugify(td.in), "unexpected slug")
		})
	}
}

type padTest struct {
	str string // input string
	pad string // pad character
	n   int    // size to pad to
	x   string // expected result
}

// TestPadLeft tests PadLeft
func TestPadLeft(t *testing.T) {
	t.Parallel()

	var padLeftTests = []padTest{
		// Simple cases
		{"wow", "-", 5, "--wow"},
		{"pow", " ", 4, " pow"},
		// Input same length as n
		{"pow", " ", 3, "pow"},
		// Input longer than n
		{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padLeftTests {
		td := td
		t.Run(td.str, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, PadLeft(td.str, td.pad, td.n), "unexpected output")
		})
	}
}

// TestPadRight tests PadRight
func TestPadRight(t *testing.T) {
	t.Parallel()

	var padRightTests = []padTest{
		// Simple cases
		{"wow", "-", 5, "wow--"},
		{"pow", " ", 4, "pow "},
		// Input same length as n
		{"pow", " ", 3, "pow"},
		// Input longer than n
		{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padRightTests {
		td := td
		t.Run(td.str, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, PadRight(td.str, td.pad, td.n), "unexpected output")
		})
	}
}

// TestPad tests Pad
func TestPad(t *testing.T) {
	t.Parallel()

	var padTests = []padTest{
		// Simple cases
		{"wow", "-", 5, "-wow-"},
		{"pow", " ", 4, "pow "},
		// Even-length str
		{"wow", "-", 10, "---wow----"},
		// Input same length as n
		{"pow", " ", 3, "pow"},
		// Input longer than n
		{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padTests {
		td := td
		t.Run(td.str, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, Pad(td.str, td.pad, td.n), "unexpected output")
		})
	}
}
