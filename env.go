// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"strings"
)

// Env is the data source for configuration lookups.
//
// Pass a custom implementation to NewFromEnv() to provide a custom
// source for the required workflow configuration settings.
//
// As an absolute minimum, the following variables must be set:
//
//     alfred_workflow_bundleid
//     alfred_workflow_cache
//     alfred_workflow_data
//
// See EnvVar* consts for all variables set by Alfred.
type Env interface {
	// Lookup retrieves the value of the variable named by key.
	//
	// It follows the same semantics as os.LookupEnv(). If a variable
	// is unset, the boolean will be false. If a variable is set, the
	// boolean will be true, but the variable may still be an empty
	// string.
	Lookup(key string) (string, bool)
}

// MapEnv is a testing helper that makes it simple to convert a map[string]string
// to an Env.
type MapEnv map[string]string

// Lookup implements Env. It returns values from the map.
func (env MapEnv) Lookup(key string) (string, bool) {
	s, ok := env[key]
	return s, ok
}

// sysEnv implements Env based on the real environment.
type sysEnv struct{}

// Lookup wraps os.LookupEnv().
func (e sysEnv) Lookup(key string) (string, bool) { return os.LookupEnv(key) }

// Check that minimum required values are set.
func validateEnv(env Env) error {

	var (
		issues   []string
		required = []string{
			EnvVarBundleID,
			EnvVarCacheDir,
			EnvVarDataDir,
		}
	)

	for _, k := range required {
		v, ok := env.Lookup(k)
		if !ok || v == "" {
			issues = append(issues, k+" is not set")
		}
	}

	if issues != nil {
		return fmt.Errorf("Invalid Workflow environment: %s", strings.Join(issues, ", "))
	}

	return nil
}
