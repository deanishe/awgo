// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestConfigEnv verifies that Config holds the expected values.
func TestConfigEnv(t *testing.T) {

	data := []struct {
		name, x, key string
	}{
		{"Version", tVersion, EnvVarVersion},
		{"Name", tName, EnvVarName},
		{"BundleID", tBundleID, EnvVarBundleID},
		{"UID", tUID, EnvVarUID},
		{"ConfigVersion", tAlfredVersion, EnvVarAlfredVersion},
		{"ConfigBuild", tAlfredBuild, EnvVarAlfredBuild},
		{"Theme", tTheme, EnvVarTheme},
		{"ThemeBackground", tThemeBackground, EnvVarThemeBG},
		{"ThemeSelectionBackground", tThemeSelectionBackground, EnvVarThemeSelectionBG},
		{"Preferences", tPreferences, EnvVarPreferences},
		{"Localhash", tLocalhash, EnvVarLocalhash},
		{"CacheDir", tCacheDir, EnvVarCacheDir},
		{"CacheDir", tDataDir, EnvVarDataDir},
	}

	cfg := NewConfig(testEnv)

	v := cfg.GetBool(EnvVarDebug)
	if v != tDebug {
		t.Errorf("bad Debug. Expected=%v, Got=%v", tDebug, v)
	}

	for _, td := range data {
		s := cfg.Get(td.key)
		if s != td.x {
			t.Errorf("Bad %s. Expected=%v, Got=%v", td.name, td.x, s)
		}
	}
}

