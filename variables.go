//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-07-31
//

package workflow

import "encoding/json"

// VarSet is a collection of environment/workflow variables that can
// be sent to Alfred from a normal Run Script Action (i.e. not a Script Filter).
type VarSet struct {
	arg  *string
	vars map[string]string
}

// Var sets a variable.
func (v *VarSet) Var(key, value string) *VarSet {
	if v.vars == nil {
		v.vars = map[string]string{}
	}
	v.vars[key] = value
	return v
}

// Arg sets the main output.
func (v *VarSet) Arg(value string) *VarSet {
	v.arg = &value
	return v
}

// String returns a string suitable for sending to Alfred.
// Basically a JSON string, not []byte.
func (v *VarSet) String() (string, error) {
	if len(v.vars) == 0 {
		return *v.arg, nil
	}
	data, err := v.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalJSON serialises VarSet to JSON.
func (v *VarSet) MarshalJSON() ([]byte, error) {
	if v.vars == nil {
		if v.arg == nil {
			return []byte{}, nil
		}
		// No variables, so just return `arg`
		return json.Marshal(&v.arg)
	}

	return json.Marshal(&struct {
		Root interface{} `json:"alfredworkflow"`
	}{
		Root: &struct {
			Arg  *string           `json:"arg,omitempty"`
			Vars map[string]string `json:"variables"`
		}{
			Arg:  v.arg,
			Vars: v.vars,
		},
	})
}
