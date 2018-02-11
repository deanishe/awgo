//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-13
//

package aw

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/deanishe/awgo/util"
)

// Environment variables set by Alfred.
const (
	EnvVarVersion                  = "alfred_workflow_version"
	EnvVarName                     = "alfred_workflow_name"
	EnvVarBundleID                 = "alfred_workflow_bundleid"
	EnvVarUID                      = "alfred_workflow_uid"
	EnvVarDebug                    = "alfred_debug"
	EnvVarAlfredVersion            = "alfred_version"
	EnvVarAlfredBuild              = "alfred_version_build"
	EnvVarTheme                    = "alfred_theme"
	EnvVarThemeBackground          = "alfred_theme_background"
	EnvVarThemeSelectionBackground = "alfred_theme_selection_background"
	EnvVarPreferences              = "alfred_preferences"
	EnvVarLocalhash                = "alfred_preferences_localhash"
	EnvVarCacheDir                 = "alfred_workflow_cache"
	EnvVarDataDir                  = "alfred_workflow_data"
)

// Env is the datasource for configuration lookups.
// It is an optional parameter to NewConfig(). By specifying a custom Env,
// it's possible to populate the Config from an alternative source.
type Env interface {
	// Lookup retrieves the value of the variable named by key.
	//
	// It follows the same semantics as os.LookupEnv(). If a variable
	// is unset, the boolean will be false. If a variable is set, the
	// boolean will be true, but the variable may still be an empty
	// string.
	Lookup(key string) (string, bool)
}

// Config contains Alfred and workflow settings from environment variables.
type Config struct {
	Env
}

// NewConfig creates a new Config from environment variables.
// It accepts one optional Env argument. If an Env is passed, Config
// is initialised from that instead of the system environment.
func NewConfig(env ...Env) *Config {

	var (
		c *Config
		e Env
	)

	if len(env) > 0 {
		e = env[0]
	} else {
		e = sysEnv{}
	}

	c = &Config{e}
	return c
}

// Get returns the value for envvar "key".
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or an empty string.
//
// If a variable is set, but empty, its value is used.
func (c Config) Get(key string, fallback ...string) string {

	var fb string

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := c.Lookup(key)
	if !ok {
		return fb
	}
	return s
}

// GetString is a synonym for Get.
func (c Config) GetString(key string, fallback ...string) string {
	return c.Get(key, fallback...)
}

// GetInt returns the value for envvar "key" as an int.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.
//
// Values are parsed with strconv.ParseInt(). If strconv.ParseInt() fails,
// tries to parse the number with strconv.ParseFloat() and truncate it to an
// int.
func (c Config) GetInt(key string, fallback ...int) int {

	var fb int

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := c.Lookup(key)
	if !ok {
		return fb
	}

	i, err := parseInt(s)
	if err != nil {
		return fb
	}

	return int(i)
}

// GetFloat returns the value for envvar "key" as a float.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.0.
//
// Values are parsed with strconv.ParseFloat().
func (c Config) GetFloat(key string, fallback ...float64) float64 {

	var fb float64

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := c.Lookup(key)
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
func (c Config) GetDuration(key string, fallback ...time.Duration) time.Duration {

	var fb time.Duration

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := c.Lookup(key)
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
func (c Config) GetBool(key string, fallback ...bool) bool {

	var fb bool

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := c.Lookup(key)
	if !ok {
		return fb
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return fb
	}

	return b
}

// Save saves a value to the workflow's configuration.
func (c Config) Save(key string, value interface{}, export ...bool) error {

	var (
		bid = c.Get(EnvVarBundleID)
		val = fmt.Sprintf("%v", value)
		exp bool
	)
	if len(export) > 0 && export[0] {
		exp = true
	}

	script := setConfigAS(key, val, bid, exp)

	// log.Printf("AppleScript=\n%v", script)

	cmd := exec.Command("/usr/bin/osascript", "-l", "AppleScript", "-e", script)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Generate AppleScript to store a configuration variable.
func setConfigAS(key, value, bundleid string, export bool) string {

	key = util.QuoteAS(key)
	value = util.QuoteAS(value)
	bundleid = util.QuoteAS(bundleid)

	suffix := "exportable false"

	if export {
		suffix = "exportable true"
	}

	script := `tell application "Alfred 3"
set configuration %s to value %s in workflow %s %s
end tell`

	return fmt.Sprintf(script, key, value, bundleid, suffix)
}

// Check that minimum required values are set.
func validateConfig(c *Config) error {

	var (
		issues   []string
		required = map[string]string{
			EnvVarBundleID: c.Get(EnvVarBundleID),
			EnvVarCacheDir: c.Get(EnvVarCacheDir),
			EnvVarDataDir:  c.Get(EnvVarDataDir),
		}
	)

	for k, v := range required {
		if v == "" {
			issues = append(issues, k+" is not set")
		}
	}

	if issues != nil {
		return fmt.Errorf("Invalid Workflow environment: %s", strings.Join(issues, ", "))
	}

	return nil
}

// sysEnv implements Env based on the real environment.
type sysEnv struct{}

// Lookup wraps os.LookupEnv().
func (e sysEnv) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
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