func TestGet(t *testing.T) {
	env := MapEnv{
		"key":   "value",
		"key2":  "value2",
		"empty": "",
	}

	data := []struct {
		key string
		fb  []string
		out string
	}{
		// valid
		{"key", []string{}, "value"},
		{"key", []string{"value2"}, "value"},
		{"key2", []string{}, "value2"},
		{"key2", []string{"value"}, "value2"},
		// empty
		{"empty", []string{}, ""},
		{"empty", []string{"dave"}, ""},
		// unset
		{"key3", []string{}, ""},
		{"key3", []string{"bob"}, "bob"},
	}

	cfg := NewConfig(env)

	// Verify env is the same
	for k, x := range env {
		v := cfg.Get(k)
		if v != x {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", k, x, v)
		}
	}

	// Test Get
	for _, td := range data {
		v := cfg.Get(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetInt(t *testing.T) {
	env := MapEnv{
		"one":   "1",
		"two":   "2",
		"zero":  "0",
		"float": "3.5",
		"word":  "henry",
		"empty": "",
	}

	data := []struct {
		key string
		fb  []int
		out int
	}{
		// numbers
		{"one", []int{}, 1},
		{"two", []int{1}, 2},
		{"zero", []int{}, 0},
		{"zero", []int{2}, 0},
		// empty values
		{"empty", []int{}, 0},
		{"empty", []int{5}, 5},
		// non-existent values
		{"five", []int{}, 0},
		{"five", []int{5}, 5},
		// invalid values
		{"word", []int{}, 0},
		{"word", []int{5}, 5},
		// floats
		{"float", []int{}, 3},
		{"float", []int{5}, 3},
	}

	cfg := NewConfig(env)
	// Test GetInt
	for _, td := range data {
		v := cfg.GetInt(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetFloat(t *testing.T) {
	env := MapEnv{
		"one.three": "1.3",
		"two":       "2.0",
		"zero":      "0",
		"empty":     "",
		"word":      "henry",
	}

	data := []struct {
		key string
		fb  []float64
		out float64
	}{
		// numbers
		{"one.three", []float64{}, 1.3},
		{"two", []float64{1}, 2.0},
		{"zero", []float64{}, 0.0},
		{"zero", []float64{3.0}, 0.0},
		// empty
		{"empty", []float64{}, 0.0},
		{"empty", []float64{5.2}, 5.2},
		// non-existent
		{"five", []float64{}, 0.0},
		{"five", []float64{5.0}, 5.0},
		// invalid
		{"word", []float64{}, 0.0},
		{"word", []float64{5.0}, 5.0},
	}

	cfg := NewConfig(env)
	// Test GetFloat
	for _, td := range data {
		v := cfg.GetFloat(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetDuration(t *testing.T) {
	env := MapEnv{
		"5mins": "5m",
		"1hour": "1h",
		"zero":  "0",
		"empty": "",
		"word":  "henry",
	}

	data := []struct {
		key string
		fb  []time.Duration
		out time.Duration
	}{
		// valid
		{"5mins", []time.Duration{}, time.Minute * 5},
		{"1hour", []time.Duration{time.Second * 1}, time.Hour * 1},
		// zero
		{"zero", []time.Duration{}, 0},
		{"zero", []time.Duration{time.Second * 2}, 0},
		// empty
		{"empty", []time.Duration{}, 0},
		{"empty", []time.Duration{time.Second * 2}, time.Second * 2},
		// unset
		{"missing", []time.Duration{}, 0},
		{"missing", []time.Duration{time.Second * 2}, time.Second * 2},
		// invalid
		{"word", []time.Duration{}, 0},
		{"word", []time.Duration{time.Second * 5}, time.Second * 5},
	}

	cfg := NewConfig(env)

	// Test GetDuration
	for _, td := range data {
		v := cfg.GetDuration(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetBool(t *testing.T) {
	env := MapEnv{
		"empty": "",
		"t":     "t",
		"f":     "f",
		"1":     "1",
		"0":     "0",
		"true":  "true",
		"false": "false",
		"word":  "nonsense",
	}

	data := []struct {
		key string
		fb  []bool
		out bool
	}{
		// valid
		{"t", []bool{}, true},
		{"f", []bool{true}, false},
		{"1", []bool{}, true},
		{"0", []bool{true}, false},
		{"true", []bool{}, true},
		{"false", []bool{true}, false},
		// empty
		{"empty", []bool{}, false},
		{"empty", []bool{true}, true},
		// missing
		{"missing", []bool{}, false},
		{"missing", []bool{true}, true},
		// invalid
		{"word", []bool{}, false},
		{"word", []bool{true}, true},
	}

	cfg := NewConfig(env)

	// Test GetBool
	for _, td := range data {
		v := cfg.GetBool(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestStringify(t *testing.T) {
	data := []struct {
		in  interface{}
		out string
	}{
		{"", ""},
		{"plaintext", "plaintext"},
		{"A whole sentence", "A whole sentence"},
		{true, "true"},
		{false, "false"},
		{0, "0"},
		{1, "1"},
		{4.1, "4.1"},
		{time.Second * 60, "1m0s"},
	}

	for _, td := range data {
		s := stringify(td.in)
		if s != td.out {
			t.Errorf("Bad String for %#v. Expected=%v, Got=%v", td.in, td.out, s)
		}

	}
}

// Basic usage of Config.Get. Returns an empty string if variable is unset.
func ExampleConfig_Get() {
	// Set some test variables
	os.Setenv("TEST_NAME", "Bob Smith")
	os.Setenv("TEST_ADDRESS", "7, Dreary Lane")

	// New Config from environment
	cfg := NewConfig()

	fmt.Println(cfg.Get("TEST_NAME"))
	fmt.Println(cfg.Get("TEST_ADDRESS"))
	fmt.Println(cfg.Get("TEST_NONEXISTENT")) // unset variable

	// GetString is a synonym
	fmt.Println(cfg.GetString("TEST_NAME"))

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
	cfg := NewConfig()

	fmt.Println(cfg.Get("TEST_NAME", "default name"))       // fallback ignored
	fmt.Println(cfg.Get("TEST_ADDRESS", "default address")) // fallback ignored
	fmt.Println(cfg.Get("TEST_EMAIL", "test@example.com"))  // fallback ignored (var is empty, not unset)
	fmt.Println(cfg.Get("TEST_NONEXISTENT", "hi there!"))   // unset variable

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
	cfg := NewConfig()

	fmt.Println(cfg.GetInt("PORT"))
	fmt.Println(cfg.GetInt("PORT", 5000))        // fallback is ignored
	fmt.Println(cfg.GetInt("PING_INTERVAL"))     // returns zero value
	fmt.Println(cfg.GetInt("PING_INTERVAL", 60)) // returns fallback
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
	cfg := NewConfig()

	fmt.Printf("%0.2f\n", cfg.GetFloat("TOTAL_SCORE"))
	fmt.Printf("%0.1f\n", cfg.GetFloat("AVERAGE_SCORE"))
	fmt.Println(cfg.GetFloat("NON_EXISTENT_SCORE", 120.5))
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
	cfg := NewConfig()

	// returns time.Duration
	fmt.Println(cfg.GetDuration("DURATION_NAP"))
	fmt.Println(cfg.GetDuration("DURATION_EGG") * 2)
	// fallback with unset variable
	fmt.Println(cfg.GetDuration("DURATION_POWERNAP", time.Minute*45))
	// or an empty one
	fmt.Println(cfg.GetDuration("DURATION_BIG_EGG", time.Minute*10))
	fmt.Println(cfg.GetDuration("DURATION_MATCH").Minutes())

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
	cfg := NewConfig()

	// strconv.ParseBool() supports many formats
	fmt.Println(cfg.GetBool("LIKE_PEAS"))
	fmt.Println(cfg.GetBool("LIKE_CARROTS"))
	fmt.Println(cfg.GetBool("LIKE_BEANS"))
	fmt.Println(cfg.GetBool("LIKE_LIVER"))
	fmt.Println(cfg.GetBool("LIKE_TOMATOES"))
	fmt.Println(cfg.GetBool("LIKE_BVB"))
	fmt.Println(cfg.GetBool("LIKE_BAYERN"))

	// Fallback
	fmt.Println(cfg.GetBool("LIKE_BEER", true))

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
