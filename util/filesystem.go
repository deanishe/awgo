package util

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
		p, err := FindFile("info.plist", dir)
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

// Exists checks for the existence of path.
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// FindFile searches for a file named filename. It first looks in startdir,
// then its parent directory and so on until it reaches /
func FindFile(filename string, startdir string) (string, error) {
	dirpath, _ := filepath.Abs(startdir)
	for dirpath != "/" {
		p := path.Join(dirpath, filename)
		if Exists(p) {
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
