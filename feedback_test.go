// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItem_Icon(t *testing.T) {
	t.Parallel()

	it := Item{}
	it.Icon(&Icon{"first", "fileicon"})
	assert.Equal(t, "first", it.icon.Value, "unexpected icon Value")
	assert.Equal(t, IconType("fileicon"), it.icon.Type, "unexpected icon Type")
}

func p(s string) *string { return &s }

// Feedback is empty.
func TestFeedback_IsEmpty(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()
	assert.True(t, fb.IsEmpty(), "feedback not empty")

	fb.NewItem("test")
	assert.False(t, fb.IsEmpty(), "feedback empty")
}

func TestItem_MarshalJSON(t *testing.T) {
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
				"cmd": {
					Key:      "cmd",
					subtitle: p("command sub")}}},
			x: `{"title":"title","subtitle":"subtitle",` +
				`"valid":false,"mods":{"cmd":{"subtitle":"command sub"}}}`},
		// Valid item
		{in: &Item{title: "title", valid: true},
			x: `{"title":"title","valid":true}`},
		// With arg
		{in: &Item{title: "title", arg: []string{"arg1"}},
			x: `{"title":"title","arg":"arg1","valid":false}`},
		// Empty arg
		{in: &Item{title: "title", arg: []string{""}},
			x: `{"title":"title","arg":"","valid":false}`},
		// Multiple args
		{in: &Item{title: "title", arg: []string{"one", "two"}},
			x: `{"title":"title","arg":["one","two"],"valid":false}`},
		// Arg contains escapes
		{in: &Item{title: "title", arg: []string{"\x00arg\x00"}},
			x: `{"title":"title","arg":"\u0000arg\u0000","valid":false}`},
		// Valid with arg
		{in: &Item{title: "title", arg: []string{"arg1"}, valid: true},
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
		{in: &Item{title: "title", arg: []string{"value"}, vars: map[string]string{"foo": "bar"}},
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
			assert.Nil(t, err, "marshal Item failed")
			assert.Equal(t, td.x, string(data), "unexpected JSON")
		})
	}
}

func TestModifier_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in *Modifier
		x  string
	}{
		// Empty item
		{in: &Modifier{}, x: `{}`},
		// With arg
		{in: &Modifier{arg: []string{"title"}}, x: `{"arg":"title"}`},
		// Empty arg
		{in: &Modifier{arg: []string{""}}, x: `{"arg":""}`},
		// Multiple args
		{in: &Modifier{arg: []string{"one", "two"}}, x: `{"arg":["one","two"]}`},
		// With subtitle
		{in: &Modifier{subtitle: p("sub here")}, x: `{"subtitle":"sub here"}`},
		// valid
		{in: &Modifier{valid: true}, x: `{"valid":true}`},
		// icon
		{in: &Modifier{icon: &Icon{"icon.png", ""}}, x: `{"icon":{"path":"icon.png"}}`},
		// With all
		{in: &Modifier{
			arg:      []string{"title"},
			subtitle: p("sub here"),
			valid:    true,
		},
			x: `{"arg":"title","subtitle":"sub here","valid":true}`},
		// With variable
		{in: &Modifier{
			arg:      []string{"title"},
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
			assert.Nil(t, err, "marshal Modifier failed")
			assert.Equal(t, td.x, string(data), "unexpected JSON")
		})
	}
}

func TestArgVars_MarshalJSON(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		in *ArgVars
		x  string
	}{
		// Empty
		{in: &ArgVars{}, x: `""`},
		// With arg
		{in: &ArgVars{arg: []string{"title"}}, x: `"title"`},
		// With multiple args
		{in: &ArgVars{arg: []string{"one", "two"}},
			x: `{"alfredworkflow":{"arg":["one","two"]}}`},
		// With non-ASCII arg
		{in: &ArgVars{arg: []string{"fübär"}}, x: `"fübär"`},
		// With escapes
		{in: &ArgVars{arg: []string{"\x00"}}, x: `"\u0000"`},
		// With variable
		{in: &ArgVars{vars: map[string]string{"foo": "bar"}},
			x: `{"alfredworkflow":{"variables":{"foo":"bar"}}}`},
		// Multiple variables
		{in: &ArgVars{vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
			x: `{"alfredworkflow":{"variables":{"ducky":"fuzz","foo":"bar"}}}`},
		// Multiple variables and arg
		{in: &ArgVars{arg: []string{"title"}, vars: map[string]string{"foo": "bar", "ducky": "fuzz"}},
			x: `{"alfredworkflow":{"arg":"title","variables":{"ducky":"fuzz","foo":"bar"}}}`},
	}

	for i, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("MarshalArgVar(%d)", i), func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(td.in)
			assert.Nil(t, err, "marshal ArgVars failed")
			assert.Equal(t, td.x, string(data), "unexpected JSON")
		})
	}
}

