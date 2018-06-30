//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-06-30
//

package aw

import "os"

// Env is the datasource for configuration lookups.
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

// sysEnv implements Env based on the real environment.
type sysEnv struct{}

// Lookup wraps os.LookupEnv().
func (e sysEnv) Lookup(key string) (string, bool) { return os.LookupEnv(key) }
