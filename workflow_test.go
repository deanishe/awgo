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

var testOptions = []struct {
	opt  Option
	test func(wf *Workflow) bool
	desc string
}{
	{HelpURL("http://www.example.com"), func(wf *Workflow) bool { return wf.HelpURL == "http://www.example.com" }, "Set HelpURL"},
	{MaxResults(10), func(wf *Workflow) bool { return wf.MaxResults == 10 }, "Set MaxResults"},
	{LogPrefix("blah"), func(wf *Workflow) bool { return wf.LogPrefix == "blah" }, "Set LogPrefix"},
	{SortOptions(), func(wf *Workflow) bool { return wf.SortOptions == nil }, "Set SortOptions"},
}

func TestOptions(t *testing.T) {
	for _, td := range testOptions {
		wf := New(td.opt)
		if !td.test(wf) {
			t.Errorf("option %s failed", td.desc)
		}
	}
}

func TestParseInfo(t *testing.T) {
	info := DefaultWorkflow().Info()
	if info.BundleID != "net.deanishe.awgo" {
		t.Fatalf("Incorrect bundle ID: %v", info.BundleID)
	}

	if info.Author != "Dean Jackson" {
		t.Fatalf("Incorrect author: %v", info.Author)
	}

	if info.Description != "AwGo sample info.plist" {
		t.Fatalf("Incorrect description: %v", info.Description)
	}

	if info.Name != "AwGo" {
		t.Fatalf("Incorrect name: %v", info.Name)
	}

	if info.Website != "https://git.deanishe.net/deanishe/awgo" {
		t.Fatalf("Incorrect website: %v", info.Website)
	}
}

// TestParseVars tests that variables are read from info.plist
func TestParseVars(t *testing.T) {
	i := DefaultWorkflow().Info()
	if i.Var("exported_var") != "exported_value" {
		t.Fatalf("exported_var=%v, expected=exported_value", i.Var("exported_var"))
	}

	// Should unexported variables be ignored?
	if i.Var("unexported_var") != "unexported_value" {
		t.Fatalf("unexported_var=%v, expected=unexported_value", i.Var("unexported_var"))
	}
}

func ExampleInfoPlist_Var() {
	info := DefaultWorkflow().Info()
	fmt.Println(info.Var("exported_var"))
	// Output: exported_value
}

// New initialises a Workflow with the default settings. Name,
// bundle ID, version etc. are read from the environment and info.plist.
func ExampleNew() {
	wf := New()
	// BundleID is read from environment or info.plist
	fmt.Println(wf.BundleID())
	// Version is from info.plist
	fmt.Println(wf.Version())
	// Output:
	// net.deanishe.awgo
	// 0.2.2
}

// The normal way to create a new Item, but not the normal way to use it.
//
// Normally, when you're done adding Items, you call SendFeedback() to
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
