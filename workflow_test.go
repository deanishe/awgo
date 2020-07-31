// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"errors"
	"fmt"

	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.deanishe.net/env"

	"github.com/deanishe/awgo/util"
)

// TestWorkflowValues tests workflow name, bundle ID etc.
func TestWorkflowValues(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		assert.Equal(t, tName, wf.Name(), "unexpected name")
		assert.Equal(t, tBundleID, wf.BundleID(), "unexpected bundle ID")
	})
}

// TestInvalidEnv executes workflow in an invalid environment.
func TestInvalidEnv(t *testing.T) {
	assert.Panics(t, func() { NewFromEnv(env.MapEnv{}) })
}

// Options correctly alter Workflow.
func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
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
			AddMagic(&mockMA{}),
			func(wf *Workflow) bool { return wf.magicActions.actions["test"] != nil },
			"Add Magic"},
		{
			RemoveMagic(logMA{}),
			func(wf *Workflow) bool { return wf.magicActions.actions["log"] == nil },
			"Remove Magic"},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("Option(%#v)", td.opt), func(t *testing.T) {
			t.Parallel()
			wf := New(td.opt)
			assert.True(t, td.test(wf), "option failed")
		})
	}
}

func TestWorkflow_Run(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		var called bool

		run := func() { called = true }
		wf.Run(run)

		assert.True(t, called, "run wasn't called")
	})
}

func TestWorkflow_Run_Rescue(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		me := &mockExit{}
		exitFunc = me.Exit
		defer func() { exitFunc = os.Exit }()
		wf.Run(func() { panic("aaaargh!") })
		assert.Equal(t, 1, me.code, "workflow did not catch panic")
	})
}

// TestWorkflowDir verifies that AwGo finds the right directory.
func TestWorkflow_Dir(t *testing.T) {
	t.Parallel()

	var (
		cwd string
		err error
	)

	cwd, err = os.Getwd()
	require.Nil(t, err, "Getwd failed")

	tests := []struct {
		in, x string
	}{
		{"testdata", "testdata"},
		{"testdata/subdir", "testdata"},
		{".", "."},
		{"", ""},
	}

	for _, td := range tests {
		td := td
		t.Run(fmt.Sprintf("findWorkflowRoot(%q)", td.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, findWorkflowRoot(td.in), "unexpected root")
		})
	}

	wf := New()
	assert.Equal(t, cwd, wf.Dir(), "unexpected workflow dir")
}

// Check that AW's directories exist
func TestWorkflow_awDirs(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		p := wf.awCacheDir()
		assert.True(t, util.PathExists(p), "AW cache dir does not exist")
		assert.True(t, strings.HasSuffix(p, "_aw"), "AW cache is not called '_aw'")

		p = wf.awDataDir()
		assert.True(t, util.PathExists(p), "AW data dir does not exist")
		assert.True(t, strings.HasSuffix(p, "_aw"), "AW data is not called '_aw'")
	})
}

// Check log is rotated
func TestWorkflow_logRotate(t *testing.T) {
	logInitialized = false // ensure log is created
	withTestWf(func(wf *Workflow) {
		wf.Configure(MaxLogSize(10))
		log.Print("more than 10 bytes")

		assert.True(t, util.PathExists(wf.LogFile()), "log file does not exist")

		logInitialized = false // ensure log is created
		wf.initializeLogging()
		assert.True(t, util.PathExists(wf.LogFile()+".1"), "log file not rotated")
	})
}

// Variables are correctly set
func TestWorkflow_Vars(t *testing.T) {
	t.Parallel()

	vars := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
		"key4": "val4",
		"key5": "val5",
	}

	withTestWf(func(wf *Workflow) {
		for k, v := range vars {
			wf.Var(k, v)
		}
		assert.Equal(t, vars, wf.Vars(), "Unexpected Vars")
	})
}

func TestWorkflow_Rerun(t *testing.T) {
	t.Parallel()

	withTestWf(func(wf *Workflow) {
		v := 0.1
		wf.Rerun(v)
		assert.Equal(t, v, wf.Feedback.rerun, "Unexpected Rerun")
	})
}

func TestWorkflow_Fatal(t *testing.T) {
	var exit bool
	exitFunc = func(code int) { exit = true }
	withTestWf(func(wf *Workflow) {
		wf.Fatal("")
		assert.True(t, exit, "Fatal did not exit")
	})

	exit = false
	withTestWf(func(wf *Workflow) {
		wf.FatalError(errors.New("some error"))
		assert.True(t, exit, "FatalError did not exit")
	})

	exit = false
	withTestWf(func(wf *Workflow) {
		wf.Fatalf("die")
		assert.True(t, exit, "Fatalf did not exit")
	})
}

func TestRunCommand(t *testing.T) {
	t.Parallel()

	err := runCommand("/usr/bin/true")
	assert.Nil(t, err, `call "/usr/bin/true" failed`)
	err = runCommand("/does/not/exist")
	assert.NotNil(t, err, `call to "/does/not/exist" returned no error`)
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
	// 1.2.0
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

func ExampleWorkflow_Warn() {
	wf := New()
	// Add some items
	wf.NewItem("Item One").
		Subtitle("Subtitle one")
	wf.NewItem("Item Two").
		Subtitle("Subtitle two")

	// Delete existing items, add a warning, then
	// immediately send feedback
	wf.Warn("Bad Items", "Those items are boring")

	// Output:
	// {
	//   "variables": {
	//     "AW_SESSION_ID": "test-session-id"
	//   },
	//   "items": [
	//     {
	//       "title": "Bad Items",
	//       "subtitle": "Those items are boring",
	//       "valid": false,
	//       "icon": {
	//         "path": "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
	//       }
	//     }
	//   ]
	// }
}

func ExampleArgVars() {
	// Set workflow variables from Alfred's Run Script Action
	av := NewArgVars()
	av.Arg("baz")        // Set output (i.e. next action's {query}) to "baz"
	av.Var("foo", "bar") // Set workflow variable "foo" to "bar"
	if err := av.Send(); err != nil {
		panic(err)
	}
	// Output: {"alfredworkflow":{"arg":"baz","variables":{"foo":"bar"}}}
}
