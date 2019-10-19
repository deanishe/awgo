// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/deanishe/awgo/util"
)

// Environment variables containing workflow and Alfred info.
//
// Read the values with os.Getenv(EnvVarName) or via Config:
//
//    // Returns a string
//    Config.Get(EnvVarName)
//    // Parse string into a bool
//    Config.GetBool(EnvVarDebug)
//
const (
	// Workflow info assigned in Alfred Preferences
	EnvVarName     = "alfred_workflow_name"     // Name of workflow
	EnvVarBundleID = "alfred_workflow_bundleid" // Bundle ID
	EnvVarVersion  = "alfred_workflow_version"  // Workflow version

	EnvVarUID = "alfred_workflow_uid" // Random UID assigned by Alfred

	// Workflow storage directories
	EnvVarCacheDir = "alfred_workflow_cache" // For temporary data
	EnvVarDataDir  = "alfred_workflow_data"  // For permanent data

	// Set to 1 when Alfred's debugger is open
	EnvVarDebug = "alfred_debug"

	// Theme info. Colours are in rgba format, e.g. "rgba(255,255,255,1.0)"
	EnvVarTheme            = "alfred_theme"                      // ID of user's selected theme
	EnvVarThemeBG          = "alfred_theme_background"           // Background colour
	EnvVarThemeSelectionBG = "alfred_theme_selection_background" // BG colour of selected item

	// Alfred info
	EnvVarAlfredVersion = "alfred_version"       // Alfred's version number
	EnvVarAlfredBuild   = "alfred_version_build" // Alfred's build number
	EnvVarPreferences   = "alfred_preferences"   // Path to "Alfred.alfredpreferences" file
	// Machine-specific hash. Machine preferences are stored in
	// Alfred.alfredpreferences/local/<hash>
	EnvVarLocalhash = "alfred_preferences_localhash"
)

// mockable JS script runner
var runJS = func(script string) error {
	_, err := util.RunJS(script)
	return err
}

// Config loads workflow settings from Alfred's environment variables.
//
// The Get* methods read a variable from the environment, converting it to
// the desired type, and the Set() method saves a variable to info.plist.
//
// NOTE: Because calling Alfred via AppleScript is very slow (~0.2s/call),
// Config users a "Doer" API for setting variables, whereby calls are collected
// and all executed at once when Config.Do() is called:
//
//     cfg := NewConfig()
//     if err := cfg.Set("key1", "value1").Set("key2", "value2").Do(); err != nil {
//         // handle error
//     }
//
// Finally, you can use Config.To() to populate a struct from environment
// variables, and Config.From() to read a struct's fields and save them
// to info.plist.
type Config struct {
	Env
	scripts []string
}

// NewConfig creates a new Config from the environment.
//
// It accepts one optional Env argument. If an Env is passed, Config
// is initialised from that instead of the system environment.
func NewConfig(env ...Env) *Config {
	var e Env
	if len(env) > 0 {
		e = env[0]
	} else {
		e = sysEnv{}
	}
	return &Config{
		Env:     e,
		scripts: []string{},
	}
}

// Get returns the value for envvar "key".
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or an empty string.
//
// If a variable is set, but empty, its value is used.
func (cfg *Config) Get(key string, fallback ...string) string {
	var fb string

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := cfg.Lookup(key)
	if !ok {
		return fb
	}
	return s
}

// GetString is a synonym for Get.
func (cfg *Config) GetString(key string, fallback ...string) string {
	return cfg.Get(key, fallback...)
}

// GetInt returns the value for envvar "key" as an int.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.
//
// Values are parsed with strconv.ParseInt(). If strconv.ParseInt() fails,
// tries to parse the number with strconv.ParseFloat() and truncate it to an
// int.
func (cfg *Config) GetInt(key string, fallback ...int) int {
	var fb int

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := cfg.Lookup(key)
	if !ok {
		return fb
	}

	i, err := parseInt(s)
	if err != nil {
		return fb
	}

	return i
}

// GetFloat returns the value for envvar "key" as a float.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.0.
//
// Values are parsed with strconv.ParseFloat().
func (cfg *Config) GetFloat(key string, fallback ...float64) float64 {
	var fb float64

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := cfg.Lookup(key)
	if !ok {
		return fb
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fb
	}

	return n
}

// GetDuration returns the value for envvar "key" as a time.Duration.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.
//
// Values are parsed with time.ParseDuration().
func (cfg *Config) GetDuration(key string, fallback ...time.Duration) time.Duration {
	var fb time.Duration

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := cfg.Lookup(key)
	if !ok {
		return fb
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return fb
	}

	return d
}

// GetBool returns the value for envvar "key" as a boolean.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or false.
//
// Values are parsed with strconv.ParseBool().
func (cfg *Config) GetBool(key string, fallback ...bool) bool {
	var fb bool

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := cfg.Lookup(key)
	if !ok {
		return fb
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return fb
	}

	return b
}

// Set saves a workflow variable to info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, it defaults to the current workflow's.
func (cfg *Config) Set(key, value string, export bool, bundleID ...string) *Config {
	bid := cfg.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"toValue":    value,
		"inWorkflow": bid,
		"exportable": export,
	}

	return cfg.addScript(scriptSetConfig, key, opts)
}

// Unset removes a workflow variable from info.plist.
//
// It accepts one optional bundleID argument, which is the bundle ID of the
// workflow whose configuration should be changed.
// If not specified, it defaults to the current workflow's.
func (cfg *Config) Unset(key string, bundleID ...string) *Config {
	bid := cfg.getBundleID(bundleID...)
	opts := map[string]interface{}{
		"inWorkflow": bid,
	}

	return cfg.addScript(scriptRmConfig, key, opts)
}

// Do calls Alfred and runs the accumulated actions.
//
// Returns an error if there are no commands to run, or if the call to Alfred fails.
// Succeed or fail, any accumulated scripts and errors are cleared when Do()
// is called.
func (cfg *Config) Do() error {
	if len(cfg.scripts) == 0 {
		return errors.New("no commands to run")
	}

	script := strings.Join(cfg.scripts, "\n")
	// reset
	cfg.scripts = []string{}

	return runJS(script)
}

// Extract bundle ID from argument or default.
func (cfg *Config) getBundleID(bundleID ...string) string {
	if len(bundleID) > 0 {
		return bundleID[0]
	}

	bid, _ := cfg.Lookup(EnvVarBundleID)
	return bid
}

// Add a JavaScript that takes two arguments, a string and an object.
func (cfg *Config) addScript(script, name string, opts map[string]interface{}) *Config {
	script = fmt.Sprintf(script, util.QuoteJS(scriptAppName()), util.QuoteJS(name), util.QuoteJS(opts))
	cfg.scripts = append(cfg.scripts, script)

	return cfg
}

// parse an int, falling back to parsing it as a float
func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return int(i), nil
	}

	// Try to parse as float, then convert
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid int: %v", s)
	}
	return int(n), nil
}

// Convert interface{} to a string.
func stringify(v interface{}) string { return fmt.Sprintf("%v", v) }
