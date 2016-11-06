//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FindWorkflowRoot returns the workflow's root directory.
// Tries to find info.plist in or above current working directory
// and the executable's parent directory.
//
// TODO: Make function FindWorkflowRoot private.
func FindWorkflowRoot() (string, error) {
	candidateDirs := []string{}
	// Current working directory
	cwd, err := os.Getwd()
	if err == nil {
		cwd, _ = filepath.Abs(cwd)
		// log.Printf("cwd=%v", dir)
		candidateDirs = append(candidateDirs, cwd)
	}

	// Parent directory of running program
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil && dir != cwd {
		candidateDirs = append(candidateDirs, dir)
	}

	for _, dir := range candidateDirs {
		p, err := FindFileUpwards("info.plist", dir)
		if err == nil {
			dirpath, _ := filepath.Split(p)
			return dirpath, nil
		}
	}
	return "", fmt.Errorf("info.plist not found")
}

// EnsureExists takes and returns a directory path, creating the directory
// if necessary. Any created directories have permission set to 700.
func EnsureExists(dirpath string) string {
	err := os.MkdirAll(dirpath, 0700)
	if err != nil {
		panic(fmt.Errorf("Couldn't create directory `%s` : %v", dirpath, err))
	}
	return dirpath
}

// PathExists checks for the existence of path.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// FindFileUpwards searches for a file named filename. It first looks in startdir,
// then its parent directory and so on until it reaches /
//
// TODO: Make function FindFileUpwards private.
func FindFileUpwards(filename string, startdir string) (string, error) {
	dirpath, _ := filepath.Abs(startdir)
	for dirpath != "/" {
		p := filepath.Join(dirpath, filename)
		if PathExists(p) {
			// log.Printf("%v found at %v", filename, p)
			return p, nil
		}
		dirpath = filepath.Dir(dirpath)
	}
	err := fmt.Errorf("File %v not found in or above %v", filename, startdir)
	return "", err
}

// ShortenPath replaces $HOME with ~ in path
func ShortenPath(path string) string {
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

// SensibleDuration returns a sensibly-formatted string for
// non-benchmarking purposes.
func SensibleDuration(d time.Duration) string {
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
		return fmt.Sprintf("%0.2fs", d.Seconds())
	}
	if d.Seconds() >= 0.1 {
		return fmt.Sprintf("%0.3fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Nanoseconds()/1000000)
}

// clearDirectory deletes all files within a directory.
func clearDirectory(p string) error {
	if !PathExists(p) {
		return nil
	}
	err := os.RemoveAll(p)
	EnsureExists(p)
	if err == nil {
		log.Printf("Delete contents of `%s`", p)
	}
	return err
}
