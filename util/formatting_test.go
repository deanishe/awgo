// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"testing"
)

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
		padTest{"wow", "-", 5, "--wow"},
		padTest{"pow", " ", 4, " pow"},
		// Input same length as n
		padTest{"pow", " ", 3, "pow"},
		// Input longer than n
		padTest{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padLeftTests {
		if out := PadLeft(td.str, td.pad, td.n); out != td.x {
			t.Fatalf("PadLeft output incorrect. Expected=%v, Got=%v", td.x, out)
		}
	}
}

// TestPadRight tests PadRight
func TestPadRight(t *testing.T) {
	t.Parallel()

	var padRightTests = []padTest{
		// Simple cases
		padTest{"wow", "-", 5, "wow--"},
		padTest{"pow", " ", 4, "pow "},
		// Input same length as n
		padTest{"pow", " ", 3, "pow"},
		// Input longer than n
		padTest{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padRightTests {
		if out := PadRight(td.str, td.pad, td.n); out != td.x {
			t.Fatalf("PadRight output incorrect. Expected=%v, Got=%v", td.x, out)
		}
	}
}

// TestPad tests Pad
func TestPad(t *testing.T) {
	t.Parallel()

	var padTests = []padTest{
		// Simple cases
		padTest{"wow", "-", 5, "-wow-"},
		padTest{"pow", " ", 4, "pow "},
		// Even-length str
		padTest{"wow", "-", 10, "---wow----"},
		// Input same length as n
		padTest{"pow", " ", 3, "pow"},
		// Input longer than n
		padTest{"powwow", " ", 3, "powwow"},
	}
	for _, td := range padTests {
		if out := Pad(td.str, td.pad, td.n); out != td.x {
			t.Fatalf("Pad output incorrect. Expected=%v, Got=%v", td.x, out)
		}
	}
}
