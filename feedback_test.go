package workflow

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileItem(t *testing.T) {
	ipPath := filepath.Join(WorkflowDir(), "info.plist")
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
	Item        *Item
	ExpectedXML string
}{
	// Minimal item
	{Item: &Item{Title: "title"},
		ExpectedXML: `<item valid="no"><title>title</title></item>`},
	// With UID
	{Item: &Item{Title: "title", UID: "xxx-yyy"},
		ExpectedXML: `<item uid="xxx-yyy" valid="no">` +
			`<title>title</title></item>`},
	// With autocomplete
	{Item: &Item{Title: "title", Autocomplete: "xxx-yyy"},
		ExpectedXML: `<item autocomplete="xxx-yyy" valid="no">` +
			`<title>title</title></item>`},
	// With empty autocomplete
	{Item: &Item{Title: "title", KeepEmptyAutocomplete: true},
		ExpectedXML: `<item autocomplete="" valid="no">` +
			`<title>title</title></item>`},
	// With subtitle
	{Item: &Item{Title: "title", Subtitle: "subtitle"},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<subtitle>subtitle</subtitle>` +
			`</item>`},
	// Alternate subtitle
	{Item: &Item{Title: "title", Subtitle: "subtitle",
		AlternateSubtitles: map[string]string{"cmd": "command sub"}},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<subtitle>subtitle</subtitle>` +
			`<subtitle mod="cmd">command sub</subtitle>` +
			`</item>`},
	// Valid item
	{Item: &Item{Title: "title", Valid: true},
		ExpectedXML: `<item valid="yes"><title>title</title></item>`},
	// With arg
	{Item: &Item{Title: "title", Arg: "arg1"},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<arg>arg1</arg></item>`},
	// Valid with arg
	{Item: &Item{Title: "title", Arg: "arg1", Valid: true},
		ExpectedXML: `<item valid="yes">` +
			`<title>title</title>` +
			`<arg>arg1</arg></item>`},
	// With icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "icon.png", Type: ""}},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<icon>icon.png</icon>` +
			`</item>`},
	// With file icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "icon.png", Type: "fileicon"}},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<icon type="fileicon">icon.png</icon>` +
			`</item>`},
	// With filetype icon
	{Item: &Item{Title: "title",
		Icon: &ItemIcon{Value: "public.folder", Type: "filetype"}},
		ExpectedXML: `<item valid="no">` +
			`<title>title</title>` +
			`<icon type="filetype">public.folder</icon>` +
			`</item>`},
	// TODO: copytext
	// TODO: largetext
}

func TestMarshalItem(t *testing.T) {
	for i, test := range marshalTests {
		data, err := xml.Marshal(test.Item)
		if err != nil {
			t.Errorf("#%d: marshal(%v): %v", i, test.Item, err)
			continue
		}

		if got, want := string(data), test.ExpectedXML; got != want {
			t.Errorf("#%d: got: %v wanted: %v", i, got, want)
		}
	}
}

func TestMarshalFeedback(t *testing.T) {
	// Empty feedback
	fb := Feedback{}
	want := `<items></items>`
	got, err := xml.Marshal(fb)
	if err != nil {
		t.Errorf("Error marshalling feedback: got: %v want: %v: %v",
			got, want, err)
	}
	if string(got) != want {
		t.Errorf("Incorrect feedback: got: %v, wanted: %v", got, want)
	}

	// Feedback with item
	want = `<items><item valid="no"><title>item 1</title></item></items>`
	it := fb.NewItem()
	it.Title = "item 1"

	got, err = xml.Marshal(fb)
	if err != nil {
		t.Errorf("Error marshalling feedback: got: %v want: %v: %v",
			got, want, err)
	}

}
