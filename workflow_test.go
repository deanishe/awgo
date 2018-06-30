//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestWorkflowValues tests workflow name, bundle ID etc.
func TestWorkflowValues(t *testing.T) {

	withTestWf(func(wf *Workflow) {

		if wf.Name() != tName {
			t.Errorf("Bad Name. Expected=%s, Got=%s", tName, wf.Name())
		}
		if wf.BundleID() != tBundleID {
			t.Errorf("Bad BundleID. Expected=%s, Got=%s", tBundleID, wf.BundleID())
		}
	})
}

// TestOptions verifies that options correctly alter Workflow.
func TestOptions(t *testing.T) {

	data := []struct {
		opt  Option                  // option to set
		test func(wf *Workflow) bool // function to verify change was made
		desc string                  // test title
	}{
		{
			HelpURL("http://www.example.com"),
			func(wf *Workflow) bool { return wf.helpURL == "http://www.example.com" },
			"Set HelpURL"},
		{
			MaxResults(10),
			func(wf *Workflow) bool { return wf.maxResults == 10 },
			"Set MaxResults"},
		{
			LogPrefix("blah"),
			func(wf *Workflow) bool { return wf.logPrefix == "blah" },
			"Set LogPrefix"},
		{
			SessionName("SESH"),
			func(wf *Workflow) bool { return wf.sessionName == "SESH" },
			"Set SessionName"},
		{
			SortOptions(),
			func(wf *Workflow) bool { return wf.sortOptions == nil },
			"Set SortOptions"},
		{
			SuppressUIDs(true),
			func(wf *Workflow) bool { return wf.Feedback.NoUIDs == true },
			"Set SuppressUIDs"},
		{
			MagicPrefix("aw:"),
			func(wf *Workflow) bool { return wf.magicPrefix == "aw:" },
			"Set MagicPrefix"},
		{
			MaxLogSize(2048),
			func(wf *Workflow) bool { return wf.maxLogSize == 2048 },
			"Set MaxLogSize"},
		{
			TextErrors(true),
			func(wf *Workflow) bool { return wf.textErrors == true },
			"Set TextErrors"},
		{
			AddMagic(&testMA{}),
			func(wf *Workflow) bool { return wf.MagicActions.actions["test"] != nil },
			"Add Magic"},
		{
			RemoveMagic(logMA{}),
			func(wf *Workflow) bool { return wf.MagicActions.actions["log"] == nil },
			"Remove Magic"},
		{
			customEnv(mapEnv{
				"alfred_workflow_bundleid": "fakeid",
				"alfred_workflow_cache":    os.Getenv("alfred_workflow_cache"),
				"alfred_workflow_data":     os.Getenv("alfred_workflow_data"),
			}),
			func(wf *Workflow) bool { return wf.BundleID() == "fakeid" },
			"CustomEnv"},
	}

	for _, td := range data {

		wf := New(td.opt)

		if !td.test(wf) {
			t.Errorf("option %s failed", td.desc)
		}
	}
}

func TestWorkflowRun(t *testing.T) {

	withTestWf(func(wf *Workflow) {

		var called bool

		run := func() {
			called = true
		}

		wf.Run(run)

		if !called {
			t.Errorf("run wasn't called")
		}
	})
}

// TestWorkflowDir verifies that AwGo finds the right directory.
func TestWorkflowDir(t *testing.T) {

	withTestWf(func(wf *Workflow) {

		// Set up environment
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		subdir := "sub"
		if err := os.Mkdir(subdir, 0700); err != nil {
			t.Fatal(err)
		}

		// workflow root (alongside info.plist)
		if wf.Dir() != cwd {
			t.Errorf("Bad Dir (root). Expected=%v, Got=%v", cwd, wf.Dir())
		}

		// Change to subdirectory
		if err := os.Chdir(subdir); err != nil {
			t.Fatal(err)
		}

		// Reset cached path
		wf.dir = ""
		// Should find parent directory (where info.plist is)
		if wf.Dir() != cwd {
			t.Errorf("Bad Dir (sub). Expected=%v, Got=%v", cwd, wf.Dir())
		}
	})
}

// New initialises a Workflow with the default settings. Name,
// bundle ID, version etc. are read from the environment variables set by Alfred.
func ExampleNew() {
	wf := New()
	// Name is read from environment
	fmt.Println(wf.Name())
	// BundleID is read from environment
	fmt.Println(wf.BundleID())
	// Version is from info.plist
	fmt.Println(wf.Version())
	// Output:
	// AwGo
	// net.deanishe.awgo
	// 0.14
}

// Pass one or more Options to New() to configure the created Workflow.
func ExampleNew_withOptions() {
	wf := New(HelpURL("http://www.example.com"), MaxResults(200))
	fmt.Println(wf.helpURL)
	fmt.Println(wf.maxResults)
	// Output:
	// http://www.example.com
	// 200
}

// The normal way to create a new Item, but not the normal way to use it.
//
// Typically, when you're done adding Items, you call SendFeedback() to
// send the results to Alfred.
func ExampleWorkflow_NewItem() {
	wf := New()
	// Create a new item via the Workflow object, which will
	// track the Item and send it to Alfred when you call
	// Workflow.SendFeedback()
	//
	// Title is the only required value.
	it := wf.NewItem("First Result").
		Subtitle("Some details here")

	// Just to see what it looks like...
	data, _ := json.Marshal(it)
	fmt.Println(string(data))
	// Output: {"title":"First Result","subtitle":"Some details here","valid":false}
}

// Change Workflow's configuration after creation, then revert it.
func ExampleWorkflow_Configure() {
	wf := New()
	// Default settings (false and 0)
	fmt.Println(wf.textErrors)
	fmt.Println(wf.maxResults)
	// Turn text errors on, set max results and save Option to revert
	// to previous configuration
	previous := wf.Configure(TextErrors(true), MaxResults(200))
	fmt.Println(wf.textErrors)
	fmt.Println(wf.maxResults)
	// Revert to previous configuration
	wf.Configure(previous)
	fmt.Println(wf.textErrors)
	fmt.Println(wf.maxResults)
	// Output:
	// false
	// 0
	// true
	// 200
	// false
	// 0
}

func ExampleArgVars() {
	// Set workflow variables from Alfred's Run Script Action
	av := NewArgVars()
	av.Arg("baz")        // Set output (i.e. next action's {query}) to "baz"
	av.Var("foo", "bar") // Set workflow variable "foo" to "bar"
	av.Send()
	// Output: {"alfredworkflow":{"arg":"baz","variables":{"foo":"bar"}}}
}
