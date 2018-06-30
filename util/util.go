//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

/*

Package util contains general helper functions for workflow (library) authors.

The functions can be divided into roughly three groups: paths, formatting
and scripting.


Paths

There are a couple of convenience path functions, MustExist and
ClearDirectory.


Formatting

PrettyPath for user-friendly paths, and the Pad* functions for padding
strings.

Scripting

QuoteAS quotes strings for insertion into AppleScript code and there
are several Run* functions for executing script code and files.

	Run()     // run a script file or executable & return output
	RunAS()   // run AppleScript code & return output
	RunJS()   // run JXA code & return output
	RunCmd()  // run *exec.Cmd & return output

Run takes the path to a script or executable. If file is executable,
it runs the file directly. If it's a script file, it tries to guess the
appropriate interpreter.

See Runner for more information.

*/
package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Timed logs the duration since start & title. Use it with defer.
//
//    func doSomething() {
//        defer Timed(time.Now(), "long running task")
//        // do thing here
//        // and another thing
//    }
//    // Output: ... long running task
//
func Timed(start time.Time, title string) {
	log.Printf("%s \U000029D7 %s", time.Now().Sub(start), title)
}

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
