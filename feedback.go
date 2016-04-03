package workflow

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	validModifiers = []string{"cmd", "opt", "ctrl", "shift", "fn"}
)

// Item is a single Alfred result. Add them to a Feedback struct to
// generate valid XML.
type Item struct {
	// Result title (only required field)
	Title string `xml:"title"`

	// Result subtitle
	Subtitle string `xml:"subtitle,omitempty"`

	// Custom subtitles for when modifier keys are held
	AlternateSubtitles []Subtitle

	// What the query will expand to when the user TABs it (or hits
	// RETURN on an invalid result)
	Autocomplete string `xml:"autocomplete,attr"`

	// If true, send autocomplete="" to Alfred. If autocomplete is not
	// specified, TAB will do nothing. If autocomplete is an empty
	// string, TAB will autocomplete to an empty string, i.e. Alfred's
	// query will be deleted.
	KeepEmptyAutocomplete bool

	// Used by Alfred to remember your choices. Use blank string
	// to force results to appear in the order you generate them.
	UID string `xml:"uid,attr,omitempty"`

	// The value that is passed as {query} to the next action in the workflow
	Arg string `xml:"arg,omitempty"`

	// Whether the result is "actionable". Must be set to "yes" or "no".
	// Use SetValid() to set from a boolean value.
	Valid string `xml:"valid,attr,omitempty"`

	// The type of the result. Currently, "file" is the only value Alfred
	// understands. If set to "file" and Arg is a valid filepath, user
	// can use Alfred's File Actions on the item.
	Type string `xml:"type,attr,omitempty"`

	// The icon for the result. Can point to an image file, a filepath
	// of a file whose icon should be used, or a UTI, such as
	// "com.apple.folder".
	Icon ItemIcon `xml:",omitempty"`

	// Name of the XML tag in the output. Don't fuck with this.
	XMLName xml.Name `xml:"item"`
}

// SetAlternateSubtitle sets custom subtitles for modifier keys.
// `modifier` must be one of "cmd", "opt", "ctrl", "shift", "fn".
func (it *Item) SetAlternateSubtitle(modifier string, value string) error {
	modifier = strings.ToLower(modifier)
	valid := false
	for _, m := range validModifiers {
		if modifier == m {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("modifier must be one of %v not %v", validModifiers, modifier)
	}
	sub := Subtitle{}
	sub.Value = value
	sub.Modifier = modifier
	it.AlternateSubtitles = append(it.AlternateSubtitles, sub)
	return nil
}

// SetIcon sets the icon for a result item.
// Pass "" for kind if value is the path to an icon file.
func (it *Item) SetIcon(value string, kind string) {
	if kind != "" && kind != "fileicon" && kind != "filetype" {
		log.Printf(
			"Icon kind must be \"fileicon\", \"filetype\" or \"\", not %v",
			kind)

	}
	it.Icon.Value = value
	it.Icon.Type = kind
}

// SetValid sets Valid using a boolean.
// The actual value must be the string "yes" or "no"
func (it *Item) SetValid(value bool) {
	if value == true {
		it.Valid = "YES"
	} else {
		it.Valid = "NO"
	}
}

// ItemIcon represents the icon for an Item.
//
// Alfred supports PNG or ICNS files, UTIs (e.g. "public.folder") or
// can use the icon of a specified file (e.g. "/Applications/Safari.app"
// to use Safari's icon.
//
// Type = "" (the default) will treat Value as the path to a PNG or ICNS
// file.
//
// Type = "fileicon" will treat Value as the path to a file or directory
// and use that file's icon, e.g:
//
//    icon := ItemIcon{"/Applications/Mail.app", "fileicon"}
//
// will display Mail.app's icon.
//
// Type = "filetype" will treat Value as a UTI, such as "public.movie"
// or "com.microsoft.word.doc". UTIs are useful when you don't have
// a local path to point to.
//
// You can find out the UTI of a filetype by dragging one of the files
// to a File Filter's File Types list in Alfred, or in a shell with:
//
//    mdls /path/to/the/file
type ItemIcon struct {
	Value   string   `xml:",chardata"`
	Type    string   `xml:"type,attr,omitempty"`
	XMLName xml.Name `xml:"icon,omitempty"`
}

// Subtitle is a custom subtitle for when a modifier key is pressed.
type Subtitle struct {
	Value    string   `xml:",chardata"`
	Modifier string   `xml:"mod,attr"`
	XMLName  xml.Name `xml:"subtitle,omitempty"`
}

// Feedback contains Items. This is the top-level object for generating
// Alfred XML (i.e. serialise this and send it to Alfred).
type Feedback struct {
	Items   []*Item
	XMLName xml.Name `xml:"items"`
}

// NewItem adds a new Item and returns a pointer to it.
func (fb *Feedback) NewItem() *Item {
	item := Item{}
	item.Icon = ItemIcon{}
	fb.Items = append(fb.Items, &item)
	return &item
}

// NewFileItem adds and returns a pointer to a new item pre-populated from path.
// Title is the base name of the file
// Subtitle is the path to the file (using "~" for $HOME)
// Valid is "YES"
// UID, Arg and Autocomplete are set to path
// Type is "file"
// Icon is the icon of the file at path
func (fb *Feedback) NewFileItem(path string) *Item {
	item := fb.NewItem()
	item.Title = filepath.Base(path)
	item.Subtitle = shortenPath(path)
	item.Arg = path
	item.Valid = "YES"
	item.UID = path
	item.Autocomplete = path
	item.Type = "file"
	item.SetIcon(path, "fileicon")
	return item
}

// Send generates XML from this struct and sends it to Alfred.
func (fb *Feedback) Send() error {
	// fb2 := Feedback{}
	// for _, it := range fb.Items {
	// 	if it.Autocomplete != "" || it.KeepEmptyAutocomplete == false {
	// 		fb2.Items = append(fb2.Items, it)
	// 	} else {
	// 		a := &ItemAlias{Item: it, Autocomplete: it.Autocomplete}
	// 		// TODO: Use different struct
	// 		fb2.Items = append(fb2.Items, a)
	// 	}
	// }
	output, err := xml.MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("Error generating XML : %v", err)
	}
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
	return nil
}

func init() {
}

// shortenPath replaces $HOME with ~ in path
func shortenPath(path string) string {
	return strings.Replace(path, os.Getenv("HOME"), "~", -1)
}
