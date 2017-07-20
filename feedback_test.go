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
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileItem(t *testing.T) {
	ipPath := filepath.Join(Dir(), "info.plist")
	ipShort := strings.Replace(ipPath, os.Getenv("HOME"), "~", -1)
	fb := Feedback{}
	it := fb.NewFileItem(ipPath)
	if it.title != "info.plist" {
		t.Fatalf("Incorrect title: %v", it.Title)
	}
	if *it.subtitle != ipShort {
		t.Fatalf("Incorrect subtitle: %v", it.Subtitle)
	}

	if *it.uid != ipPath {
		t.Fatalf("Incorrect UID: %v", it.UID)
	}

	if it.file != true {
		t.Fatalf("Incorrect file: %v", it.file)
	}

	if it.icon.Type != "fileicon" {
		t.Fatalf("Incorrect type: %v", it.icon.Type)
	}

	if it.icon.Value != ipPath {
		t.Fatalf("Incorrect Value: %v", it.icon.Value)
	}
}

func TestSetIcon(t *testing.T) {
	it := Item{}
	it.Icon(&Icon{"first", "fileicon"})
	if it.icon.Value != "first" {
		t.Fatalf("Incorrect icon value: %v", it.icon.Value)
	}

	if it.icon.Type != "fileicon" {
		t.Fatalf("Incorrect type: %v", it.icon.Type)
	}
}

func p(s string) *string {
	var v *string
	v = &s
	return v
}

var marshalItemTests = []struct {
	Item         *Item
	ExpectedJSON string
}{
	// Minimal item
	{Item: &Item{title: "title"},
		ExpectedJSON: `{"title":"title","valid":false}`},
	// With UID
	{Item: &Item{title: "title", uid: p("xxx-yyy")},
		ExpectedJSON: `{"title":"title","uid":"xxx-yyy","valid":false}`},
	// With autocomplete
	{Item: &Item{title: "title", autocomplete: p("xxx-yyy")},
		ExpectedJSON: `{"title":"title","autocomplete":"xxx-yyy","valid":false}`},
	// With empty autocomplete
	{Item: &Item{title: "title", autocomplete: p("")},
		ExpectedJSON: `{"title":"title","autocomplete":"","valid":false}`},
	// With subtitle
	{Item: &Item{title: "title", subtitle: p("subtitle")},
		ExpectedJSON: `{"title":"title","subtitle":"subtitle","valid":false}`},
	// Alternate subtitle
	{Item: &Item{title: "title", subtitle: p("subtitle"),
		mods: map[string]*Modifier{
			"cmd": &Modifier{
				Key:      "cmd",
				subtitle: p("command sub")}}},
		ExpectedJSON: `{"title":"title","subtitle":"subtitle",` +
			`"valid":false,"mods":{"cmd":{"subtitle":"command sub"}}}`},
	// Valid item
	{Item: &Item{title: "title", valid: true},
		ExpectedJSON: `{"title":"title","valid":true}`},
	// With arg
	{Item: &Item{title: "title", arg: p("arg1")},
		ExpectedJSON: `{"title":"title","arg":"arg1","valid":false}`},
	// Empty arg
	{Item: &Item{title: "title", arg: p("")},
		ExpectedJSON: `{"title":"title","arg":"","valid":false}`},
	// Arg contains escapes
	{Item: &Item{title: "title", arg: p("\x00arg\x00")},
		ExpectedJSON: `{"title":"title","arg":"\u0000arg\u0000","valid":false}`},
	// Valid with arg
	{Item: &Item{title: "title", arg: p("arg1"), valid: true},
		ExpectedJSON: `{"title":"title","arg":"arg1","valid":true}`},
	// With icon
	{Item: &Item{title: "title",
		icon: &Icon{Value: "icon.png", Type: ""}},
		ExpectedJSON: `{"title":"title","valid":false,"icon":{"path":"icon.png"}}`},
	// With file icon
	{Item: &Item{title: "title",
		icon: &Icon{Value: "icon.png", Type: "fileicon"}},
		ExpectedJSON: `{"title":"title","valid":false,"icon":{"path":"icon.png","type":"fileicon"}}`},
	// With filetype icon
	{Item: &Item{title: "title",
		icon: &Icon{Value: "public.folder", Type: "filetype"}},
		ExpectedJSON: `{"title":"title","valid":false,"icon":{"path":"public.folder","type":"filetype"}}`},
	// With type = file
	{Item: &Item{title: "title", file: true},
		ExpectedJSON: `{"title":"title","valid":false,"type":"file"}`},
	// With copy text
	{Item: &Item{title: "title", copytext: p("copy")},
		ExpectedJSON: `{"title":"title","valid":false,"text":{"copy":"copy"}}`},
	// With large text
	{Item: &Item{title: "title", largetype: p("large")},
		ExpectedJSON: `{"title":"title","valid":false,"text":{"largetype":"large"}}`},
	// With copy and large text
	{Item: &Item{title: "title", copytext: p("copy"), largetype: p("large")},
		ExpectedJSON: `{"title":"title","valid":false,"text":{"copy":"copy","largetype":"large"}}`},
	// With arg and variable
	{Item: &Item{title: "title", arg: p("value"), vars: map[string]string{"foo": "bar"}},
		// ExpectedJSON: `{"title":"title","arg":"{\"alfredworkflow\":{\"arg\":\"value\",\"variables\":{\"foo\":\"bar\"}}}","valid":false}`},
		ExpectedJSON: `{"title":"title","arg":"value","valid":false,"variables":{"foo":"bar"}}`},
}

