//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-09
//

package aw

import (
	"encoding/json"
	"fmt"
)

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
	fmt.Println(wf.HelpURL)
	fmt.Println(wf.MaxResults)
	// Output:
	// http://www.example.com
	// 200
}

// Change Workflow's configuration after creation, then revert it.
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
