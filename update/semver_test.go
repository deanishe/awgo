// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"reflect"
	"testing"
)

func TestVersionValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, x string
		valid bool
	}{
		// invalid versions
		{"", "", false},
		{"bob", "", false},
		{"1.x.8", "", false},
		{"1.0b", "", false},
		{"1.0.3a", "", false},
		{"1.0.0.0", "", false},
		{"01", "", false},
		{"01.2.3", "", false},
		{"blah.2.3", "", false},
		{"1.blah.3", "", false},
		{"1.2.blah", "", false},
		// valid versions
		{"1", "1.0.0", true},
		{"1.0.0", "1.0.0", true},
		{"1.9", "1.9.0", true},
		{"1.9.0", "1.9.0", true},
		{"10.0", "10.0.0", true},
		{"10.0.0", "10.0.0", true},
		{"1.0.1", "1.0.1", true},
		{"2.2.1", "2.2.1", true},
		{"10.11.12", "10.11.12", true},
		{"9.99.9999", "9.99.9999", true},
		{"12.333.0-alpha", "12.333.0-alpha", true},
		{"8.10.11", "8.10.11", true},
		{"9.4.3+20144353453", "9.4.3+20144353453", true},
		{"3.1.4-beta+20144334", "3.1.4-beta+20144334", true},
		{"1.1-beta", "1.1.0-beta", true},
		{"1-beta", "1.0.0-beta", true},
		{"5.1-beta+20170915", "5.1.0-beta+20170915", true},
		{"1.01", "1.1.0", true},
		{"0.0.1", "0.0.1", true},
		// prefixed version strings
		{"v1", "1.0.0", true},
		{"v5.2.1-beta", "5.2.1-beta", true},
		{"v2.01.02-alpha+759", "2.1.2-alpha+759", true},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("SemVer(%#v)", td.in), func(t *testing.T) {
			t.Parallel()
			v, err := NewSemVer(td.in)
			if err != nil {
				if td.valid {
					t.Error("parse valid failed")
				}
			} else {
				if v.String() != td.x {
					t.Errorf("Expected=%s, Got=%s", td.x, v)
				}
			}
		})
	}
}

// TestVersionCompare compares versions strings
func TestVersionCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
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
		{"1.1.0", "1.1.1", -1},
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

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("%q vs %q", td.v1, td.v2), func(t *testing.T) {
			t.Parallel()
			v1, err1 := NewSemVer(td.v1)
			v2, err2 := NewSemVer(td.v2)
			if err1 != nil || err2 != nil {
				t.Fatalf("Different errors. v1=%s, v2=%s", err1, err2)
			}
			r := v1.Compare(v2)
			if r != td.r {
				t.Fatalf("Expected=%d, Got=%d", td.r, r)
			}
			if td.r == 0 {
				if !v1.Eq(v2) {
					t.Fatal("[EQ] Did not compare as equal")
				}
				if !v1.Gte(v2) {
					t.Fatal("[GTE] Did not compare as equal")
				}
				if !v1.Lte(v2) {
					t.Fatal("[LTE] Did not compare as equal")
				}
			} else if td.r == 1 {
				if v1.Eq(v2) {
					t.Fatal("[EQ] Compared as equal")
				}
				if !v1.Gte(v2) {
					t.Fatal("[GTE] Did not compare as greater")
				}
				if v1.Lte(v2) {
					t.Fatal("[LTE] Compared as LTE")
				}
			} else if td.r == -1 {
				if v1.Eq(v2) {
					t.Fatal("[EQ] Compared as equal")
				}
				if v1.Gte(v2) {
					t.Fatal("[GTE] Compared as GTE")
				}
				if !v1.Lte(v2) {
					t.Fatal("[LTE] Did not compare as LTE")
				}
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		v    string
		zero bool
	}{
		{"", true},
		{"0", true},
		{"0.0", true},
		{"0.0.0", true},
		// invalid strings also return zero SemVer
		{"one", true},
		{"1.two.3", true},

		{"1.0", false},
		{"1.0.2", false},
	}

	for _, td := range tests {
		v, _ := NewSemVer(td.v)
		if v.IsZero() != td.zero {
			t.Errorf("Bad IsZero for %q. Expected=%v, Got=%v", td.v, td.zero, v.IsZero())
		}
	}
}
func TestVersionSorting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  []string
		out []string
	}{
		{
			[]string{"5", "10", "1"},
			[]string{"1.0.0", "5.0.0", "10.0.0"},
		},
		{
			[]string{"v1", "2", "1.0.0-beta"},
			[]string{"1.0.0-beta", "1.0.0", "2.0.0"},
		},
	}

	for _, td := range tests {

		td := td // capture variable
		t.Run(fmt.Sprintf("%#v", td.in), func(t *testing.T) {
			t.Parallel()
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
			if !reflect.DeepEqual(out, td.out) {
				t.Errorf("Expected=%#v, Got=%#v", td.out, out)
			}
		})
	}
}
