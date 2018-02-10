//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-13
//

package aw

import (
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
		{"AlfredVersion", tAlfredVersion, EnvVarAlfredVersion},
		{"AlfredBuild", tAlfredBuild, EnvVarAlfredBuild},
		{"Theme", tTheme, EnvVarTheme},
		{"ThemeBackground", tThemeBackground, EnvVarThemeBackground},
		{"ThemeSelectionBackground", tThemeSelectionBackground, EnvVarThemeSelectionBackground},
		{"Preferences", tPreferences, EnvVarPreferences},
		{"Localhash", tLocalhash, EnvVarLocalhash},
		{"CacheDir", tCacheDir, EnvVarCacheDir},
		{"CacheDir", tDataDir, EnvVarDataDir},
	}

	ctx := NewConfig(testEnv)

	v := ctx.GetBool(EnvVarDebug)
	if v != tDebug {
		t.Errorf("bad Debug. Expected=%v, Got=%v", tDebug, v)
	}

	for _, td := range data {
		s := ctx.Get(td.key)
		if s != td.x {
			t.Errorf("Bad %s. Expected=%v, Got=%v", td.name, td.x, s)
		}
	}
}

func TestGet(t *testing.T) {
	env := mapEnv{
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

	e := NewConfig(env)

	// Verify env is the same
	for k, x := range env {
		v := e.Get(k)
		if v != x {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", k, x, v)
		}
	}

	// Test Get
	for _, td := range data {
		v := e.Get(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetInt(t *testing.T) {
	env := mapEnv{
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

	e := NewConfig(env)
	// Test GetInt
	for _, td := range data {
		v := e.GetInt(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetFloat(t *testing.T) {
	env := mapEnv{
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

	e := NewConfig(env)
	// Test GetFloat
	for _, td := range data {
		v := e.GetFloat(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetDuration(t *testing.T) {
	env := mapEnv{
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

	e := NewConfig(env)

	// Test GetDuration
	for _, td := range data {
		v := e.GetDuration(td.key, td.fb...)
		if v != td.out {
			t.Errorf("Bad '%s'. Expected=%v, Got=%v", td.key, td.out, v)
		}

	}
}

func TestGetBool(t *testing.T) {
	env := mapEnv{
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

	e := NewConfig(env)

	// Test GetBool
	for _, td := range data {
		v := e.GetBool(td.key, td.fb...)
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

func TestStoreCommands(t *testing.T) {
	data := []struct {
		s, x string
	}{
		// Unexportable
		{setConfigAS("name", "dean", tBundleID, false),
			`tell application "Alfred 3"
set configuration "name" to value "dean" in workflow "net.deanishe.awgo" exportable false
end tell`},
		// Exportable
		{setConfigAS("name", "dean", tBundleID, true),
			`tell application "Alfred 3"
set configuration "name" to value "dean" in workflow "net.deanishe.awgo" exportable true
end tell`},
		// Quotes in value
		{setConfigAS("name", `"dean"`, tBundleID, false),
			`tell application "Alfred 3"
set configuration "name" to value quote & "dean" & quote in workflow "net.deanishe.awgo" exportable false
end tell`},
	}

	for _, td := range data {
		if td.s != td.x {
			t.Errorf("Bad Store script. Expected=%v, Got=%v", td.x, td.s)
		}

	}
}
