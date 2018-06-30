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
	"strconv"
	"strings"
	"time"
)

// Get returns the value for envvar "key".
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or an empty string.
//
// If a variable is set, but empty, its value is used.
func (a *Alfred) Get(key string, fallback ...string) string {

	var fb string

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := a.Lookup(key)
	if !ok {
		return fb
	}
	return s
}

// GetString is a synonym for Get.
func (a *Alfred) GetString(key string, fallback ...string) string {
	return a.Get(key, fallback...)
}

// GetInt returns the value for envvar "key" as an int.
// It accepts one optional "fallback" argument. If no envvar is set, returns
// fallback or 0.
//
// Values are parsed with strconv.ParseInt(). If strconv.ParseInt() fails,
// tries to parse the number with strconv.ParseFloat() and truncate it to an
// int.
func (a *Alfred) GetInt(key string, fallback ...int) int {

	var fb int

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := a.Lookup(key)
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
func (a *Alfred) GetFloat(key string, fallback ...float64) float64 {

	var fb float64

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := a.Lookup(key)
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
func (a *Alfred) GetDuration(key string, fallback ...time.Duration) time.Duration {

	var fb time.Duration

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := a.Lookup(key)
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
func (a *Alfred) GetBool(key string, fallback ...bool) bool {

	var fb bool

	if len(fallback) > 0 {
		fb = fallback[0]
	}
	s, ok := a.Lookup(key)
	if !ok {
		return fb
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return fb
	}

	return b
}

// Check that minimum required values are set.
func validateAlfred(a *Alfred) error {

	var (
		issues   []string
		required = map[string]string{
			EnvVarBundleID: a.Get(EnvVarBundleID),
			EnvVarCacheDir: a.Get(EnvVarCacheDir),
			EnvVarDataDir:  a.Get(EnvVarDataDir),
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
