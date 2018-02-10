//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package util

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// RunScriptAS executes AppleScript.
func RunScriptAS(script string) ([]byte, error) {

	cmd := exec.Command("/usr/bin/osascript", "-l", "AppleScript", "-e", script)

	return RunCmd(cmd)
}

// RunCmd executes a command and returns its output.
func RunCmd(cmd *exec.Cmd) ([]byte, error) {

	var (
		output         []byte
		stdout, stderr bytes.Buffer
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("------------- %v ---------------", cmd.Args)
		log.Println(stderr.String())
		log.Println("----------------------------------------------")
		return nil, err
	}

	output = stdout.Bytes()

	return output, nil
}

// AppleScriptify escapes a string for insertion into quotes in AppleScript.
func AppleScriptify(s string) string {
	return strings.Replace(s, `"`, `" & quote & "`, -1)
}

// QuoteAS quotes a string for insertion into AppleScript code.
func QuoteAS(s string) string {

	if s == "" {
		return `""`
	}

	if s == `"` {
		return "quote"
	}

	chars := []string{}
	for i, c := range s {
		if c == '"' {
			if i == 0 {
				chars = append(chars, `quote & "`)
			} else if i == len(s)-1 {
				chars = append(chars, `" & quote`)
			} else {
				chars = append(chars, `" & quote & "`)
			}
			continue
		}
		if i == 0 {
			chars = append(chars, `"`)
		}
		chars = append(chars, string(c))
		if i == len(s)-1 {
			chars = append(chars, `"`)
		}
	}

	return strings.Join(chars, "")
}
