//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Valid modifier keys for Item.NewModifier(). You can't combine these
// in any way: Alfred only permits one modifier at a time.
const (
	ModCmd   = "cmd"
	ModAlt   = "alt"
	ModCtrl  = "ctrl"
	ModShift = "shift"
	ModFn    = "fn"
)

// Valid icon types for ItemIcon. You can use an image file, the icon of a file,
// e.g. an application's icon, or the icon for a filetype (specified by a UTI).
const (
	// Use with image files you wish to show in Alfred.
	IconTypeImageFile = ""
	// Use to show the icon of a file, e.g. combine with "/Applications/Safari.app"
	// to show Safari's icon in Alfred.
	IconTypeFileIcon = "fileicon"
	// Use together with a UTI to show the icon for a filetype, e.g. "public.folder",
	// which will give you the icon for a folder.
	IconTypeFileType = "filetype"
)

var (
	// ValidModifiers are permitted modifier keys for Modifiers.
	// See Item.NewModifier() for application.
	ValidModifiers = []string{ModCmd, ModAlt, ModCtrl, ModShift, ModFn}

	// ValidIconTypes are the values you may specify for an icon type.
	ValidIconTypes = []string{IconTypeImageFile, IconTypeFileIcon, IconTypeFileType}

	// Maps to shadow the above to make lookup easier.
	validModifiers = make(map[string]bool, len(ValidModifiers))
	validIconTypes = make(map[string]bool, len(ValidIconTypes))
)

func init() {
	// Build lookup maps (why doesn't Go have sets?)
	for _, s := range ValidModifiers {
		validModifiers[s] = true
	}
	for _, s := range ValidIconTypes {
		validIconTypes[s] = true
	}
}

// itemArg is a result (Item) argument. It may contain a single string, or it
// may also contain workflow variables.
//
// This is a helper struct to simplify encoding Items and Modifiers to JSON.
type itemArg struct {
	arg    string
	argSet bool
	vars   map[string]string
}

// newArg returns an initialised Arg.
func newArg() *itemArg {
	return &itemArg{vars: map[string]string{}}
}

// Arg returns Arg's arg.
func (a *itemArg) Arg() string {
	return a.arg
}

// SetArg sets Arg's arg.
func (a *itemArg) SetArg(s string) {
	a.arg = s
	a.argSet = true
}

// Vars returns Arg's variables.
func (a *itemArg) Vars() map[string]string {
	return a.vars
}

// Var returns value set for key k.
func (a *itemArg) Var(k string) string {
	return a.vars[k]
}

// SetVar sets the value of a variable.
func (a *itemArg) SetVar(k, v string) {
	a.vars[k] = v
}

