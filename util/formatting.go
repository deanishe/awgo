// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	rxAlphaNum  = regexp.MustCompile(`[^a-zA-Z0-9.-]+`)
	rxMultiDash = regexp.MustCompile(`-+`)
)

// Slugify makes a string filesystem- and URL-safe.
func Slugify(s string) string {
	s = fold(s)
	s = rxAlphaNum.ReplaceAllString(s, "-")
	s = rxMultiDash.ReplaceAllString(s, "-")
	return s
}

// isMn returns true if rune is a non-spacing mark
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: non-spacing mark
}

// fold strips diacritics from string.
func fold(s string) string {
	stripper := transform.Chain(norm.NFD, transform.RemoveFunc(isMn))
	ascii, _, err := transform.String(stripper, s)
	if err != nil {
		panic(err)
	}
	return ascii
}

// PrettyPath replaces $HOME with ~ in path
func PrettyPath(path string) string {
	return strings.Replace(path, os.Getenv("HOME"), "~", -1)
}

// PadLeft pads str to length n by adding pad to its left.
func PadLeft(str, pad string, n int) string {
	if len(str) >= n {
		return str
	}
	for {
		str = pad + str
		if len(str) >= n {
			return str[len(str)-n:]
		}
	}
}

// PadRight pads str to length n by adding pad to its right.
func PadRight(str, pad string, n int) string {
	if len(str) >= n {
		return str
	}
	for {
		str = str + pad
		if len(str) >= n {
			return str[len(str)-n:]
		}
	}
}

// Pad pads str to length n by adding pad to both ends.
func Pad(str, pad string, n int) string {
	if len(str) >= n {
		return str
	}
	for {
		str = pad + str + pad
		if len(str) >= n {
			return str[len(str)-n:]
		}
	}
}