// Simple arg marshalled to single string
func TestArgVars_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in *ArgVars
		x  string
	}{
		// Empty
		{in: &ArgVars{}, x: ""},
		// With arg
		{in: &ArgVars{arg: []string{"title"}}, x: "title"},
		// With multiple args
		{in: &ArgVars{arg: []string{"one", "two"}},
			x: `{"alfredworkflow":{"arg":["one","two"]}}`},
		// With non-ASCII
		{in: &ArgVars{arg: []string{"fübär"}},
			x: "fübär"},
		// With escapes
		{in: &ArgVars{arg: []string{"\x00"}},
			x: "\x00"},
	}

	for i, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("StringifyArg(%d)", i), func(t *testing.T) {
			t.Parallel()
			v, err := td.in.String()
			assert.Nil(t, err, "stringify ArgVars failed")
			assert.Equal(t, td.x, v, "unexpected value")
		})
	}
}

// Vars set correctly
func TestArgVars_Vars(t *testing.T) {
	t.Parallel()

	vars := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
		"key4": "val4",
		"key5": "val5",
	}

	av := NewArgVars()
	for k, v := range vars {
		av.Var(k, v)
	}

	assert.Equal(t, vars, av.Vars(), "Unexpected Vars")
}

// Marshal Feedback to JSON
func TestFeedback_MarshalJSON(t *testing.T) {
	t.Parallel()

	// Empty feedback
	fb := NewFeedback()
	want := `{"items":[]}`
	got, err := json.Marshal(fb)
	assert.Nil(t, err, "marshal Feedback failed")
	assert.Equal(t, string(got), want, "unexpected value")

	// Feedback with item
	want = `{"items":[{"title":"item 1","valid":false}]}`
	fb.NewItem("item 1")

	got, err = json.Marshal(fb)
	assert.Nil(t, err, "marshal Feedback failed")
	assert.Equal(t, string(got), want, "unexpected value")
}

// Modifier inherits variables from parent Item
func TestModifierInheritVars(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()
	it := fb.NewItem("title")
	it.Var("foo", "bar")
	m := it.NewModifier("cmd")
	assert.Equal(t, "bar", m.Vars()["foo"], "unexpected var value")
}

// Empty/invalid modifiers
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
		td := td
		t.Run(fmt.Sprintf("%v", td.keys), func(t *testing.T) {
			it := fb.NewItem("title")
			require.Equal(t, 0, len(it.mods), "unexpected modifier count")

			_ = it.NewModifier(td.keys...)
			if td.ok {
				assert.Equal(t, 1, len(it.mods), "unexpected modifier count")
			} else {
				assert.Equal(t, 0, len(it.mods), "unexpected modifier count")
			}
		})
	}
}

// Combined modifiers
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
		td := td
		t.Run(fmt.Sprintf("%v", td.keys), func(t *testing.T) {
			m := it.NewModifier(td.keys...)
			assert.Equal(t, td.x, string(m.Key), "unexpected modifier")
		})
	}
}

// Modifier creation shortcut methods
func TestModifierShortcuts(t *testing.T) {
	t.Parallel()

	it := &Item{}
	tests := []struct {
		m *Modifier
		k ModKey
	}{
		{it.Cmd(), ModCmd},
		{it.Opt(), ModOpt},
		{it.Shift(), ModShift},
		{it.Ctrl(), ModCtrl},
		{it.Fn(), ModFn},
	}

	for _, td := range tests {
		assert.Equal(t, td.k, td.m.Key, "Bad ModKey for %q", td.k)
	}
}

// TestFeedback_Rerun verifies that rerun is properly set.
func TestFeedback_Rerun(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()

	fb.Rerun(1.5)

	want := `{"rerun":1.5,"items":[]}`
	got, err := json.Marshal(fb)
	assert.Nil(t, err, "marshal Feedback failed")
	assert.Equal(t, string(got), want, "unexpected value")
}

// Vars are properly inherited by Items and Modifiers
func TestFeedback_Vars(t *testing.T) {
	t.Parallel()

	fb := NewFeedback()

	fb.Var("foo", "bar")
	if fb.Vars()["foo"] != "bar" {
		t.Fatalf("Feedback var has wrong value. Expected=bar, Received=%v", fb.Vars()["foo"])
	}

	want := `{"variables":{"foo":"bar"},"items":[]}`
	got, err := json.Marshal(fb)
	assert.Nil(t, err, "marshal Feedback failed")
	assert.Equal(t, string(got), want, "unexpected value")

	// Top-level vars are inherited
	it := fb.NewItem("title")
	assert.Equal(t, "bar", it.Vars()["foo"], "unexpected var")

	// Modifier inherits Item and top-level vars
	it.Var("baz", "qux")
	m := it.NewModifier("cmd")
	assert.Equal(t, "qux", m.Vars()["baz"], "unexpected var")
	assert.Equal(t, "bar", m.Vars()["foo"], "unexpected var")
}

