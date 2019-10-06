// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSetIcon(t *testing.T) {
	t.Parallel()

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

// TestEmpty asserts feedback is empty.
func TestEmpty(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	tests := []struct {
		in *Item
		x  string
	}{
		// Minimal item
		{in: &Item{title: "title"},
			x: `{"title":"title","valid":false}`},
		// With UID
		{in: &Item{title: "title", uid: p("xxx-yyy")},
			x: `{"title":"title","uid":"xxx-yyy","valid":false}`},
		// With autocomplete
		{in: &Item{title: "title", autocomplete: p("xxx-yyy")},
			x: `{"title":"title","autocomplete":"xxx-yyy","valid":false}`},
		// With empty autocomplete
		{in: &Item{title: "title", autocomplete: p("")},
			x: `{"title":"title","autocomplete":"","valid":false}`},
		// With subtitle
		{in: &Item{title: "title", subtitle: p("subtitle")},
			x: `{"title":"title","subtitle":"subtitle","valid":false}`},
		// Alternate subtitle
		{in: &Item{title: "title", subtitle: p("subtitle"),
			mods: map[ModKey]*Modifier{
				"cmd": &Modifier{
					Key:      "cmd",
					subtitle: p("command sub")}}},
			x: `{"title":"title","subtitle":"subtitle",` +
				`"valid":false,"mods":{"cmd":{"subtitle":"command sub"}}}`},
		// Valid item
		{in: &Item{title: "title", valid: true},
			x: `{"title":"title","valid":true}`},
		// With arg
		{in: &Item{title: "title", arg: p("arg1")},
			x: `{"title":"title","arg":"arg1","valid":false}`},
		// Empty arg
		{in: &Item{title: "title", arg: p("")},
			x: `{"title":"title","arg":"","valid":false}`},
		// Arg contains escapes
		{in: &Item{title: "title", arg: p("\x00arg\x00")},
			x: `{"title":"title","arg":"\u0000arg\u0000","valid":false}`},
		// Valid with arg
		{in: &Item{title: "title", arg: p("arg1"), valid: true},
			x: `{"title":"title","arg":"arg1","valid":true}`},
		// With icon
		{in: &Item{title: "title",
			icon: &Icon{Value: "icon.png", Type: ""}},
			x: `{"title":"title","valid":false,"icon":{"path":"icon.png"}}`},
		// With file icon
		{in: &Item{title: "title",
			icon: &Icon{Value: "icon.png", Type: "fileicon"}},
			x: `{"title":"title","valid":false,"icon":{"path":"icon.png","type":"fileicon"}}`},
		// With filetype icon
		{in: &Item{title: "title",
			icon: &Icon{Value: "public.folder", Type: "filetype"}},
			x: `{"title":"title","valid":false,"icon":{"path":"public.folder","type":"filetype"}}`},
		// With type = file
		{in: &Item{title: "title", file: true},
			x: `{"title":"title","valid":false,"type":"file"}`},
		// With copy text
		{in: &Item{title: "title", copytext: p("copy")},
			x: `{"title":"title","valid":false,"text":{"copy":"copy"}}`},
		// With large text
		{in: &Item{title: "title", largetype: p("large")},
			x: `{"title":"title","valid":false,"text":{"largetype":"large"}}`},
		// With copy and large text
		{in: &Item{title: "title", copytext: p("copy"), largetype: p("large")},
			x: `{"title":"title","valid":false,"text":{"copy":"copy","largetype":"large"}}`},
		// With arg and variable
		{in: &Item{title: "title", arg: p("value"), vars: map[string]string{"foo": "bar"}},
			x: `{"title":"title","arg":"value","valid":false,"variables":{"foo":"bar"}}`},
		// With match
		{in: &Item{title: "title", match: p("one two three")},
			x: `{"title":"title","match":"one two three","valid":false}`},
		// With quicklook
		{in: &Item{title: "title", ql: p("http://www.example.com")},
			x: `{"title":"title","valid":false,"quicklookurl":"http://www.example.com"}`},
	}

	for i, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("MarshalItem(%d)", i), func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(td.in)
			if err != nil {
				t.Fatalf("[ERROR] %v", err)
			}

			if v := string(data); v != td.x {
				t.Errorf("Expected=%v, Got=%v", td.x, v)
			}
		})
	}
}

func TestMarshalModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in *Modifier
		x  string
	}{
		// Empty item
		{in: &Modifier{}, x: `{}`},
		// With arg
		{in: &Modifier{arg: p("title")}, x: `{"arg":"title"}`},
		// Empty arg
		{in: &Modifier{arg: p("")}, x: `{"arg":""}`},
		// With subtitle
		{in: &Modifier{subtitle: p("sub here")}, x: `{"subtitle":"sub here"}`},
		// valid
		{in: &Modifier{valid: true}, x: `{"valid":true}`},
		// icon
		{in: &Modifier{icon: &Icon{"icon.png", ""}}, x: `{"icon":{"path":"icon.png"}}`},
		// With all
		{in: &Modifier{
			arg:      p("title"),
			subtitle: p("sub here"),
			valid:    true,
		},
			x: `{"arg":"title","subtitle":"sub here","valid":true}`},
		// With variable
		{in: &Modifier{
			arg:      p("title"),
			subtitle: p("sub here"),
			valid:    true,
			vars:     map[string]string{"foo": "bar"},
		},
			x: `{"arg":"title","subtitle":"sub here","valid":true,"variables":{"foo":"bar"}}`},
	}

	for i, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("MarshalModifier(%d)", i), func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(td.in)
			if err != nil {
				t.Fatalf("[ERROR] %v", err)
			}

			if v := string(data); v != td.x {
				t.Errorf("Expected=%v, Got=%v", td.x, v)
			}
		})
	}
}

func TestMarshalArg(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		in *ArgVars
		x  string
	}{
		// Empty
		{in: &ArgVars{},
			x: `""`},
		// With arg
		{in: &ArgVars{arg: p("title")},
			x: `"title"`},
		// With non-ASCII arg
		{in: &ArgVars{arg: p("fübär")},
			x: `"fübär"`},
		// With escapes
		{in: &ArgVars{arg: p("\x00")},
			x: `"\u0000"`},
		// With variable
		{in: &ArgVars{vars: map[string]string{"foo": "bar"}},
			x: `{"alfredworkflow":{"variables":{"foo":"bar"}}}`},
		// Multiple variables
		{in: &ArgVars{vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
			x: `{"alfredworkflow":{"variables":{"ducky":"fuzz","foo":"bar"}}}`},
		// Multiple variables and arg
		{in: &ArgVars{arg: p("title"), vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
			x: `{"alfredworkflow":{"arg":"title","variables":{"ducky":"fuzz","foo":"bar"}}}`},
	}

	for i, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("MarshalArgVar(%d)", i), func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(td.in)
			if err != nil {
				t.Fatalf("[ERROR] %v", err)
			}

			if v := string(data); v != td.x {
				t.Errorf("Expected=%v, Got=%v", td.x, v)
			}
		})
	}
}

func TestStringifyArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in *ArgVars
		x  string
	}{
		// Empty
		{in: &ArgVars{},
			x: ""},
		// With arg
		{in: &ArgVars{arg: p("title")},
			x: "title"},
		// With non-ASCII
		{in: &ArgVars{arg: p("fübär")},
			x: "fübär"},
		// With escapes
		{in: &ArgVars{arg: p("\x00")},
			x: "\x00"},
	}

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("StringifyArg(%#v)", td.in), func(t *testing.T) {
			t.Parallel()
			v, err := td.in.String()
			if err != nil {
				t.Fatalf("[ERROR] %v", err)
			}
			if v != td.x {
				t.Errorf("Expected=%q, Got=%q", td.x, v)
			}
		})
	}
}

func TestMarshalFeedback(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	fb := NewFeedback()
	it := fb.NewItem("title")
	it.Var("foo", "bar")
	m := it.NewModifier("cmd")

	if m.Vars()["foo"] != "bar" {
		t.Fatalf("Modifier var has wrong value. Expected=bar, Received=%v", m.Vars()["foo"])
	}
}

func TestEmptyModifiersIgnored(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()

	tests := []struct {
		keys []ModKey
		ok   bool
	}{
		{[]ModKey{}, false},
		{[]ModKey{""}, false},
		{[]ModKey{"", ""}, false},
		{[]ModKey{"rick flair"}, false},
		{[]ModKey{"andre the giant", ""}, false},
		{[]ModKey{"ultimate warrior", "cmd"}, true},
		{[]ModKey{"ctrl", "", "giant haystacks"}, true},
	}

	for _, td := range tests {
		it := fb.NewItem("title")
		v := len(it.mods)
		if v != 0 {
			t.Fatalf("Unexpected modifiers: %+v", it.mods)
		}
		_ = it.NewModifier(td.keys...)
		v = len(it.mods)
		if td.ok {
			if v != 1 {
				t.Errorf("Good mod %+v not accepted", td.keys)
			}
		} else {
			if v != 0 {
				t.Errorf("Bad mod %+v accepted", td.keys)
			}
		}
	}
}

