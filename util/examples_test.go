// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Shorten paths by replacing user's home directory with ~
func ExamplePrettyPath() {

	paths := []string{
		"",
		"$HOME",
		"$HOME/",
		"$HOME/Documents",
		"/Applications",
	}

	for _, s := range paths {
		// Expand $HOME
		p := os.ExpandEnv(s)

		fmt.Println(PrettyPath(p))
	}
	// Output:
	//
	// ~
	// ~/
	// ~/Documents
	// /Applications
}

func ExamplePadLeft() {
	fmt.Println(PadLeft("wow", "-", 5))
	// Output: --wow
}

func ExamplePadRight() {
	fmt.Println(PadRight("wow", "-", 5))
	// Output: wow--
}

func ExamplePad() {
	fmt.Println(Pad("wow", "-", 10))
	// Output: ---wow----
}

func ExamplePathExists() {

	name := "my-test-file.txt"

	// Non-existent file
	fmt.Println(PathExists(name))

	// Create the file
	if err := ioutil.WriteFile(name, []byte("test"), 0600); err != nil {
		panic(err)
	}

	// Now it exists
	fmt.Println(PathExists(name))
	// Output:
	// false
	// true

	if err := os.Remove(name); err != nil {
		panic(err)
	}
}

// QuoteAS wraps the string in quotes and escapes quotes within the string.
func ExampleQuoteAS() {
	values := []string{
		"",
		"simple",
		"with spaces",
		`has "quotes" within`,
		`"within quotes"`,
		`"`,
	}

	// Quote values for insertion into AppleScript
	for _, s := range values {
		fmt.Println(QuoteAS(s))
	}
	// Output:
	// ""
	// "simple"
	// "with spaces"
	// "has " & quote & "quotes" & quote & " within"
	// quote & "within quotes" & quote
	// quote
}

func ExampleRunAS() {

	// Some test words
	data := []string{
		"Hello, AppleScript!",
		`"Just Do It!"`,
		`He said, "I'm fine!" then died :(`,
		`"`,
	}

	for _, input := range data {

		// Simple script to return input
		// QuoteAS adds quotation marks, so don't add any more
		quoted := QuoteAS(input)
		script := "return " + quoted

		// Run script and collect result
		output, err := RunAS(script)
		if err != nil {
			// handle error
		}

		fmt.Printf("> %s\n", input)
		fmt.Printf("< %s\n", output)
	}

	// Output:
	// > Hello, AppleScript!
	// < Hello, AppleScript!
	// > "Just Do It!"
	// < "Just Do It!"
	// > He said, "I'm fine!" then died :(
	// < He said, "I'm fine!" then died :(
	// > "
	// < "

}

// You can pass additional arguments to your scripts.
func ExampleRunJS_arguments() {

	// Some test values
	argv := []string{"angular", "react", "vue"}

	script := `function run(argv) { return argv.join('\n') }`
	output, err := RunJS(script, argv...)
	if err != nil {
		// handle error
	}

	fmt.Println(output)

	// Output:
	// angular
	// react
	// vue
}

// Run calls any executable file. It does *not* use $PATH to find commands.
func ExampleRun() {

	// Create a simple test script
	filename := "test-script"
	script := `#!/bin/bash
	echo -n Happy Hour
	`

	// Make sure script is executable!
	if err := ioutil.WriteFile(filename, []byte(script), 0700); err != nil {
		panic(err)
	}

	// Note: we're running "test-script", but Run looks for "./test-script",
	// not a command "test-script" on your $PATH.
	out, err := Run(filename)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	// Output:
	// Happy Hour

	if err := os.Remove(filename); err != nil {
		panic(err)
	}
}

// You can pass arguments to the program/script you run.
func ExampleRun_arguments() {

	// Run an executable with arguments
	out, err := Run("/bin/bash", "-c", "echo -n Stringfellow Hawke")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	// Output:
	// Stringfellow Hawke
}

// Run recognises certain kinds of script files and knows which
// interpreter to run them with.
func ExampleRun_scripts() {

	// Test scripts that output $1.
	// Run will run them based on their file extension.
	scripts := []struct {
		name, code string
	}{
		{"test-file.py", "import sys; print(sys.argv[1])"},
		{"test-file.txt", "ignored"}, // invalid
		{"test-file.sh", `echo "$1"`},
		{"test-file.scpt", "on run(argv)\nreturn first item of argv\nend run"},
		{"test-file.doc", "irrelevant"}, // invalid
	}

	// Create test scripts. Note: they aren't executable.
	for _, script := range scripts {
		if err := ioutil.WriteFile(script.name, []byte(script.code), 0600); err != nil {
			panic(err)
		}
	}

	// Run scripts
	for _, script := range scripts {

		// Run runs file based on file extension
		// Pass script's own name as $1
		data, err := Run(script.name, script.name)
		if err != nil {

			// We're expecting 2 unknown types
			if err == ErrUnknownFileType {
				fmt.Printf("[err] %s: %s\n", err, script.name)
				continue
			}

			// Oops :(
			panic(err)
		}

		// Script's own name
		str := strings.TrimSpace(string(data))
		fmt.Println(str)
	}

	// Output:
	// test-file.py
	// [err] unknown filetype: test-file.txt
	// test-file.sh
	// test-file.scpt
	// [err] unknown filetype: test-file.doc

	for _, script := range scripts {
		if err := os.Remove(script.name); err != nil {
			panic(err)
		}
	}
}
