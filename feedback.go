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
	// Result item
	Title string `xml:"title"`

	// Result subtitle
	Subtitle string `xml:"subtitle,omitempty"`

	// Custom subtitles for when modifier keys are held
	AlternateSubtitles []Subtitle

	// What the query will expand to when the user TABs it (or hits
	// RETURN on an invalid result)
	Autocomplete string `xml:"autocomplete,attr"`

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

// SetSubtitle sets custom subtitles for modifier keys.
// `modifier` must be one of "cmd", "opt", "ctrl", "shift", "fn".
func (this *Item) SetSubtitle(modifier string, value string) error {
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
	this.AlternateSubtitles = append(this.AlternateSubtitles, sub)
	return nil
}

// SetIcon sets the icon for a result item.
// Pass "" for `kind` if `value` is the path to an icon file.
func (this *Item) SetIcon(value string, kind string) {
	if kind != "" && kind != "fileicon" && kind != "filetype" {
		log.Printf(
			"Icon kind must be \"fileicon\", \"filetype\" or \"\", not %v",
			kind)

	}
	this.Icon.Value = value
	this.Icon.Type = kind
}

// SetValid sets Valid using a boolean.
// The actual value must be the string "yes" or "no"
func (this *Item) SetValid(value bool) {
	if value == true {
		this.Valid = "YES"
	} else {
		this.Valid = "NO"
	}
}

// ItemIcon represents the icon for an Item.
// Type must be "fileicon", "filetype" or ""
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
func (this *Feedback) NewItem() *Item {
	item := Item{}
	item.Icon = ItemIcon{}
	this.Items = append(this.Items, &item)
	return &item
}

// NewFileItem adds and returns a pointer to a new item pre-populated from path.
// Title is the base name of the file
// Subtitle is the path to the file (using "~" for $HOME)
// Valid is "YES"
// UID, Arg and Autocomplete are set to path
// Type is "file"
// Icon is the icon of the file at path
func (this *Feedback) NewFileItem(path string) *Item {
	item := this.NewItem()
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
func (this *Feedback) Send() error {
	output, err := xml.MarshalIndent(this, "", "  ")
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
