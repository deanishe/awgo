// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// SemVers implements sort.Interface for SemVer.
type SemVers []SemVer

// Len implements sort.Interface
func (vs SemVers) Len() int { return len(vs) }

// Less implements sort.Interface
func (vs SemVers) Less(i, j int) bool { return vs[i].Lt(vs[j]) }

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
//	- Version string may be prefixed with "v", e.g. "v1" or "v3.0.1-beta".
//	  The "v" prefix is stripped, so "v1" == "1.0.0".
//	- Dots and integers are ignored in pre-release identifiers: they are
//	  compared purely alphanumerically, e.g. "v1-beta.11" < "v1-beta.2".
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
		return SemVer{}, fmt.Errorf("empty version string: %q", s)
	}
	// Remove "v" prefix and extend short versions to full length.
	s = strings.TrimPrefix(s, "v")

	// Extract build and pre tags
	if i := strings.IndexRune(s, '+'); i != -1 {
		s, build = s[:i], s[i+1:]
	}
	if i := strings.IndexRune(s, '-'); i != -1 {
		s, pre = s[:i], s[i+1:]
	}

	parts := strings.SplitN(s, ".", -1)
	for len(parts) < 3 { // Pad version
		parts = append(parts, "0")
	}

	if len(parts) != 3 {
		return SemVer{}, fmt.Errorf("%d part(s), not 3: %q", len(parts), s)
	}

	// Major
	if hasLeadingZeroes(parts[0]) {
		return SemVer{}, fmt.Errorf("major version may not contain leading zeroes: %q", parts[0])
	}
	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid major version %q: %s", parts[0], err)
	}

	// Minor
	minor, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid minor version %q: %s", parts[1], err)
	}

	// Patch
	patch, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid patch version %q: %s", parts[2], err)
	}

	return SemVer{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: pre,
		Build:      build,
	}, nil
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

// Eq checks if v == v2
func (v SemVer) Eq(v2 SemVer) bool { return v.Compare(v2) == 0 }

// Ne checks if v != v2
func (v SemVer) Ne(v2 SemVer) bool { return !v.Eq(v2) }

// Gt checks if v > v2
func (v SemVer) Gt(v2 SemVer) bool { return v.Compare(v2) == 1 }

// Gte checks if v >= v2
func (v SemVer) Gte(v2 SemVer) bool { return v.Compare(v2) >= 0 }

// Lt checks if v < v2
func (v SemVer) Lt(v2 SemVer) bool { return v.Compare(v2) == -1 }

// Lte checks if v <= v2
func (v SemVer) Lte(v2 SemVer) bool { return v.Compare(v2) <= 0 }

// IsZero returns true if SemVer has no value.
func (v SemVer) IsZero() bool { return v.Eq(SemVer{}) }

func hasLeadingZeroes(s string) bool {
	return s[0] == '0' && len(s) > 1
}
