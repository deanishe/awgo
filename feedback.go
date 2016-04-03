package workflow

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogs.deanishe.net/deanishe/awgo/util"
)

var (
	validModifiers = map[string]bool{
		"cmd":   true,
		"opt":   true,
		"ctrl":  true,
		"shift": true,
		"fn":    true,
	}
	validIconTypes = map[string]bool{
		"filetype": true,
		"fileicon": true,
		"":         true,
	}
)

// Item is a single Alfred result. Add them to a Feedback struct to
// generate valid XML.
type Item struct {
	// Result title (only required field)
	Title string

	// Result subtitle
	Subtitle string

	// Custom subtitles for when modifier keys are held
	AlternateSubtitles map[string]string
	// AlternateSubtitles []Subtitle

	// The value that is passed as {query} to the next action in the workflow
	Arg string

	// Used by Alfred to remember your choices. Use blank string
	// to force results to appear in the order you generate them.
	UID string

	// What the query will expand to when the user TABs it (or hits
	// RETURN on an invalid result)
	Autocomplete string

	// If true, send autocomplete="" to Alfred. If autocomplete is not
	// specified, TAB will do nothing. If autocomplete is an empty
	// string, TAB will autocomplete to an empty string, i.e. Alfred's
	// query will be deleted.
	KeepEmptyAutocomplete bool

	// Copytext is what CMD+C should copy instead of Arg (the default).
	Copytext string

	// Largetext is what is shown in Alfred's Large Text window on CMD+L
	// instead of Arg (the default).
	Largetext string

	// Whether the result is "actionable", i.e. ENTER will pass Arg to
	// subsequent action.
	Valid bool

	// The type of the result. Currently, "file" is the only value Alfred
	// understands. If set to "file" and Arg is a valid filepath, user
	// can use Alfred's File Actions on the item.
	// Type string `xml:"type,attr,omitempty"`

	// IsFile tells Alfred that this Item is a file, i.e. Arg is a path
	// and Alfred's File Actions should be made available.
	IsFile bool

	// The icon for the result. Can point to an image file, a filepath
	// of a file whose icon should be used, or a UTI, such as
	// "com.apple.folder".
	Icon *ItemIcon `xml:"icon,omitempty"`
}

// SetAlternateSubtitle sets custom subtitles for modifier keys.
// `modifier` must be one of "cmd", "opt", "ctrl", "shift", "fn".
func (it *Item) SetAlternateSubtitle(modifier string, value string) error {
	modifier = strings.ToLower(modifier)
	if _, valid := validModifiers[modifier]; !valid {
		return fmt.Errorf("Invalid modifier: %v", modifier)
	}
	if it.AlternateSubtitles == nil {
		it.AlternateSubtitles = map[string]string{}
	}
	it.AlternateSubtitles[modifier] = value
	// sub := Subtitle{}
	// sub.Value = value
	// sub.Modifier = modifier
	// it.AlternateSubtitles = append(it.AlternateSubtitles, sub)
	return nil
}

// SetIcon sets the icon for a result item.
// Pass "" for kind if value is the path to an icon file.
func (it *Item) SetIcon(value string, kind string) error {
	kind = strings.ToLower(kind)
	if _, valid := validIconTypes[kind]; !valid {
		return fmt.Errorf("Invalid icon kind: %v", kind)
	}
	if it.Icon == nil {
		it.Icon = &ItemIcon{}
	}
	it.Icon.Value = value
	it.Icon.Type = kind
	return nil
}

// Helper function to add an XML element/tag.
func (it *Item) addElement(name string, cdata string, attrs map[string]string) []xml.Token {
	var attr []xml.Attr
	if attrs != nil {
		for n, v := range attrs {
			attr = append(attr, xml.Attr{xml.Name{"", n}, v})
		}
	}
	elem := xml.StartElement{xml.Name{"", name}, attr}
	tokens := []xml.Token{elem}
	if cdata != "" {
		tokens = append(tokens, xml.CharData(cdata))
	}
	tokens = append(tokens, xml.EndElement{elem.Name})
	return tokens
}

