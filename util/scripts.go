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
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrUnknownFileType is returned by Run for files it can't identify.
var ErrUnknownFileType = errors.New("unknown filetype")

// Default Runners used by Run to determine how to execute a file.
var (
	Executable Runner // run executable files directly
	Script     Runner // run script files with commands from Interpreters

	// DefaultInterpreters maps script file extensions to interpreters.
	// Used by the Script Runner (and by extension Run()) to to
	// determine how to run files that aren't executable.
	DefaultInterpreters = map[string][]string{
		".py":          []string{"/usr/bin/python"},
		".rb":          []string{"/usr/bin/ruby"},
		".sh":          []string{"/bin/bash"},
		".zsh":         []string{"/bin/zsh"},
		".scpt":        []string{"/usr/bin/osascript"},
		".scptd":       []string{"/usr/bin/osascript"},
		".applescript": []string{"/usr/bin/osascript"},
		".js":          []string{"/usr/bin/osascript", "-l", "JavaScript"},
	}
)

// Available runners in order they should be tried.
// Executable and Script are added by init.
var runners Runners

func init() {

	// Default runners
	Executable = &ExecRunner{}
	Script = NewScriptRunner(DefaultInterpreters)

	runners = Runners{
		Executable,
		Script,
	}
}

// Runner knows how to execute a file passed to it.
// It is used by Run to determine how to run a file.
//
// When Run is passed a filepath, it asks each registered Runner
// in turn whether it can handle the file.
type Runner interface {
	// Can Runner execute this (type of) file?
	CanRun(filename string) bool
	// Cmd that executes file (via Runner's execution mechanism).
	Cmd(filename string, args ...string) *exec.Cmd
}

// Runners implements Runner over a sequence of Runner objects.
type Runners []Runner

// CanRun returns true if one of the runners can run this file.
func (rs Runners) CanRun(filename string) bool {

	for _, r := range rs {
		if r.CanRun(filename) {
			return true
		}
	}
	return false
}

// Cmd returns a command to run the (script) file.
func (rs Runners) Cmd(filename string, args ...string) *exec.Cmd {

	for _, r := range rs {
		if r.CanRun(filename) {
			return r.Cmd(filename, args...)
		}
	}

	return nil
}

// Run runs the executable or script at path and returns the output.
// If it can't figure out how to run the file (see Runner), it
// returns ErrUnknownFileType.
func (rs Runners) Run(filename string, args ...string) ([]byte, error) {

	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, ErrUnknownFileType
	}

	// See if a runner will accept file
	for _, r := range rs {

		if r.CanRun(filename) {

			cmd := r.Cmd(filename, args...)

			return RunCmd(cmd)
		}
	}

	return nil, ErrUnknownFileType
}

// Run runs the executable or script at path and returns the output.
// If it can't figure out how to run the file (see Runner), it
// returns ErrUnknownFileType.
func Run(filename string, args ...string) ([]byte, error) {
	return runners.Run(filename, args...)
}

// RunAS executes AppleScript and returns the output.
func RunAS(script string, args ...string) (string, error) {
	return runOsaScript(script, "AppleScript", args...)
}

// RunJS executes JavaScript (JXA) and returns the output.
func RunJS(script string, args ...string) (string, error) {
	return runOsaScript(script, "JavaScript", args...)
}

// runOsaScript executes a script with /usr/bin/osascript.
// It returns the output from STDOUT.
func runOsaScript(script, lang string, args ...string) (string, error) {

	argv := []string{"-l", lang, "-e", script}
	argv = append(argv, args...)

	cmd := exec.Command("/usr/bin/osascript", argv...)
	data, err := RunCmd(cmd)
	if err != nil {
		return "", err
	}

	s := string(data)

	// Remove trailing newline added by osascript
	if strings.HasSuffix(s, "\n") {
		s = s[0 : len(s)-1]
	}

	return s, nil
}

// RunCmd executes a command and returns its output.
//
// The main difference to exec.Cmd.Output() is that RunCmd writes all
// STDERR output to the log if a command fails.
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

// QuoteAS quotes a string for insertion into AppleScript code.
// It wraps the value in quotation marks, so don't insert additional ones.
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

// QuoteJS quotes a value for insertion into JavaScript.
// It calls json.Marshal(v), and returns an empty string if an error occurs.
func QuoteJS(v interface{}) string {

	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("couldn't convert %#v to JS: %v", v, err)
		return ""
	}

	return string(data)
}

// ExecRunner implements Runner for executable files.
type ExecRunner struct{}

// CanRun returns true if file exists and is executable.
func (r ExecRunner) CanRun(filename string) bool {

	fi, err := os.Stat(filename)
	if err != nil || fi.IsDir() {
		return false
	}

	perms := uint32(fi.Mode().Perm())
	return perms&0111 != 0
}

// Cmd returns a Cmd to run executable with args.
func (r ExecRunner) Cmd(executable string, args ...string) *exec.Cmd {

	executable, err := filepath.Abs(executable)
	if err != nil {
		panic(err)
	}

	return exec.Command(executable, args...)
}

// ScriptRunner implements Runner for the specified file extensions.
// It calls the given script with the interpreter command from Interpreters.
//
// A ScriptRunner (combined with Runners, which implements Run) is a useful
// base for adding support for running scripts to your own program.
type ScriptRunner struct {
	// Interpreters is an "extension: command" mapping of file extensions
	// to commands to invoke interpreters that can run the files.
	//
	//     Interpreters = map[string][]string{
	//         ".py": []string{"/usr/bin/python"},
	//         ".rb": []string{"/usr/bin/ruby"},
	//     }
	//
	Interpreters map[string][]string
}

// NewScriptRunner creates a new ScriptRunner for interpreters.
func NewScriptRunner(interpreters map[string][]string) *ScriptRunner {

	if interpreters == nil {
		interpreters = map[string][]string{}
	}

	r := &ScriptRunner{
		Interpreters: make(map[string][]string, len(interpreters)),
	}

	// Copy over defaults
	for k, v := range interpreters {
		r.Interpreters[k] = v
	}

	return r
}

// CanRun returns true if file exists and its extension is in Interpreters.
func (r ScriptRunner) CanRun(filename string) bool {

	if fi, err := os.Stat(filename); err != nil || fi.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))

	_, ok := r.Interpreters[ext]
	return ok
}

// Cmd returns a Cmd to run filename with its interpreter.
func (r ScriptRunner) Cmd(filename string, args ...string) *exec.Cmd {

	var (
		argv    []string
		command string
	)

	ext := strings.ToLower(filepath.Ext(filename))
	interpreter := DefaultInterpreters[ext]

	command = interpreter[0]

	argv = append(argv, interpreter[1:]...) // any remainder of interpreter command
	argv = append(argv, filename)           // path to script file
	argv = append(argv, args...)            // arguments to script

	return exec.Command(command, argv...)

}
