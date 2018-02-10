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
	"io/ioutil"
	"os"
	"time"
)

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

func ExampleReadableDuration() {
	fmt.Println(HumanDuration(time.Hour * 96))
	fmt.Println(HumanDuration(time.Hour * 48))
	fmt.Println(HumanDuration(time.Hour * 12))
	fmt.Println(HumanDuration(time.Minute * 130))
	fmt.Println(HumanDuration(time.Minute * 90))
	fmt.Println(HumanDuration(time.Second * 315))
	fmt.Println(HumanDuration(time.Second * 70))
	fmt.Println(HumanDuration(time.Second * 5))
	fmt.Println(HumanDuration(time.Millisecond * 320))
	fmt.Println(HumanDuration(time.Millisecond * 50))
	// Output: 4d
	// 48h
	// 12h0m
	// 2h10m
	// 90m
	// 5m15s
	// 70s
	// 5.0s
	// 0.32s
	// 50ms
}

func ExamplePathExists() {

	name := "my-test-file.txt"

	// Non-existent file
	fmt.Println(PathExists(name))

	// Create the file
	ioutil.WriteFile(name, []byte("test"), 0600)

	// Now it exists
	fmt.Println(PathExists(name))
	// Output:
	// false
	// true

	if err := os.Remove(name); err != nil {
		panic(err)
	}
}

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
