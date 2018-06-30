//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-01
//

package update

import "testing"

var invalidV = []string{
	"",
	"bob",
	"1.x.8",
	"1.0b",
	"1.0.3a",
	"1.0.0.0",
	"01",
	"01.2.3",
}
var validV = []string{
	"1",
	"1.0.0",
	"1.9",
	"1.9.0",
	"10.0",
	"10.0.0",
	"1.0.1",
	"2.2.1",
	"10.11.12",
	"9.99.9999",
	"12.333.0-alpha",
	"8.10.11",
	"9.4.3+20144353453",
	"3.1.4-beta+20144334",
	"1.1-beta",
	"1-beta",
	"5.1-beta+20170915",
}

var canonicalV = []struct {
	in  string
	out string
}{
	{"v1", "1.0.0"},
	{"1.01", "1.1.0"},
	{"0.0.1", "0.0.1"},
	{"v5.2.1-beta", "5.2.1-beta"},
	{"v2.01.02-alpha+759", "2.1.2-alpha+759"},
}

var comparisonV = []struct {
	v1 string
	v2 string
	r  int
}{
	{"1", "1.0", 0},
	{"1", "1.0.0", 0},
	{"v1", "1.0", 0},
	{"v1", "1.0.0", 0},
	{"1", "v1.0", 0},
	{"1", "v1.0.0", 0},
	{"v2", "1.0", 1},
	{"1", "1.0", 0},
	{"1.1.0", "1.0", 1},
	{"1.1.0", "1.2", -1},
	{"1.1.0-alpha", "1.1.0", -1},
	{"1.1.0-beta", "1.1.0-alpha", 1},
	{"1.1.0-alpha", "1.1.0-alpha", 0},
	{"1.1.0-rc1", "1.1.0-rc2", -1},
	{"10.1.0", "1.1.0", 1},
	{"0.4.5", "0.5.0-beta", -1},
	// Build metadata ignored
	{"1.1.0-rc1+749", "1.1.0-rc1+750", 0},
	{"1.1.0+10", "1.1.0+11", 0},
	{"1.1.0+12", "1.1.0+11", 0},
}

var sortedV = []struct {
	in  []string
	out []string
}{
	{[]string{"5", "10", "1"},
		[]string{"1.0.0", "5.0.0", "10.0.0"}},
	{[]string{"v1", "2", "1.0.0-beta"},
		[]string{"1.0.0-beta", "1.0.0", "2.0.0"}},
}

// TestInvalidV tests invalid version strings
func TestInvalidV(t *testing.T) {
	for _, s := range invalidV {
		v, err := NewSemVer(s)
		t.Logf("v=%v, err=%v", v, err)
		if err == nil {
			t.Fatalf("Bad version accepted: %s", s)
		}
	}
}

// TestValidV tests valid version strings
func TestValidV(t *testing.T) {
	for _, s := range validV {
		v, err := NewSemVer(s)
		t.Logf("v=%v, err=%v", v, err)
		if err != nil {
			t.Fatalf("Good version failed: %q -> %s", s, err)
		}
		// Check with "v" prefix
		s2 := "v" + s
		v, err = NewSemVer(s)
		t.Logf("v=%v, err=%v", v, err)
		if err != nil {
			t.Fatalf("Good version failed: %q -> %s", s2, err)
		}
	}
}

// TestCanonicalV tests canonical forms of version strings
func TestCanonicalV(t *testing.T) {
	for _, td := range canonicalV {
		v, err := NewSemVer(td.in)
		if err != nil {
			t.Fatalf("Canonical error: %q -> %s", td.in, err)
		}
		s := v.String()
		if s != td.out {
			t.Fatalf("Bad canonical: %q -> Expected %q Got %q", td.in, td.out, s)
		}
	}
}

// TestComparisonV compares versions strings
func TestComparisonV(t *testing.T) {
	for _, td := range comparisonV {
		v1, err1 := NewSemVer(td.v1)
		v2, err2 := NewSemVer(td.v2)
		if err1 != nil || err2 != nil {
			t.Fatalf("Version error(s). v1=%s, v2=%s", err1, err2)
		}
		r := v1.Compare(v2)
		if r != td.r {
			t.Fatalf("Failed comparison %q vs %q. Expected=%d, Got=%d", v1, v1, td.r, r)
		}
		if td.r == 0 {
			if !v1.Eq(v2) {
				t.Fatalf("[EQ] Did not compare as equal: %q %q", v1, v2)
			}
			if !v1.Gte(v2) {
				t.Fatalf("[GTE] Did not compare as equal: %q %q", v1, v2)
			}
			if !v1.Lte(v2) {
				t.Fatalf("[LTE] Did not compare as equal: %q %q", v1, v2)
			}
		} else if td.r == 1 {
			if v1.Eq(v2) {
				t.Fatalf("[EQ] Compared as equal: %q %q", v1, v2)
			}
			if !v1.Gte(v2) {
				t.Fatalf("[GTE] Did not compare as greater: %q %q", v1, v2)
			}
			if v1.Lte(v2) {
				t.Fatalf("[LTE] Compared as LTE: %q %q", v1, v2)
			}
		} else if td.r == -1 {
			if v1.Eq(v2) {
				t.Fatalf("[EQ] Compared as equal: %q %q", v1, v2)
			}
			if v1.Gte(v2) {
				t.Fatalf("[GTE] Compared as GTE: %q %q", v1, v2)
			}
			if !v1.Lte(v2) {
				t.Fatalf("[LTE] Did not compare as LTE: %q %q", v1, v2)
			}
		}
	}
}

func TestSortedV(t *testing.T) {
	for _, td := range sortedV {
		vin := []SemVer{}
		out := []string{}
		for _, s := range td.in {
			v, _ := NewSemVer(s)
			vin = append(vin, v)
		}
		SortSemVer(vin)
		for _, v := range vin {
			out = append(out, v.String())
		}
		if len(out) != len(td.out) {
			t.Fatalf("Bad length. Expected=%d, Got=%d", len(td.out), len(out))
		}
		for i, s := range td.out {
			if s != out[i] {
				t.Fatalf("Bad sort. Expected=%q, Got=%q", out[i], s)
			}
		}
	}
}