// MarshalXML serializes Item to Alfred's XML format. You shouldn't
// need to call this directly: use Feedback.Send() instead.
func (it *Item) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// var attrs []xml.Attr
	var attr xml.Attr

	// Validate Item (<title> is the only required value)
	if it.Title == "" {
		return fmt.Errorf("You must specify Title")
	}

	if it.IsFile {
		attr = xml.Attr{xml.Name{"", "type"}, "file"}
		start.Attr = append(start.Attr, attr)
	}
	// <item>
	start.Name.Local = "item"

	// Attributes on <item>
	if it.UID != "" {
		attr = xml.Attr{xml.Name{"", "uid"}, it.UID}
		start.Attr = append(start.Attr, attr)
	}

	if it.Autocomplete != "" || it.KeepEmptyAutocomplete == true {
		attr = xml.Attr{xml.Name{"", "autocomplete"}, it.Autocomplete}
		start.Attr = append(start.Attr, attr)
	}

	if it.Valid == true {
		attr = xml.Attr{xml.Name{"", "valid"}, "yes"}
	} else {
		attr = xml.Attr{xml.Name{"", "valid"}, "no"}
	}
	start.Attr = append(start.Attr, attr)

	tokens := []xml.Token{start}

	// Sub-elements of <item>
	tokens = append(tokens, it.addElement("title", it.Title, nil)...)
	if it.Subtitle != "" {
		tokens = append(tokens, it.addElement("subtitle", it.Subtitle, nil)...)
	}

	if it.AlternateSubtitles != nil {
		for mod, text := range it.AlternateSubtitles {
			attrs := make(map[string]string, 1)
			attrs["mod"] = mod
			tokens = append(tokens, it.addElement("subtitle", text, attrs)...)
		}
	}

	if it.Arg != "" {
		tokens = append(tokens, it.addElement("arg", it.Arg, nil)...)
	}
	if it.Copytext != "" {
		attrs := make(map[string]string, 1)
		attrs["type"] = "copy"
		tokens = append(tokens, it.addElement("text", it.Copytext, attrs)...)
	}
	if it.Largetext != "" {
		attrs := make(map[string]string, 1)
		attrs["type"] = "largetype"
		tokens = append(tokens, it.addElement("text", it.Largetext, attrs)...)
	}

	if it.Icon != nil {
		attrs := make(map[string]string, 1)
		if it.Icon.Type != "" {
			attrs["type"] = it.Icon.Type
		}
		tokens = append(tokens, it.addElement("icon", it.Icon.Value, attrs)...)
	}
	// </item>
	tokens = append(tokens, xml.EndElement{start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	err := e.Flush()
	if err != nil {
		return err
	}
	return nil
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
//    mdls -name kMDItemContentType -raw /path/to/the/file
//
// This will only work on Spotlight-indexed files.
type ItemIcon struct {
	Value   string
	Type    string
	XMLName xml.Name
}

// Feedback contains Items. This is the top-level object for generating
// Alfred XML (i.e. serialise this and send it to Alfred).
type Feedback struct {
	Items   []*Item
	XMLName xml.Name `xml:"items"`
}

// NewItem adds a new Item and returns a pointer to it.
func (fb *Feedback) NewItem() *Item {
	it := &Item{}
	fb.Items = append(fb.Items, it)
	return it
}

// NewFileItem adds and returns a pointer to a new item pre-populated from path.
// Title is the base name of the file
// Subtitle is the path to the file (using "~" for $HOME)
// Valid is "YES"
// UID, Arg and Autocomplete are set to path
// Type is "file"
// Icon is the icon of the file at path
func (fb *Feedback) NewFileItem(path string) *Item {
	it := fb.NewItem()
	it.Title = filepath.Base(path)
	it.Subtitle = util.ShortenPath(path)
	it.Arg = path
	it.Valid = true
	it.UID = path
	it.Autocomplete = path
	it.IsFile = true
	it.SetIcon(path, "fileicon")
	return it
}

// Send generates XML from this struct and sends it to Alfred.
func (fb *Feedback) Send() error {
	output, err := xml.MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("Error generating XML : %v", err)
	}
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
	return nil
}
