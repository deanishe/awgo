//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"encoding/json"
	"fmt"
	"testing"
)

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
	{RemoveMagic(openLogMagic{}), func(wf *Workflow) bool { return wf.MagicActions["log"] == nil }, "Remove Magic"},
}

func TestOptions(t *testing.T) {
	for _, td := range testOptions {
		wf := New(td.opt)
		if !td.test(wf) {
			t.Errorf("option %s failed", td.desc)
		}
	}
}

// TestWorkflowValues tests workflow name, bundle ID etc.
func TestWorkflowValues(t *testing.T) {
	name := "AwGo"
	bundleID := "net.deanishe.awgo"
	wf := New()
	if wf.Name() != name {
		t.Errorf("wrong name. Expected=%s, Got=%s", name, wf.Name())
	}
	if wf.BundleID() != bundleID {
		t.Errorf("wrong bundle ID. Expected=%s, Got=%s", bundleID, wf.BundleID())
	}
}

// New initialises a Workflow with the default settings. Name,
// bundle ID, version etc. are read from the environment variables set by Alfred.
func ExampleNew() {
	wf := New()
	// BundleID is read from environment or info.plist
	fmt.Println(wf.BundleID())
	// Version is from info.plist
	fmt.Println(wf.Version())
	// Output:
	// net.deanishe.awgo
	// 0.13
}

// Pass one or more Options to New() to configure the created Workflow.
func ExampleNew_withOptions() {
	wf := New(HelpURL("http://www.example.com"), MaxResults(200))
	fmt.Println(wf.HelpURL)
	fmt.Println(wf.MaxResults)
	// Output:
	// http://www.example.com
	// 200
}

// Temporarily change Workflow's behaviour then revert it.
func ExampleOption() {
	wf := New()
	// Default settings (false and 0)
	fmt.Println(wf.TextErrors)
	fmt.Println(wf.MaxResults)
	// Turn text errors on, set max results and save Option to revert
	// to previous configuration
	previous := wf.Configure(TextErrors(true), MaxResults(200))
	fmt.Println(wf.TextErrors)
	fmt.Println(wf.MaxResults)
	// Revert to previous configuration
	wf.Configure(previous)
	fmt.Println(wf.TextErrors)
	fmt.Println(wf.MaxResults)
	// Output:
	// false
	// 0
	// true
	// 200
	// false
	// 0
}

// The normal way to create a new Item, but not the normal way to use it.
//
// Typically, when you're done adding Items, you call SendFeedback() to
// send the results to Alfred.
func ExampleNewItem() {
	// Create a new item via the default Workflow object, which will
	// track the Item and send it to Alfred when you call SendFeedback()
	//
	// Title is the only required value.
	it := NewItem("First Result").
		Subtitle("Some details here")

	// Just to see what it looks like...
	data, _ := json.Marshal(it)
	fmt.Println(string(data))
	// Output: {"title":"First Result","subtitle":"Some details here","valid":false}
}
