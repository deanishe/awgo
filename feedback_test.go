//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package workflow

import (
	"encoding/json"
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
	if it.Title != "info.plist" {
		t.Fatalf("Incorrect title: %v", it.Title)
	}
	if it.Subtitle != ipShort {
		t.Fatalf("Incorrect subtitle: %v", it.Subtitle)
	}

	if it.UID != ipPath {
		t.Fatalf("Incorrect UID: %v", it.UID)
	}

	if it.IsFile != true {
		t.Fatalf("Incorrect IsFile: %v", it.IsFile)
	}

	if it.Icon.Type != "fileicon" {
		t.Fatalf("Incorrect type: %v", it.Icon.Type)
	}

	if it.Icon.Value != ipPath {
		t.Fatalf("Incorrect Value: %v", it.Icon.Value)
	}
}

func TestSetIcon(t *testing.T) {
	it := Item{}
	it.SetIcon("first", "fileicon")
	if it.Icon.Value != "first" {
		t.Fatalf("Incorrect icon value: %v", it.Icon.Value)
	}

	if it.Icon.Type != "fileicon" {
		t.Fatalf("Incorrect type: %v", it.Icon.Type)
	}
}

var marshalItemTests = []struct {
	Item         *Item
	ExpectedJSON string
}{
	// Minimal item
	{Item: &Item{Title: "title"},
		ExpectedJSON: `{"title":"title"}`},
	// With UID
	{Item: &Item{Title: "title", UID: "xxx-yyy"},
		ExpectedJSON: `{"title":"title","uid":"xxx-yyy"}`},
	// With autocomplete
	{Item: &Item{Title: "title", Autocomplete: "xxx-yyy"},
		ExpectedJSON: `{"autocomplete":"xxx-yyy","title":"title"}`},
	// With empty autocomplete
	{Item: &Item{Title: "title", KeepEmptyAutocomplete: true},
		ExpectedJSON: `{"autocomplete":"","title":"title"}`},
	// With subtitle
	{Item: &Item{Title: "title", Subtitle: "subtitle"},
		ExpectedJSON: `{"title":"title","subtitle":"subtitle"}`},
	// Alternate subtitle
	{Item: &Item{Title: "title", Subtitle: "subtitle",
		Modifiers: map[string]*Modifier{
			"cmd": &Modifier{
				Key:         "cmd",
				subtitle:    "command sub",
				subtitleSet: true}}},
		ExpectedJSON: `{"title":"title","subtitle":"subtitle",` +
			`"mods":{"cmd":{"subtitle":"command sub"}}}`},
	// Valid item
	{Item: &Item{Title: "title", Valid: true},
		ExpectedJSON: `{"title":"title","valid":true}`},
	// With arg
	{Item: &Item{Title: "title", Arg: "arg1"},
		ExpectedJSON: `{"arg":"arg1","title":"title"}`},
	// Valid with arg
	{Item: &Item{Title: "title", Arg: "arg1", Valid: true},
		ExpectedJSON: `{"arg":"arg1","title":"title","valid":true}`},
	// With icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "icon.png", Type: ""}},
		ExpectedJSON: `{"title":"title","icon":{"path":"icon.png"}}`},
	// With file icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "icon.png", Type: "fileicon"}},
		ExpectedJSON: `{"title":"title","icon":{"path":"icon.png","type":"fileicon"}}`},
	// With filetype icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "public.folder", Type: "filetype"}},
		ExpectedJSON: `{"title":"title","icon":{"path":"public.folder","type":"filetype"}}`},
	// With type = file
	{Item: &Item{Title: "title", IsFile: true},
		ExpectedJSON: `{"type":"file","title":"title"}`},
	// With copy text
	{Item: &Item{Title: "title", Copytext: "copy"},
		ExpectedJSON: `{"text":{"copy":"copy"},"title":"title"}`},
	// With large text
	{Item: &Item{Title: "title", Largetext: "large"},
		ExpectedJSON: `{"text":{"largetype":"large"},"title":"title"}`},
	// With copy and large text
	{Item: &Item{Title: "title", Copytext: "copy", Largetext: "large"},
		ExpectedJSON: `{"text":{"copy":"copy","largetype":"large"},"title":"title"}`},
	// With arg and variable
	{Item: &Item{Title: "title", Arg: "value", Vars: map[string]string{"foo": "bar"}},
		ExpectedJSON: `{"arg":"{\"alfredworkflow\":{\"arg\":\"value\",\"variables\":{\"foo\":\"bar\"}}}","title":"title"}`},
}

