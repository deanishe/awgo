// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
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

// fold strips diacritics from string.
func fold(s string) string {
	stripper := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)))
	ascii, _, err := transform.String(stripper, s)
	if err != nil {
		panic(err)
	}
	return ascii
}

// PrettyPath replaces $HOME with ~ in path
func PrettyPath(path string) string {
	return strings.ReplaceAll(path, os.Getenv("HOME"), "~")
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
		str += pad
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
