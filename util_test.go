//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-05-23
//

package workflow

import (
	"fmt"
	"testing"
)

type padTest struct {
	str      string
	pad      string
	n        int
	expected string
}

var padLeftTests = []padTest{
	// Simple cases
	padTest{"wow", "-", 5, "--wow"},
	padTest{"pow", " ", 4, " pow"},
	// Input same length as n
	padTest{"pow", " ", 3, "pow"},
	// Input longer than n
	padTest{"powwow", " ", 3, "powwow"},
}

var padRightTests = []padTest{
	// Simple cases
	padTest{"wow", "-", 5, "wow--"},
	padTest{"pow", " ", 4, "pow "},
	// Input same length as n
	padTest{"pow", " ", 3, "pow"},
	// Input longer than n
	padTest{"powwow", " ", 3, "powwow"},
}

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

// TestPadLeft tests PadLeft
func TestPadLeft(t *testing.T) {
	for _, td := range padLeftTests {
		if out := PadLeft(td.str, td.pad, td.n); out != td.expected {
			t.Fatalf("PadLeft output incorrect. Expected=%v, Got=%v", td.expected, out)
		}
	}
}

// TestPadRight tests PadRight
func TestPadRight(t *testing.T) {
	for _, td := range padRightTests {
		if out := PadRight(td.str, td.pad, td.n); out != td.expected {
			t.Fatalf("PadRight output incorrect. Expected=%v, Got=%v", td.expected, out)
		}
	}
}

// TestPad tests Pad
func TestPad(t *testing.T) {
	for _, td := range padTests {
		if out := Pad(td.str, td.pad, td.n); out != td.expected {
			t.Fatalf("Pad output incorrect. Expected=%v, Got=%v", td.expected, out)
		}
	}
}

func ExamplePadLeft() {
	fmt.Println(PadLeft("wow", "-", 5))
	// Output: --wow
}

func ExamplePadRight() {
	fmt.Println(PadRight("wow", "-", 5))
	// Output: wow--
}

func ExamplePad() {
	fmt.Println(Pad("wow", "-", 10))
	// Output: ---wow----
}
