// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// MustExist creates all specified directories and returns the last one.
// Panics if any directory cannot be created.
// All created directories have permission set to 0700.
func MustExist(dirpath ...string) string {
	var path string
	for _, path = range dirpath {
		if err := os.MkdirAll(path, 0700); err != nil {
			panic(fmt.Sprintf("create directory %q: %v", path, err))
		}
	}
	return path
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

// ClearDirectory deletes all files within a directory, but not directory itself.
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

// WriteFile is an atomic version of ioutil.WriteFile.
// It first writes data to a temporary file and renames this to
// filename if the write is successful.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dir, name := filepath.Split(filename)
	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return err
	}
	defer closeOrPanic(f)

	name = f.Name()
	defer func() {
		// Ensure tempfile is deleted
		if err := os.Remove(name); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("[ERROR] tempfile: %v", err)
			}
		}
	}()

	if err := ioutil.WriteFile(name, data, perm); err != nil {
		return err
	}

	return os.Rename(name, filename)
}

func closeOrPanic(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}