func TestMultipleModifiers(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()
	it := fb.NewItem("title")

	tests := []struct {
		keys []ModKey
		x    string
	}{
		{[]ModKey{"cmd"}, "cmd"},
		{[]ModKey{"alt"}, "alt"},
		{[]ModKey{"opt"}, "alt"},
		{[]ModKey{"fn"}, "fn"},
		{[]ModKey{"shift"}, "shift"},
		{[]ModKey{"alt", "cmd"}, "alt+cmd"},
		{[]ModKey{"cmd", "alt"}, "alt+cmd"},
		{[]ModKey{"cmd", "opt"}, "alt+cmd"},
		{[]ModKey{"cmd", "opt", "ctrl"}, "alt+cmd+ctrl"},
		{[]ModKey{"cmd", "opt", "shift"}, "alt+cmd+shift"},
		// invalid keys ignored
		{[]ModKey{}, ""},
		{[]ModKey{""}, ""},
		{[]ModKey{"shift", "cmd", ""}, "cmd+shift"},
		{[]ModKey{"shift", "ctrl", "hulk hogan"}, "ctrl+shift"},
		{[]ModKey{"shift", "undertaker", "cmd", ""}, "cmd+shift"},
	}

	for _, td := range tests {
		m := it.NewModifier(td.keys...)
		v := string(m.Key)
		if v != td.x {
			t.Errorf("Bad Modifier for %#v. Expected=%q, Got=%q", td.keys, td.x, v)
		}
	}
}

// TestFeedbackRerun verifies that rerun is properly set.
func TestFeedbackRerun(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

	// Top-level vars are inherited
	it := fb.NewItem("title")
	if it.Vars()["foo"] != "bar" {
		t.Fatalf("Item var has wrong value. Expected=bar, Received=%v", it.Vars()["foo"])
	}

	// Modifier inherits Item and top-level vars
	it.Var("baz", "qux")
	m := it.NewModifier("cmd")
	if m.Vars()["baz"] != "qux" {
		t.Fatalf("Modifier var has wrong value. Expected=qux, Received=%v", m.Vars()["baz"])
	}
	if m.Vars()["foo"] != "bar" {
		t.Fatalf("Modifier var has wrong value. Expected=bar, Received=%v", m.Vars()["foo"])
	}
}

// TestSortFeedback sorts Feedback.Items
func TestSortFeedback(t *testing.T) {

	for _, td := range feedbackTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		r := fb.Sort(td.q)
		for i, it := range fb.Items {
			if it.title != td.out[i] {
				t.Errorf("query=%#v, pos=%d, expected=%s, got=%s", td.q, i+1, td.out[i], it.title)
			}
			if r[i].Match != td.m[i] {
				t.Errorf("query=%#v, keywords=%#v, expected=%v, got=%v", td.q, it.title, td.m[i], r[i].Match)
			}
		}
	}
}

var feedbackTitles = []struct {
	q   string
	in  []string
	out []string
	m   []bool
}{
	{
		q:   "got",
		in:  []string{"game of thrones", "no match", "got milk?", "got"},
		out: []string{"got", "game of thrones", "got milk?", "no match"},
		m:   []bool{true, true, true, false},
	},
	{
		q:   "of",
		in:  []string{"out of time", "spelunking", "OmniFocus", "game of thrones"},
		out: []string{"OmniFocus", "out of time", "game of thrones", "spelunking"},
		m:   []bool{true, true, true, false},
	},
	{
		q:   "safa",
		in:  []string{"see all fellows' armpits", "Safari", "french canada", "spanish harlem"},
		out: []string{"Safari", "see all fellows' armpits", "spanish harlem", "french canada"},
		m:   []bool{true, true, false, false},
	},
}

var filterTitles = []struct {
	q   string
	in  []string
	out []string
}{
	{
		q:   "got",
		in:  []string{"game of thrones", "no match", "got milk?", "got"},
		out: []string{"got", "game of thrones", "got milk?"},
	},
	{
		q:   "of",
		in:  []string{"out of time", "spelunking", "OmniFocus", "game of thrones"},
		out: []string{"OmniFocus", "out of time", "game of thrones"},
	},
	{
		q:   "safa",
		in:  []string{"see all fellows' armpits", "Safari", "french canada", "spanish harlem"},
		out: []string{"Safari", "see all fellows' armpits"},
	},
}

// TestFilterFeedback filters Feedback.Items
func TestFilterFeedback(t *testing.T) {
	for _, td := range filterTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		fb.Filter(td.q)
		if len(fb.Items) != len(td.out) {
			t.Errorf("query=%#v, expected %d results, got %d", td.q, len(td.out), len(fb.Items))
		}
		for i, it := range fb.Items {
			if it.title != td.out[i] {
				t.Errorf("query=%#v, pos=%d, expected=%s, got=%s", td.q, i+1, td.out[i], it.title)
			}
		}
	}
}
