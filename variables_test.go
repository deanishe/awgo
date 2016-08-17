//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-07-31
//

package workflow

import (
	"encoding/json"
	"testing"
)

var (
	marshalVariablesTests []mvTest
	vs1, vs2, vs3, vs4    *VarSet
)

type mvTest struct {
	VarSet       *VarSet
	ExpectedJSON string
}

func init() {
	vs1 = &VarSet{}
	vs1.Arg("arg")
	vs2 = &VarSet{}
	vs2.Var("key", "value")
	vs3 = &VarSet{}
	vs3.Arg("arg")
	vs3.Var("key", "value")
	vs4 = &VarSet{}
	vs4.Arg("arg")
	vs4.Var("key1", "value1")
	vs4.Var("key2", "value2")
	marshalVariablesTests = []mvTest{
		{vs1, `"arg"`},
		{vs2, `{"alfredworkflow":{"variables":{"key":"value"}}}`},
		{vs3, `{"alfredworkflow":{"arg":"arg","variables":{"key":"value"}}}`},
		{vs4, `{"alfredworkflow":{"arg":"arg","variables":{"key1":"value1","key2":"value2"}}}`},
	}
}

func TestVariables(t *testing.T) {
	for i, test := range marshalVariablesTests {
		data, err := json.Marshal(test.VarSet)
		if err != nil {
			t.Errorf("#%d: marshal(%v): %v", i, test.VarSet, err)
			continue
		}

		if got, want := string(data), test.ExpectedJSON; got != want {
			t.Errorf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}
