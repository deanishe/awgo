//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-09
//

package aw

import "testing"

// Opens workflow's log file.
type testMagicAction struct{}

func (a testMagicAction) Keyword() string     { return "test" }
func (a testMagicAction) Description() string { return "Just a test" }
func (a testMagicAction) RunText() string     { return "Performing test…" }
func (a testMagicAction) Run() error          { return nil }

var testOptions = []struct {
	opt  Option
	test func(wf *Workflow) bool
	desc string
}{
	{HelpURL("http://www.example.com"), func(wf *Workflow) bool { return wf.HelpURL == "http://www.example.com" }, "Set HelpURL"},
	{MaxResults(10), func(wf *Workflow) bool { return wf.MaxResults == 10 }, "Set MaxResults"},
	{LogPrefix("blah"), func(wf *Workflow) bool { return wf.LogPrefix == "blah" }, "Set LogPrefix"},
	{SortOptions(), func(wf *Workflow) bool { return wf.SortOptions == nil }, "Set SortOptions"},
	{MagicPrefix("aw:"), func(wf *Workflow) bool { return wf.magicPrefix == "aw:" }, "Set MagicPrefix"},
	{MaxLogSize(2048), func(wf *Workflow) bool { return wf.MaxLogSize == 2048 }, "Set MaxLogSize"},
	{TextErrors(true), func(wf *Workflow) bool { return wf.TextErrors == true }, "Set TextErrors"},
	{AddMagic(testMagicAction{}), func(wf *Workflow) bool { return wf.MagicActions["test"] != nil }, "Add Magic"},
	{RemoveMagic(logMA{}), func(wf *Workflow) bool { return wf.MagicActions["log"] == nil }, "Remove Magic"},
}

func TestOptions(t *testing.T) {
	// TODO: decouple from env
	for _, td := range testOptions {
		wf := New(td.opt)
		if !td.test(wf) {
			t.Errorf("option %s failed", td.desc)
		}
	}
}