var marshalModifierTests = []struct {
	Mod          *Modifier
	ExpectedJSON string
}{
	// Empty item
	{Mod: &Modifier{},
		ExpectedJSON: `{}`},
	// With arg
	{Mod: &Modifier{arg: p("title")},
		ExpectedJSON: `{"arg":"title"}`},
	// Empty arg
	{Mod: &Modifier{arg: p("")},
		ExpectedJSON: `{"arg":""}`},
	// With subtitle
	{Mod: &Modifier{subtitle: p("sub here")},
		ExpectedJSON: `{"subtitle":"sub here"}`},
	// valid
	{Mod: &Modifier{valid: true, validSet: true},
		ExpectedJSON: `{"valid":true}`},
	// icon
	{Mod: &Modifier{icon: &Icon{"icon.png", ""}},
		ExpectedJSON: `{"icon":{"path":"icon.png"}}`},
	// With all
	{Mod: &Modifier{
		arg:      p("title"),
		subtitle: p("sub here"),
		valid:    true,
	},
		ExpectedJSON: `{"arg":"title","subtitle":"sub here","valid":true}`},
	// With variable
	{Mod: &Modifier{
		arg:      p("title"),
		subtitle: p("sub here"),
		valid:    true,
		vars:     map[string]string{"foo": "bar"},
	},
		ExpectedJSON: `{"arg":"title","subtitle":"sub here","valid":true,"variables":{"foo":"bar"}}`},
}

