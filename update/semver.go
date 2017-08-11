//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-01
//

package update

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	letters  string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	numbers         = "0123456789"
	alphanum        = letters + numbers
)

// SemVers implements sort.Interface for SemVer.
type SemVers []SemVer

// Len implements sort.Interface
func (vs SemVers) Len() int { return len(vs) }

// Less implements sort.Interface
func (vs SemVers) Less(i, j int) bool { return vs[i].LT(vs[j]) }

// Swap implements sort.Interface
func (vs SemVers) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }

// SortSemVer sorts a slice of SemVer structs.
func SortSemVer(versions []SemVer) {
	sort.Sort(SemVers(versions))
}

// SemVer is a (mostly) semantic version number.
//
// Unlike the semver standard:
//	- Minor and patch versions are not required, e.g. "v1" and "v1.0" are valid.
//	  NOTE: You may not specify pre-release data unless major, minor and patch
//	  numbers are given, i.e. "1-beta" and "1.0-beta" are invalid, "1.0.0-beta"
//	  is valid.
//	- Version string may be prefixed with "v", e.g. "v1" or "v3.0.1-beta".
//	  The "v" prefix is stripped, so "v1" == "1.0.0".
//	- Dots and integers are ignored in pre-release identifiers: they are
//	  compared purely alphanumerically, e.g. "v1-beta.11 < "v1-beta.2".
//	  Use "v1-beta.02" instead.
type SemVer struct {
	Major      uint64 // Increment for breaking changes.
	Minor      uint64 // Increment for added/deprecated functionality.
	Patch      uint64 // Increment for bugfixes.
	Build      string // Build metadata (ignored in comparisons)
	Prerelease string // Pre-release version (treated as string)
}

// NewSemVer creates a new SemVer. An error is returned if the version
// string is not valid. See the SemVer struct documentation for deviations
// from the semver standard.
func NewSemVer(s string) (SemVer, error) {
	var major, minor, patch uint64
	var build, pre string
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return SemVer{}, fmt.Errorf("Empty version string: %q", s)
	}
	// Remove "v" prefix and extend short versions to full length.
	s = strings.TrimPrefix(s, "v")

	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 3 {
		if strings.ContainsAny(parts[len(parts)-1], "+-") {
			return SemVer{}, errors.New("Short versions may not contain pre-release or build data.")
		}
	}
	for len(parts) < 3 { // Pad version
		parts = append(parts, "0")
	}

	if len(parts) != 3 {
		return SemVer{}, fmt.Errorf("%d part(s), not 3: %q", len(parts), s)
	}

	// Major
	if !containsOnly(parts[0], numbers) {
		return SemVer{}, fmt.Errorf("Invalid char(s) in major number %q", parts[0])
	}
	if hasLeadingZeroes(parts[0]) {
		return SemVer{}, fmt.Errorf("Major version may not contain leading zeroes: %q", parts[0])
	}
	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("Invalid major version %q: %s", parts[0], err)
	}

	// Minor
	if !containsOnly(parts[1], numbers) {
		return SemVer{}, fmt.Errorf("Invalid char(s) in minor number %q", parts[1])
	}
	minor, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("Invalid minor version %q: %s", parts[1], err)
	}

	// Patch
	pat := parts[2]
	if i := strings.IndexRune(pat, '+'); i != -1 {
		pat, build = pat[:i], pat[i+1:]
	}
	if i := strings.IndexRune(pat, '-'); i != -1 {
		pat, pre = pat[:i], pat[i+1:]
	}

	if !containsOnly(pat, numbers) {
		return SemVer{}, fmt.Errorf("Invalid char(s) in minor number %q", pat)
	}
	patch, err = strconv.ParseUint(pat, 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("Invalid minor version %q: %s", pat, err)
	}

	v := SemVer{}
	v.Major = major
	v.Minor = minor
	v.Patch = patch
	v.Prerelease = pre
	v.Build = build
	return v, nil
}

// String returns a canonical semver string
func (v SemVer) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s = fmt.Sprintf("%s-%s", s, v.Prerelease)
	}
	if v.Build != "" {
		s = fmt.Sprintf("%s+%s", s, v.Build)
	}
	return s
}

// Compare compares two Versions. Returns:
//	-1 if v < v2
//	 0 if v == v2
//	 1 if v > v2
func (v SemVer) Compare(v2 SemVer) int {
	if v.Major != v2.Major {
		if v.Major > v2.Major {
			return 1
		}
		return -1
	}
	if v.Minor != v2.Minor {
		if v.Minor > v2.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != v2.Patch {
		if v.Patch > v2.Patch {
			return 1
		}
		return -1
	}

	// Check if one version is prerelease and the other isn't
	if v.Prerelease == "" && v2.Prerelease != "" {
		return 1
	} else if v.Prerelease != "" && v2.Prerelease == "" {
		return -1
	}

	if v.Prerelease > v2.Prerelease {
		return 1
	} else if v.Prerelease < v2.Prerelease {
		return -1
	}

	// Semver ignores build info
	return 0
}

// Equals checks if v == v2
func (v SemVer) Equals(v2 SemVer) bool {
	return v.Compare(v2) == 0
}

// EQ checks if v == v2
func (v SemVer) EQ(v2 SemVer) bool { return v.Equals(v2) }

// NE checks if v != v2
func (v SemVer) NE(v2 SemVer) bool { return !v.EQ(v2) }

// GT checks if v > v2
func (v SemVer) GT(v2 SemVer) bool { return v.Compare(v2) == 1 }

// GTE checks if v >= v2
func (v SemVer) GTE(v2 SemVer) bool { return v.Compare(v2) >= 0 }

// LT checks if v < v2
func (v SemVer) LT(v2 SemVer) bool { return v.Compare(v2) == -1 }

// LTE checks if v <= v2
func (v SemVer) LTE(v2 SemVer) bool { return v.Compare(v2) <= 0 }

func containsOnly(s, allowed string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(allowed, r)
	}) == -1
}

func hasLeadingZeroes(s string) bool {
	return s[0] == '0' && len(s) > 1
}