// String returns a JSON string representation of Arg.
func (a *itemArg) String() (string, error) {
	if len(a.vars) == 0 {
		return a.arg, nil
	}
	data, err := a.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalJSON serialises Arg to JSON.
func (a *itemArg) MarshalJSON() ([]byte, error) {

	var arg *string

	// Return arg regardless of whether it's empty or not:
	// we have return *something*
	if len(a.vars) == 0 {
		return json.Marshal(a.Arg())
	}

	if a.argSet {
		arg = &a.arg
	}

	return json.Marshal(&struct {
		Root interface{} `json:"alfredworkflow"`
	}{
		Root: &struct {
			Arg  *string           `json:"arg,omitempty"`
			Vars map[string]string `json:"variables"`
		}{
			Arg:  arg,
			Vars: a.vars,
		},
	})
}

// Modifier encapsulates alterations to Item when a modifier key is held when
// the user actions the item.
//
// Create new Modifiers via Item.NewModifier(). This binds the Modifier to the
// Item, initializes Modifier's map and inherits Item's workflow variables.
//
// A Modifier created via Item.NewModifier() also inherits its parent Item's
// workflow variables.
type Modifier struct {
	// The modifier key. May be any of ValidModifiers.
	Key         string
	arg         string
	argSet      bool
	subtitle    string
	subtitleSet bool
	valid       bool
	validSet    bool
	vars        map[string]string
}

// newModifier creates a Modifier, validating key.
func newModifier(key string) (*Modifier, error) {
	if ok := validModifiers[key]; !ok {
		return nil, fmt.Errorf("Invalid modifier key: %s", key)
	}
	return &Modifier{Key: key, vars: map[string]string{}}, nil
}

// Arg returns the arg set for the Modifier.
func (m *Modifier) Arg() string {
	return m.arg
}

// SetArg sets the arg for the Modifier.
func (m *Modifier) SetArg(s string) {
	m.arg = s
	m.argSet = true
}

// Subtitle returns the subtitle set for the Modifier.
func (m *Modifier) Subtitle() string {
	return m.subtitle
}

// SetSubtitle sets the subtitle for the Modifier.
func (m *Modifier) SetSubtitle(s string) {
	m.subtitle = s
	m.subtitleSet = true
}

// Valid returns the valid status for the Modifier.
func (m *Modifier) Valid() bool {
	return m.valid
}

// SetValid sets the valid for the Modifier.
func (m *Modifier) SetValid(v bool) {
	m.valid = v
	m.validSet = true
}

// SetVar sets a variable for the Modifier.
func (m *Modifier) SetVar(k, v string) {
	m.vars[k] = v
}

// Var returns Modifier variable for key k.
func (m *Modifier) Var(k string) string {
	return m.vars[k]
}

// Vars returns all Modifier variables.
func (m *Modifier) Vars() map[string]string {
	return m.vars
}

// MarshalJSON implements the JSON serialization interface.
func (m *Modifier) MarshalJSON() ([]byte, error) {

	var a, s *string
	var v *bool

	if m.argSet {
		a = &m.arg
	}
	if m.subtitleSet {
		s = &m.subtitle
	}
	if m.validSet {
		v = &m.valid
	}

	// Variables
	if len(m.vars) > 0 {
		arg := newArg()
		if m.argSet {
			arg.SetArg(m.arg)
		}
		for k, v := range m.vars {
			arg.SetVar(k, v)
		}
		if s, err := arg.String(); err == nil {
			a = &s
		} else {
			log.Printf("Error encoding variables: %v", err)
		}
	}

	return json.Marshal(&struct {
		Arg      *string `json:"arg,omitempty"`
		Subtitle *string `json:"subtitle,omitempty"`
		Valid    *bool   `json:"valid,omitempty"`
	}{
		Arg:      a,
		Subtitle: s,
		Valid:    v,
	})
}

// itemText encapsulates the copytext and largetext values for a result Item.
type itemText struct {
	// Copied to the clipboard on CMD+C
	Copy string `json:"copy,omitempty"`
	// Shown in Alfred's Large Type window on CMD+L
	Large string `json:"largetype,omitempty"`
}

// Item is a single Alfred result. Add them to a Feedback struct to
// generate valid Alfred JSON.
type Item struct {
	// Result title (only required field)
	Title string `json:"title"`

	// Result subtitle
	Subtitle string `json:"subtitle,omitempty"`

	// The value that is passed as {query} to the next action in the workflow
	Arg string `json:"-"`

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

	// Modifiers are deviating values set for when the user holds down a
	// modifier key like CMD or SHIFT.
	Modifiers map[string]*Modifier `json:"mods,omitempty"`

	// Whether the result is "actionable", i.e. ENTER will pass Arg to
	// subsequent action.
	Valid bool `json:"valid,omitempty"`

	// vars are variables to pass to subsequent workflow elements.
	vars map[string]string

	// IsFile tells Alfred that this Item is a file, i.e. Arg is a path
	// and Alfred's File Actions should be made available.
	IsFile bool `json:"-"`

	// The icon for the result. Can point to an image file, a filepath
	// of a file whose icon should be used, or a UTI, such as
	// "com.apple.folder".
	Icon *ItemIcon `json:"icon,omitempty"`
}

// NewModifier returns an initialised Modifier bound to this Item.
// It also populates the Modifier with any workflow variables set in the Item.
//
// The workflow will terminate (call FatalError) if key is not a valid
// modifier.
func (it *Item) NewModifier(key string) *Modifier {
	m, err := newModifier(key)
	if err != nil {
		wf.FatalError(err)
	}

	// Add Item variables to Modifier
	if it.vars != nil {
		for k, v := range it.vars {
			m.SetVar(k, v)
		}
	}

	it.SetModifier(m)
	return m
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

// SetModifier sets a Modifier for a modifier key.
func (it *Item) SetModifier(m *Modifier) error {
	if ok := validModifiers[m.Key]; !ok {
		return fmt.Errorf("Invalid modifier: %s", m.Key)
	}
	if it.Modifiers == nil {
		it.Modifiers = map[string]*Modifier{}
	}
	it.Modifiers[m.Key] = m
	return nil
}

// SetVar sets an Alfred variable for subsequent workflow elements.
func (it *Item) SetVar(k, v string) {
	if it.vars == nil {
		it.vars = make(map[string]string, 1)
	}
	it.vars[k] = v
}

// Var returns the value of Item's workflow variable for key k.
func (it *Item) Var(k string) string {
	return it.vars[k]
}

// Vars returns the Item's workflow variables.
func (it *Item) Vars() map[string]string {
	return it.vars
}

// MarshalJSON serializes Item to Alfred 3's JSON format. You shouldn't
// need to call this directly: use Feedback.Send() instead.
//
// A custom serializer is necessary because Alfred behaves
// differently when autocomplete is missing or when present, but empty.
func (it *Item) MarshalJSON() ([]byte, error) {

	type Alias Item
	var auto *string
	var arg *string
	var t string
	var text *itemText

	if it.Autocomplete != "" || it.KeepEmptyAutocomplete {
		auto = &it.Autocomplete
	}
	if it.IsFile {
		t = "file"
	}

	if it.Copytext != "" || it.Largetext != "" {
		text = &itemText{Copy: it.Copytext, Large: it.Largetext}
	}
	// TODO: Alfred workflow config in Feedback/Item/Modifier

	if it.Arg != "" {
		arg = &it.Arg
	}

	if len(it.vars) > 0 {
		a := newArg()
		if it.Arg != "" {
			a.SetArg(it.Arg)
		}
		for k, v := range it.vars {
			a.SetVar(k, v)
		}
		if s, err := a.String(); err == nil {
			arg = &s
		} else {
			log.Printf("Error encoding variables: %v", err)
		}
	}

	// if it.vars != nil {
	// 	data, err := json.Marshal(&struct {
	// 		Root interface{} `json:"alfredworkflow"`
	// 	}{
	// 		Root: &struct {
	// 			Arg  string            `json:"arg,omitempty"`
	// 			Vars map[string]string `json:"variables"`
	// 		}{
	// 			Arg:  it.Arg,
	// 			Vars: it.vars,
	// 		},
	// 	})
	// 	// data, err := json.Marshal(&struct {
	// 	// 	Arg  string            `json:"arg,omitempty"`
	// 	// 	Vars map[string]string `json:"variables"`
	// 	// }{
	// 	// 	Arg:  it.Arg,
	// 	// 	Vars: it.Vars,
	// 	// })
	// 	if err != nil {
	// 		return []byte{}, err
	// 	}
	// 	s := string(data)
	// 	arg = &s
	// }

	return json.Marshal(&struct {
		Auto     *string   `json:"autocomplete,omitempty"`
		Argument *string   `json:"arg,omitempty"`
		Type     string    `json:"type,omitempty"`
		Text     *itemText `json:"text,omitempty"`
		*Alias
	}{
		Auto:     auto,
		Argument: arg,
		Type:     t,
		Text:     text,
		Alias:    (*Alias)(it),
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
// Alfred JSON (i.e. serialise this and send it to Alfred).
//
// Use NewFeedback() to create new (initialised) Feedback structs.
//
// It is important to use the constructor functions for Feedback, Item
// and Modifier structs.
//
// TODO: Implement Vars on Feedback.
type Feedback struct {
	Items []*Item `json:"items"`
	// Set to true when feedback has been sent.
	sent bool
	vars map[string]string
}

// NewFeedback creates a new, initialised Feedback struct.
func NewFeedback() *Feedback {
	fb := &Feedback{}
	fb.Items = []*Item{}
	fb.vars = map[string]string{}
	return fb
}

// SetVar sets an Alfred variable for subsequent workflow elements.
func (fb *Feedback) SetVar(k, v string) {
	if fb.vars == nil {
		fb.vars = make(map[string]string, 1)
	}
	fb.vars[k] = v
}

// Var returns the value of Feedback's workflow variable for key k.
func (fb *Feedback) Var(k string) string {
	return fb.vars[k]
}

// Vars returns the Feedback's workflow variables.
func (fb *Feedback) Vars() map[string]string {
	return fb.vars
}

// Clear removes any items.
func (fb *Feedback) Clear() {
	if len(fb.Items) > 0 {
		fb.Items = nil
	}
}

// NewItem adds a new Item and returns a pointer to it.
//
// The Item inherits and workflow variables set on the Feedback parent at
// time of creation.
func (fb *Feedback) NewItem(title string) *Item {
	it := &Item{Title: title, vars: map[string]string{}}

	// Variables
	if len(fb.vars) > 0 {
		for k, v := range fb.vars {
			it.SetVar(k, v)
		}
	}

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
	it := fb.NewItem(filepath.Base(path))
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
