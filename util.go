//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FindWorkflowRoot returns the workflow root directory.
// Tries to find info.plist in or above current working directory
// and the executable's parent directory.
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
func FindFileUpwards(filename string, startdir string) (string, error) {
	dirpath, _ := filepath.Abs(startdir)
	for dirpath != "/" {
		p := path.Join(dirpath, filename)
		if PathExists(p) {
			// log.Printf("%v found at %v", filename, p)
			return p, nil
		}
		dirpath = path.Dir(dirpath)
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