var marshalModifierTests = []struct {
	Mod          *Modifier
	ExpectedJSON string
}{
	// Empty item (argSet=false)
	{Mod: &Modifier{arg: "title"},
		ExpectedJSON: `{}`},
	// With arg
	{Mod: &Modifier{arg: "title", argSet: true},
		ExpectedJSON: `{"arg":"title"}`},
	// With subtitle
	{Mod: &Modifier{subtitle: "sub here", subtitleSet: true},
		ExpectedJSON: `{"subtitle":"sub here"}`},
	// valid
	{Mod: &Modifier{valid: true, validSet: true},
		ExpectedJSON: `{"valid":true}`},
	// With all
	{Mod: &Modifier{
		arg: "title", argSet: true,
		subtitle: "sub here", subtitleSet: true,
		valid: true, validSet: true,
	},
		ExpectedJSON: `{"arg":"title","subtitle":"sub here","valid":true}`},
}

var marshalArgTests = []struct {
	Arg          *Arg
	ExpectedJSON string
}{
	// Only an arg
	{Arg: &Arg{},
		ExpectedJSON: `""`},
	// With arg
	{Arg: &Arg{arg: "title"},
		ExpectedJSON: `"title"`},
	// With variable
	{Arg: &Arg{vars: map[string]string{"foo": "bar"}},
		ExpectedJSON: `{"alfredworkflow":{"variables":{"foo":"bar"}}}`},
	// Multiple variables
	{Arg: &Arg{vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
		ExpectedJSON: `{"alfredworkflow":{"variables":{"ducky":"fuzz","foo":"bar"}}}`},
	// Multiple variables and arg (arg should be absent as argSet=false)
	{Arg: &Arg{arg: "title", vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
		ExpectedJSON: `{"alfredworkflow":{"variables":{"ducky":"fuzz","foo":"bar"}}}`},
	// Multiple variables and arg (arg should be present as argSet=true)
	{Arg: &Arg{arg: "title", argSet: true, vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
		ExpectedJSON: `{"alfredworkflow":{"arg":"title","variables":{"ducky":"fuzz","foo":"bar"}}}`},
}

func TestMarshalItem(t *testing.T) {
	for i, test := range marshalItemTests {
		data, err := json.Marshal(test.Item)
		if err != nil {
			t.Fatalf("#%d: marshal(%v): %v", i, test.Item, err)
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
			t.Fatalf("#%d: marshal(%v): %v", i, test.Mod, err)
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
			t.Fatalf("#%d: marshal(%v): %v", i, test.Arg, err)
			continue
		}

		if got, want := string(data), test.ExpectedJSON; got != want {
			t.Fatalf("#%d: got: %v wanted: %v", i, got, want)
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
	want = `{"items":[{"title":"item 1"}]}`
	fb.NewItem("item 1")

	got, err = json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error marshalling feedback: got: %s want: %s: %v",
			got, want, err)
	}

}

// TestModifiersInheritVars tests that Modifiers inherit variables from their
// parent Item
func TestModifiersInheritVars(t *testing.T) {
	fb := NewFeedback()
	it := fb.NewItem("title")
	it.SetVar("foo", "bar")
	m, err := it.NewModifier("cmd")
	if err != nil {
		t.Fatalf("Error creating modifier: %v", err)
	}
	if m.Var("foo") != "bar" {
		t.Fatalf("Modifier var has wrong value. Expected=bar, Received=%v", m.Var("foo"))
	}
}
