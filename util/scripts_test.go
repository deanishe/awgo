//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

package util

import "testing"

func TestAppleScriptify(t *testing.T) {
	data := []struct {
		in, out string
	}{
		{"", ""},
		{"simple", "simple"},
		{"with spaces", "with spaces"},
		{`has "quotes" within`, `has " & quote & "quotes" & quote & " within`},
		{`"within quotes"`, `" & quote & "within quotes" & quote & "`},
		{`"`, `" & quote & "`},
	}

	for _, td := range data {
		s := AppleScriptify(td.in)
		if s != td.out {
			t.Errorf("Bad AppleScript escape. Expected=%v, Got=%v", td.out, s)
		}

	}
}
