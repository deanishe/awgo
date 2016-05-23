package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// TODO: Add Options

// Text encapsulates the copytext and largetext values for a result Item.
type Text struct {
	// Copied to the clipboard on CMD+C
	Copy string
	// Shown in Alfred's Large Type window on CMD+L
	Large string
}

// Item is a single Alfred result. Add them to a Feedback struct to
// generate valid Alfred JSON.
type Item struct {
	// Result title (only required field)
	Title string `json:"title"`

	// Result subtitle
	Subtitle string `json:"subtitle,omitempty"`

	// Custom subtitles for when modifier keys are held
	AlternateSubtitles map[string]string `json:"mods,omitempty"`
	// AlternateSubtitles []Subtitle

	// The value that is passed as {query} to the next action in the workflow
	Arg string `json:"arg,omitempty"`

	// Used by Alfred to remember your choices. Use blank string
	// to force results to appear in the order you generate them.
	UID string `json:"uid,omitempty"`

	// What the query will expand to when the user TABs it (or hits
	// RETURN on an invalid result)
	Autocomplete string `json:"-"`

	// If true, send autocomplete="" to Alfred. If autocomplete is not
	// specified, TAB will do nothing. If autocomplete is an empty
	// string, TAB will autocomplete to an empty string, i.e. Alfred's
	// query will be deleted.
	KeepEmptyAutocomplete bool `json:"-"`

	// Copytext is what CMD+C should copy instead of Arg (the default).
	Copytext string `json:"-"`

	// Largetext is what is shown in Alfred's Large Text window on CMD+L
	// instead of Arg (the default).
	Largetext string `json:"-"`

	// Text *Text `json:"text,omitempty"`

	// Whether the result is "actionable", i.e. ENTER will pass Arg to
	// subsequent action.
	Valid bool `json:"valid,omitempty"`

	// IsFile tells Alfred that this Item is a file, i.e. Arg is a path
	// and Alfred's File Actions should be made available.
	IsFile bool `json:"-"`

	// The icon for the result. Can point to an image file, a filepath
	// of a file whose icon should be used, or a UTI, such as
	// "com.apple.folder".
	Icon *ItemIcon `json:"icon,omitempty"`
}

// SetAlternateSubtitle sets custom subtitles for modifier keys.
// modifier must be one of "cmd", "opt", "ctrl", "shift", "fn".
//
// TODO: Update alternate subtitles for Alfred 3 model
func (it *Item) SetAlternateSubtitle(modifier string, value string) error {
	modifier = strings.ToLower(modifier)
	if _, valid := validModifiers[modifier]; !valid {
		return fmt.Errorf("Invalid modifier: %v", modifier)
	}
	if it.AlternateSubtitles == nil {
		it.AlternateSubtitles = map[string]string{}
	}
	it.AlternateSubtitles[modifier] = value

	return nil
}

// SetIcon sets the icon for a result item.
// Pass "" for kind if value is the path to an icon file. Other valid
// values are "fileicon" and "filetype". See ItemIcon for more information.
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

// MarshalJSON serializes Item to Alfred 3's JSON format. You shouldn't
// need to call this directly: use Feedback.Send() instead.
//
// A custom serializer is necessary because Alfred behaves
// differently when autocomplete is missing or when present, but empty.
func (it *Item) MarshalJSON() ([]byte, error) {

	type Alias Item
	var auto *string
	var t string

	if it.Autocomplete != "" || it.KeepEmptyAutocomplete {
		auto = &it.Autocomplete
	}
	if it.IsFile {
		t = "file"
	}

	// TODO: JSON: Copy & Large Text
	// TODO: JSON: Alternate subtitles

	return json.Marshal(&struct {
		Auto *string `json:"autocomplete,omitempty"`
		Type string  `json:"type,omitempty"`
		*Alias
	}{
		Auto:  auto,
		Type:  t,
		Alias: (*Alias)(it),
	})

}

// ItemIcon represents the icon for an Item.
//
// Alfred supports PNG or ICNS files, UTIs (e.g. "public.folder") or
// can use the icon of a specific file (e.g. "/Applications/Safari.app"
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
	Value string `json:"path"`
	Type  string `json:"type,omitempty"`
}

// Feedback contains Items. This is the top-level object for generating
// Alfred XML (i.e. serialise this and send it to Alfred).
//
// Use NewFeedback() to create new (initialised) Feedback structs.
type Feedback struct {
	Items []*Item `json:"items"`
	// Set to true when feedback has been sent.
	sent bool
}

// NewFeedback creates a new, initialised Feedback struct.
func NewFeedback() *Feedback {
	fb := &Feedback{}
	fb.Items = []*Item{}
	return fb
}

// Clear removes any items.
func (fb *Feedback) Clear() {
	if len(fb.Items) > 0 {
		fb.Items = nil
	}
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
	it.Subtitle = ShortenPath(path)
	it.Arg = path
	it.Valid = true
	it.UID = path
	it.Autocomplete = path
	it.IsFile = true
	it.SetIcon(path, "fileicon")
	return it
}

// Send generates JSON from this struct and sends it to Alfred.
func (fb *Feedback) Send() error {
	if fb.sent {
		log.Printf("Feedback already sent. Ignoring.")
		return nil
	}
	output, err := json.MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("Error generating JSON : %v", err)
	}

	os.Stdout.Write(output)
	fb.sent = true
	return nil
}
