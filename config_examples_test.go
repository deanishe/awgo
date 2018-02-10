//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package aw

import (
	"fmt"
	"os"
	"time"
)

// Basic usage of Config.Get. Returns an empty string if variable is unset.
func ExampleConfig_Get() {
	// Set some test variables
	os.Setenv("TEST_NAME", "Bob Smith")
	os.Setenv("TEST_ADDRESS", "7, Dreary Lane")

	// New Config from environment
	c := NewConfig()

	fmt.Println(c.Get("TEST_NAME"))
	fmt.Println(c.Get("TEST_ADDRESS"))
	fmt.Println(c.Get("TEST_NONEXISTENT")) // unset variable

	// GetString is a synonym
	fmt.Println(c.GetString("TEST_NAME"))

	// Output:
	// Bob Smith
	// 7, Dreary Lane
	//
	// Bob Smith

	unsetEnv("TEST_NAME", "TEST_ADDRESS")
}

// The fallback value is returned if the variable is unset.
func ExampleConfig_Get_fallback() {
	// Set some test variables
	os.Setenv("TEST_NAME", "Bob Smith")
	os.Setenv("TEST_ADDRESS", "7, Dreary Lane")
	os.Setenv("TEST_EMAIL", "")

	// New Config from environment
	c := NewConfig()

	fmt.Println(c.Get("TEST_NAME", "default name"))       // fallback ignored
	fmt.Println(c.Get("TEST_ADDRESS", "default address")) // fallback ignored
	fmt.Println(c.Get("TEST_EMAIL", "test@example.com"))  // fallback ignored (var is empty, not unset)
	fmt.Println(c.Get("TEST_NONEXISTENT", "hi there!"))   // unset variable

	// Output:
	// Bob Smith
	// 7, Dreary Lane
	//
	// hi there!

	unsetEnv("TEST_NAME", "TEST_ADDRESS", "TEST_EMAIL")
}

// Getting int values with and without fallbacks.
func ExampleConfig_GetInt() {
	// Set some test variables
	os.Setenv("PORT", "3000")
	os.Setenv("PING_INTERVAL", "")

	// New Config from environment
	c := NewConfig()

	fmt.Println(c.GetInt("PORT"))
	fmt.Println(c.GetInt("PORT", 5000))        // fallback is ignored
	fmt.Println(c.GetInt("PING_INTERVAL"))     // returns zero value
	fmt.Println(c.GetInt("PING_INTERVAL", 60)) // returns fallback
	// Output:
	// 3000
	// 3000
	// 0
	// 60

	unsetEnv("PORT", "PING_INTERVAL")
}

// Strings are parsed to floats using strconv.ParseFloat().
func ExampleConfig_GetFloat() {
	// Set some test variables
	os.Setenv("TOTAL_SCORE", "172.3")
	os.Setenv("AVERAGE_SCORE", "7.54")

	// New Config from environment
	c := NewConfig()

	fmt.Printf("%0.2f\n", c.GetFloat("TOTAL_SCORE"))
	fmt.Printf("%0.1f\n", c.GetFloat("AVERAGE_SCORE"))
	fmt.Println(c.GetFloat("NON_EXISTENT_SCORE", 120.5))
	// Output:
	// 172.30
	// 7.5
	// 120.5

	unsetEnv("TOTAL_SCORE", "AVERAGE_SCORE")
}

// Durations are parsed using time.ParseDuration.
func ExampleConfig_GetDuration() {
	// Set some test variables
	os.Setenv("DURATION_NAP", "20m")
	os.Setenv("DURATION_EGG", "5m")
	os.Setenv("DURATION_BIG_EGG", "")
	os.Setenv("DURATION_MATCH", "1.5h")

	// New Config from environment
	c := NewConfig()

	// returns time.Duration
	fmt.Println(c.GetDuration("DURATION_NAP"))
	fmt.Println(c.GetDuration("DURATION_EGG") * 2)
	// fallback with unset variable
	fmt.Println(c.GetDuration("DURATION_POWERNAP", time.Minute*45))
	// or an empty one
	fmt.Println(c.GetDuration("DURATION_BIG_EGG", time.Minute*10))
	fmt.Println(c.GetDuration("DURATION_MATCH").Minutes())

	// Output:
	// 20m0s
	// 10m0s
	// 45m0s
	// 10m0s
	// 90

	unsetEnv(
		"DURATION_NAP",
		"DURATION_EGG",
		"DURATION_BIG_EGG",
		"DURATION_MATCH",
	)
}

// Strings are parsed using strconv.ParseBool().
func ExampleConfig_GetBool() {

	// Set some test variables
	os.Setenv("LIKE_PEAS", "t")
	os.Setenv("LIKE_CARROTS", "true")
	os.Setenv("LIKE_BEANS", "1")
	os.Setenv("LIKE_LIVER", "f")
	os.Setenv("LIKE_TOMATOES", "0")
	os.Setenv("LIKE_BVB", "false")
	os.Setenv("LIKE_BAYERN", "FALSE")

	// New Config from environment
	c := NewConfig()

	// strconv.ParseBool() supports many formats
	fmt.Println(c.GetBool("LIKE_PEAS"))
	fmt.Println(c.GetBool("LIKE_CARROTS"))
	fmt.Println(c.GetBool("LIKE_BEANS"))
	fmt.Println(c.GetBool("LIKE_LIVER"))
	fmt.Println(c.GetBool("LIKE_TOMATOES"))
	fmt.Println(c.GetBool("LIKE_BVB"))
	fmt.Println(c.GetBool("LIKE_BAYERN"))

	// Fallback
	fmt.Println(c.GetBool("LIKE_BEER", true))

	// Output:
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// true

	unsetEnv(
		"LIKE_PEAS",
		"LIKE_CARROTS",
		"LIKE_BEANS",
		"LIKE_LIVER",
		"LIKE_TOMATOES",
		"LIKE_BVB",
		"LIKE_BAYERN",
	)
}

func unsetEnv(keys ...string) {
	for _, key := range keys {
		os.Unsetenv(key)
	}
}
