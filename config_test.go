// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfigEnv verifies that Config holds the expected values.
func TestConfigEnv(t *testing.T) {
	t.Parallel()

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
	assert.Equal(t, tDebug, v, "unexpected Debug")

	for _, td := range data {
		td := td // capture variable
		t.Run(fmt.Sprintf("Config.Get(%v)", td.name), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, cfg.Get(td.key), "unexpected result")
		})
	}
}

func TestBundleID(t *testing.T) {
	cfg := NewConfig()
	x := "net.deanishe.awgo"
	assert.Equal(t, x, cfg.getBundleID(), "unexpected bundle ID")

	x = "net.deanishe.awgo2"
	assert.Equal(t, x, cfg.getBundleID(x), "unexpected bundle ID")
}

// Basic usage of Config.Get. Returns an empty string if variable is unset.
func ExampleConfig_Get() {
	// Set some test variables
	_ = os.Setenv("TEST_NAME", "Bob Smith")
	_ = os.Setenv("TEST_ADDRESS", "7, Dreary Lane")

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
	_ = os.Setenv("TEST_NAME", "Bob Smith")
	_ = os.Setenv("TEST_ADDRESS", "7, Dreary Lane")
	_ = os.Setenv("TEST_EMAIL", "")

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
	_ = os.Setenv("PORT", "3000")
	_ = os.Setenv("PING_INTERVAL", "")

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
	_ = os.Setenv("TOTAL_SCORE", "172.3")
	_ = os.Setenv("AVERAGE_SCORE", "7.54")

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
	_ = os.Setenv("DURATION_NAP", "20m")
	_ = os.Setenv("DURATION_EGG", "5m")
	_ = os.Setenv("DURATION_BIG_EGG", "")
	_ = os.Setenv("DURATION_MATCH", "1.5h")

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
	_ = os.Setenv("LIKE_PEAS", "t")
	_ = os.Setenv("LIKE_CARROTS", "true")
	_ = os.Setenv("LIKE_BEANS", "1")
	_ = os.Setenv("LIKE_LIVER", "f")
	_ = os.Setenv("LIKE_TOMATOES", "0")
	_ = os.Setenv("LIKE_BVB", "false")
	_ = os.Setenv("LIKE_BAYERN", "FALSE")

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
		panicOnErr(os.Unsetenv(key))
	}
}
