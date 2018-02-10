//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

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

// HumanDuration returns a sensibly-formatted string for non-benchmarking purposes.
func HumanDuration(d time.Duration) string {
	if d.Hours() >= 72 { // 3 days
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if d.Hours() >= 24 { // 1 day
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d.Minutes() > 90 {
		hrs := int(d.Hours())
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hrs, mins)
	}
	if d.Minutes() >= 10 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d.Seconds() > 90 {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	}
	if d.Seconds() >= 10 {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d.Seconds() > 1 {
		return fmt.Sprintf("%0.1fs", d.Seconds())
	}
	if d.Seconds() >= 0.1 {
		return fmt.Sprintf("%0.2fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Nanoseconds()/1000000)
}
