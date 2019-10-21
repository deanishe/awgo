// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"sort"

	"github.com/deanishe/go-env"
)

// To populates (tagged) struct v with values from the environment.
func (cfg *Config) To(v interface{}) error {
	return env.Bind(v, cfg)
}

// From saves the fields of (tagged) struct v to the workflow's settings in Alfred.
// All supported and unignored fields are saved by default. The behaviour can be
// customised by passing in options from deanishe/go-env, such as env.IgnoreZeroValues
// to omit any fields set to zero values.
//
// https://godoc.org/github.com/deanishe/go-env#DumpOption
func (cfg *Config) From(v interface{}, opt ...env.DumpOption) error {
	variables, err := env.Dump(v, opt...)
	if err != nil {
		return err
	}

	return cfg.setMulti(variables, false)
}

// setMulti batches the saving of multiple variables.
func (cfg *Config) setMulti(variables map[string]string, export bool) error {
	// sort keys to make the output testable
	var keys []string
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cfg.Set(k, variables[k], export)
	}

	return cfg.Do()
}