var marshalArgTests = []struct {
	Arg          *ArgVars
	ExpectedJSON string
}{
	// Empty
	{Arg: &ArgVars{},
		ExpectedJSON: `""`},
	// With arg
	{Arg: &ArgVars{arg: p("title")},
		ExpectedJSON: `"title"`},
	// With non-ASCII arg
	{Arg: &ArgVars{arg: p("fübär")},
		ExpectedJSON: `"fübär"`},
	// With escapes
	{Arg: &ArgVars{arg: p("\x00")},
		ExpectedJSON: `"\u0000"`},
	// With variable
	{Arg: &ArgVars{vars: map[string]string{"foo": "bar"}},
		ExpectedJSON: `{"alfredworkflow":{"variables":{"foo":"bar"}}}`},
	// Multiple variables
	{Arg: &ArgVars{vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
		ExpectedJSON: `{"alfredworkflow":{"variables":{"ducky":"fuzz","foo":"bar"}}}`},
	// Multiple variables and arg
	{Arg: &ArgVars{arg: p("title"), vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
		ExpectedJSON: `{"alfredworkflow":{"arg":"title","variables":{"ducky":"fuzz","foo":"bar"}}}`},
}

var stringifyArgTests = []struct {
	Arg            *ArgVars
	ExpectedString string
}{
	// Empty
	{Arg: &ArgVars{},
		ExpectedString: ""},
	// With arg
	{Arg: &ArgVars{arg: p("title")},
		ExpectedString: "title"},
	// With non-ASCII
	{Arg: &ArgVars{arg: p("fübär")},
		ExpectedString: "fübär"},
	// With escapes
	{Arg: &ArgVars{arg: p("\x00")},
		ExpectedString: "\x00"},
}

// TestEmpty asserts feedback is empty.
func TestEmpty(t *testing.T) {
	fb := NewFeedback()
	if !fb.IsEmpty() {
		t.Errorf("Feedback not empty.")
	}
	fb.NewItem("test")
	if fb.IsEmpty() {
		t.Errorf("Feedback empty.")
	}
}

func TestMarshalItem(t *testing.T) {
	for i, test := range marshalItemTests {
		// log.Printf("#%d: %v", i, test.Item)
		data, err := json.Marshal(test.Item)
		if err != nil {
			t.Errorf("#%d: marshal(%v): %v", i, test.Item, err)
			continue
		}

		if got, want := string(data), test.ExpectedJSON; got != want {
			t.Fatalf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}

func TestMarshalModifier(t *testing.T) {
	for i, test := range marshalModifierTests {
		data, err := json.Marshal(test.Mod)
		if err != nil {
			t.Errorf("#%d: marshal(%v): %v", i, test.Mod, err)
			continue
		}

		if got, want := string(data), test.ExpectedJSON; got != want {
			t.Fatalf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}

func TestMarshalArg(t *testing.T) {
	for i, test := range marshalArgTests {
		data, err := json.Marshal(test.Arg)
		if err != nil {
			t.Errorf("#%d: marshal(%v): %v", i, test.Arg, err)
			continue
		}

		if got, want := string(data), test.ExpectedJSON; got != want {
			t.Errorf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}

func TestStringifyArg(t *testing.T) {
	for i, test := range stringifyArgTests {
		s, err := test.Arg.String()
		if err != nil {
			t.Errorf("#%d: string(%v): %v", i, test.Arg, err)
			continue
		}
		if got, want := s, test.ExpectedString; got != want {
			t.Errorf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}

func TestMarshalFeedback(t *testing.T) {
	// Empty feedback
	fb := NewFeedback()
	want := `{"items":[]}`
	got, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error marshalling feedback: got: %s want: %s: %v",
			got, want, err)
	}
	if string(got) != want {
		t.Fatalf("Incorrect feedback: got: %s, wanted: %s", got, want)
	}

	// Feedback with item
	// want = `<items><item valid="no"><title>item 1</title></item></items>`
	want = `{"items":[{"title":"item 1","valid":false}]}`
	fb.NewItem("item 1")

	got, err = json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error marshalling feedback: got: %s want: %s: %v",
			got, want, err)
	}
	if string(got) != want {
		t.Fatalf("Wrong feedback JSON. Expected=%s, got=%s", want, got)
	}

}

// TestModifiersInheritVars tests that Modifiers inherit variables from their
// parent Item
func TestModifiersInheritVars(t *testing.T) {
	fb := NewFeedback()
	it := fb.NewItem("title")
	it.Var("foo", "bar")
	m := it.NewModifier("cmd")

	if m.Vars()["foo"] != "bar" {
		t.Fatalf("Modifier var has wrong value. Expected=bar, Received=%v", m.Vars()["foo"])
	}
}

// TestFeedbackRerun verifies that rerun is properly set.
func TestFeedbackRerun(t *testing.T) {
	fb := NewFeedback()

	fb.Rerun(1.5)

	want := `{"rerun":1.5,"items":[]}`
	got, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error serializing feedback: got: %s want: %s: %s", got, want, err)
	}
	if string(got) != want {
		t.Fatalf("Wrong feedback JSON. Expected=%s, got=%s", want, got)
	}
}

// TestFeedbackVars tests if vars are properly inherited by Items and Modifiers
func TestFeedbackVars(t *testing.T) {
	fb := NewFeedback()

	fb.Var("foo", "bar")
	if fb.Vars()["foo"] != "bar" {
		t.Fatalf("Feedback var has wrong value. Expected=bar, Received=%v", fb.Vars()["foo"])
	}

	want := `{"variables":{"foo":"bar"},"items":[]}`
	got, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error serializing feedback: got: %s want: %s: %s", got, want, err)
	}
	if string(got) != want {
		t.Fatalf("Wrong feedback JSON. Expected=%s, got=%s", want, got)
	}

	// Top-level vars are not inherited
	it := fb.NewItem("title")
	if it.Vars()["foo"] != "" {
		t.Fatalf("Item var has wrong value. Expected='', Received=%v", it.Vars()["foo"])
	}

	// Modifier inherits Item vars
	it.Var("foo", "baz")
	m := it.NewModifier("cmd")
	if m.Vars()["foo"] != "baz" {
		t.Fatalf("Modifier var has wrong value. Expected=baz, Received=%v", m.Vars()["foo"])
	}
}

func ExampleArgVars() {
	// Set workflow variables from Alfred's Run Script Action
	av := NewArgVars()
	av.Arg("baz")        // Set output (i.e. next action's {query}) to "baz"
	av.Var("foo", "bar") // Set workflow variable "foo" to "bar"
	if s, err := av.String(); err == nil {
		fmt.Print(s)
	}
	// Output: {"alfredworkflow":{"arg":"baz","variables":{"foo":"bar"}}}
}