// Item methods set fields correctly
func TestItem_methods(t *testing.T) {
	t.Parallel()

	var (
		title        = "title"
		subtitle     = "subtitle"
		match        = "match"
		uid          = "uid"
		autocomplete = "autocomplete"
		arg          = []string{"arg"}
		valid        = true
		copytext     = "copytext"
		largetype    = "largetype"
		qlURL        = "http://www.example.com"
	)

	it := &Item{}

	assert.Equal(t, "", it.title, "Non-empty title")
	assert.Nil(t, it.subtitle, "Non-nil subtitle")
	assert.Nil(t, it.match, "Non-nil match")
	assert.Nil(t, it.uid, "Non-nil UID")
	assert.Nil(t, it.autocomplete, "Non-nil autocomplete")
	assert.Nil(t, it.arg, "Non-nil arg")
	assert.Nil(t, it.copytext, "Non-nil copytext")
	assert.Nil(t, it.largetype, "Non-nil largetype")
	assert.Nil(t, it.ql, "Non-nil quicklook")

	it.Title(title).
		Subtitle(subtitle).
		Match(match).
		UID(uid).
		Autocomplete(autocomplete).
		Arg(arg...).
		Valid(valid).
		Copytext(copytext).
		Largetype(largetype).
		Quicklook(qlURL)

	assert.Equal(t, title, it.title, "Bad title")
	assert.Equal(t, subtitle, *it.subtitle, "Bad subtitle")
	assert.Equal(t, match, *it.match, "Bad match")
	assert.Equal(t, uid, *it.uid, "Bad UID")
	assert.Equal(t, autocomplete, *it.autocomplete, "Bad autocomplete")
	assert.Equal(t, arg, it.arg, "Bad arg")
	assert.Equal(t, valid, valid, "Bad valid")
	assert.Equal(t, copytext, *it.copytext, "Bad copytext")
	assert.Equal(t, largetype, *it.largetype, "Bad largetext")
	assert.Equal(t, qlURL, *it.ql, "Bad quicklook URL")
}

func TestModifier_methods(t *testing.T) {
	var (
		key      = ModCmd
		arg      = []string{"arg"}
		subtitle = "subtitle"
		valid    = true
		icon     = IconAccount
	)

	m := &Modifier{}
	assert.Equal(t, ModKey(""), m.Key, "Non-empty key")
	assert.Nil(t, m.arg, "Non-nil arg")
	assert.Nil(t, m.subtitle, "Non-nil subtitle")
	assert.False(t, m.valid, "Bad valid")
	assert.Nil(t, m.icon, "Bad icon")

	m.Key = key
	m.Subtitle(subtitle).
		Arg(arg...).
		Valid(valid).
		Icon(icon)

	assert.Equal(t, key, m.Key, "Bad key")
	assert.Equal(t, arg, m.arg, "Bad arg")
	assert.Equal(t, subtitle, *m.subtitle, "Bad subtitle")
	assert.Equal(t, valid, m.valid, "Bad valid")
	assert.Equal(t, icon.Type, m.icon.Type, "Bad icon type")
	assert.Equal(t, icon.Value, m.icon.Value, "Bad icon value")
}

// Sorts Feedback.Items
func TestFeedback_Sort(t *testing.T) {
	for _, td := range feedbackTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		r := fb.Sort(td.q)
		for i, it := range fb.Items {
			assert.Equal(t, td.out[i], it.title, "unexpected title")
			assert.Equal(t, td.m[i], r[i].Match, "unexpected match")
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
	// sorting is stable
	{
		q:   "test",
		in:  []string{"test #1", "test #10", "test #2", "test #3"},
		out: []string{"test #1", "test #2", "test #3", "test #10"},
		m:   []bool{true, true, true, true},
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

// Filter Feedback.Items
func TestFeedback_Filter(t *testing.T) {
	for _, td := range filterTitles {
		fb := NewFeedback()
		for _, s := range td.in {
			fb.NewItem(s)
		}
		fb.Filter(td.q)
		assert.Equal(t, len(td.out), len(fb.Items), "unexpected result count")
		for i, it := range fb.Items {
			assert.Equal(t, td.out[i], it.title, "unexpected title")
		}
	}
}
