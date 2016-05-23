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

var marshalTests = []struct {
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
		AlternateSubtitles: map[string]string{"cmd": "command sub"}},
		ExpectedJSON: `{"title":"title","subtitle":"subtitle",` +
			`"mods":{"cmd":"command sub"}}`},
	// Valid item
	{Item: &Item{Title: "title", Valid: true},
		ExpectedJSON: `{"title":"title","valid":true}`},
	// With arg
	{Item: &Item{Title: "title", Arg: "arg1"},
		ExpectedJSON: `{"title":"title","arg":"arg1"}`},
	// Valid with arg
	{Item: &Item{Title: "title", Arg: "arg1", Valid: true},
		ExpectedJSON: `{"title":"title","arg":"arg1","valid":true}`},
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
	// TODO: copytext
	// TODO: largetext
}

func TestMarshalItem(t *testing.T) {
	for i, test := range marshalTests {
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
	it := fb.NewItem()
	it.Title = "item 1"

	got, err = json.Marshal(fb)
	if err != nil {
		t.Fatalf("Error marshalling feedback: got: %s want: %s: %v",
			got, want, err)
	}

}
