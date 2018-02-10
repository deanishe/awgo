//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

// Package util contains general helper functions for workflow (library)
// authors.
package util

import (
	"fmt"
	"log"
	"os"
)

// MustExist takes and returns a directory path, creating the directory
// if necessary. Any created directories have permission set to 700.
// Panics if the directory cannot be created.
func MustExist(dirpath string) string {
	err := os.MkdirAll(dirpath, 0700)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create directory `%s` : %v", dirpath, err))
	}
	return dirpath
}

// PathExists checks for the existence of path.
// Panics if an error is encountered.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

// ClearDirectory deletes all files within a directory.
func ClearDirectory(p string) error {
	if !PathExists(p) {
		return nil
	}
	err := os.RemoveAll(p)
	MustExist(p)
	if err == nil {
		log.Printf("deleted contents of `%s`", p)
	}
	return err
}
