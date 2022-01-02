// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"go.deanishe.net/fuzzy"
)

// Valid modifier keys pressed by the user to run an alternate
// item action in Script Filters (in combination with ↩).
// Passed to Item.NewModifier().
//
// Alfred 3 only permits one modifier at a time, but in Alfred 4+
// you can combine them arbitrarily.
const (
	ModCmd   string = "cmd"   // Alternate action for ⌘↩
	ModAlt   string = "alt"   // Alternate action for ⌥↩
	ModOpt   string = "alt"   // Synonym for ModAlt
	ModCtrl  string = "ctrl"  // Alternate action for ^↩
	ModShift string = "shift" // Alternate action for ⇧↩
	ModFn    string = "fn"    // Alternate action for fn↩
)

// Types understood by Alfred's `action` API call and item field. Added in Alfred 4.5.
const (
	TypeFile = "file" // values are paths
	TypeURL  = "url"  // values are URLs
	TypeText = "text" // values are just text
)

// Item is a single Alfred Script Filter result.
// Together with Feedback & Modifier, Item generates Script Filter feedback
// for Alfred.
//
// Create Items via NewItem(), so they are bound to their parent Feedback.
type Item struct {
	title        string
	subtitle     *string
	match        *string
	uid          *string
	autocomplete *string
	arg          []string
	valid        bool
	file         bool
	copytext     *string
	largetype    *string
	ql           *string
	vars         map[string]string
	mods         map[string]*Modifier
	actions      map[string][]string
	icon         *Icon
	noUID        bool // Suppress UID in JSON
}

// Title sets the title of the item in Alfred's results.
func (it *Item) Title(s string) *Item {
	it.title = s
	return it
}

// Subtitle sets the subtitle of the item in Alfred's results.
func (it *Item) Subtitle(s string) *Item {
	it.subtitle = &s
	return it
}

// Match sets Item's match field for filtering.
// If present, this field is preferred over the item's title for fuzzy sorting
// via Feedback, and by Alfred's "Alfred filters results" feature.
func (it *Item) Match(s string) *Item {
	it.match = &s
	return it
}

// Arg sets Item's arg, the value(s) passed as {query} to the next workflow action.
// Multiple values are allowed in Alfred 4.1 and later.
func (it *Item) Arg(s ...string) *Item {
	it.arg = s
	return it
}

// UID sets Item's unique ID, which is used by Alfred to remember your choices.
// Use a blank string to force results to appear in the order you add them.
//
// You can also use the SuppressUIDs() Option to (temporarily) suppress output of UIDs.
func (it *Item) UID(s string) *Item {
	if it.noUID {
		return it
	}
	it.uid = &s
	return it
}

// Autocomplete sets what Alfred's query expands to when the user TABs result.
// (or hits RETURN on a result where valid is false)
func (it *Item) Autocomplete(s string) *Item {
	it.autocomplete = &s
	return it
}

// Valid tells Alfred whether the result is "actionable", i.e. ENTER will
// pass Arg to subsequent action.
func (it *Item) Valid(b bool) *Item {
	it.valid = b
	return it
}

// IsFile tells Alfred that this Item is a file, i.e. Arg is a path
// and Alfred's File Actions should be made available.
func (it *Item) IsFile(b bool) *Item {
	it.file = b
	return it
}

// Copytext is what CMD+C should copy instead of Arg (the default).
func (it *Item) Copytext(s string) *Item {
	it.copytext = &s
	return it
}

// Largetype is what is shown in Alfred's Large Text window on CMD+L
// instead of Arg (the default).
func (it *Item) Largetype(s string) *Item {
	it.largetype = &s
	return it
}

// Quicklook is a path or URL shown in a macOS Quicklook window on SHIFT
// or CMD+Y.
func (it *Item) Quicklook(s string) *Item {
	it.ql = &s
	return it
}

// Icon sets the icon for the Item.
// Can point to an image file, a filepath of a file whose icon should be used,
// or a UTI.
//
// See the documentation for Icon for more details.
func (it *Item) Icon(icon *Icon) *Item {
	it.icon = icon
	return it
}

// Action sets the value(s) to be passed to Alfred's Universal Actions if
// the user actions this item. Alfred will auto-detect the type of the value(s).
//
// Added in Alfred 4.5.
func (it *Item) Action(value ...string) *Item { return it.ActionForType("", value...) }

// ActionForType sets the value(s) to be passed to Alfred's Universal Actions if
// the user actions this item. Type may be one of "file", "url" or "text".
//
// Added in Alfred 4.5.
func (it *Item) ActionForType(typ string, value ...string) *Item {
	if typ == "" {
		typ = "auto"
	}
	if it.actions == nil {
		it.actions = map[string][]string{}
	}
	it.actions[typ] = value
	return it
}

// Var sets an Alfred variable for subsequent workflow elements.
func (it *Item) Var(k, v string) *Item {
	if it.vars == nil {
		it.vars = make(map[string]string, 1)
	}
	it.vars[k] = v
	return it
}

// NewModifier returns an initialised Modifier bound to this Item.
// It also populates the Modifier with any workflow variables set in the Item.
//
// You must specify at least one modifier key. Alfred 3 only supports
// a single modifier, but Alfred 4+ allow them to be arbitrarily combined.
// Any invalid modifier keys are ignored. If you specify an unusable set of
// modifiers (i.e. they evaluate to ""), although a Modifier is returned,
// it is not retained by Item and will not be sent to Alfred. An error message
// is also logged.
func (it *Item) NewModifier(key ...string) *Modifier {
	m := newModifier(key...)
	// Add Item variables to Modifier
	if it.vars != nil {
		for k, v := range it.vars {
			m.Var(k, v)
		}
	}

	it.SetModifier(m)
	return m
}

// SetModifier sets a Modifier for a modifier key.
func (it *Item) SetModifier(m *Modifier) {
	if m.Key == "" {
		log.Printf("[ERROR] modifier has no key: %#v", m)
		return
	}
	if it.mods == nil {
		it.mods = map[string]*Modifier{}
	}
	it.mods[m.Key] = m
}

// Cmd returns an initialised Modifier bound to this Item and the CMD (⌘) key.
func (it *Item) Cmd() *Modifier { return it.NewModifier(ModCmd) }

// Alt returns an initialised Modifier bound to this Item and the ALT/OPT (⌥) key.
func (it *Item) Alt() *Modifier { return it.NewModifier(ModAlt) }

// Opt is a synonym for Alt().
func (it *Item) Opt() *Modifier { return it.Alt() }

// Ctrl returns an initialised Modifier bound to this Item and the CTRL (^) key.
func (it *Item) Ctrl() *Modifier { return it.NewModifier(ModCtrl) }

// Shift returns an initialised Modifier bound to this Item and the SHIFT (⇧) key.
func (it *Item) Shift() *Modifier { return it.NewModifier(ModShift) }

// Fn returns an initialised Modifier bound to this Item and the fn key.
func (it *Item) Fn() *Modifier { return it.NewModifier(ModFn) }

// Vars returns the Item's workflow variables.
func (it *Item) Vars() map[string]string {
	return it.vars
}

// MarshalJSON serializes Item to Alfred's JSON format.
// You shouldn't need to call this directly: use SendFeedback() instead.
func (it *Item) MarshalJSON() ([]byte, error) {
	var (
		typ  string
		ql   string
		text *itemText
	)

	if it.file {
		typ = "file"
	}

	if it.ql != nil {
		ql = *it.ql
	}

	if it.copytext != nil || it.largetype != nil {
		text = &itemText{Copy: it.copytext, Large: it.largetype}
	}

	// Serialise Item
	v := struct {
		Title     string               `json:"title"`
		Subtitle  *string              `json:"subtitle,omitempty"`
		Match     *string              `json:"match,omitempty"`
		Auto      *string              `json:"autocomplete,omitempty"`
		Arg       interface{}          `json:"arg,omitempty"`
		UID       *string              `json:"uid,omitempty"`
		Valid     bool                 `json:"valid"`
		Type      string               `json:"type,omitempty"`
		Text      *itemText            `json:"text,omitempty"`
		Icon      *Icon                `json:"icon,omitempty"`
		Quicklook string               `json:"quicklookurl,omitempty"`
		Variables map[string]string    `json:"variables,omitempty"`
		Mods      map[string]*Modifier `json:"mods,omitempty"`
		Actions   map[string][]string  `json:"action,omitempty"`
	}{
		Title:     it.title,
		Subtitle:  it.subtitle,
		Match:     it.match,
		Auto:      it.autocomplete,
		UID:       it.uid,
		Valid:     it.valid,
		Type:      typ,
		Text:      text,
		Icon:      it.icon,
		Quicklook: ql,
		Variables: it.vars,
		Mods:      it.mods,
		Actions:   it.actions,
	}
	// serialise single arg as string
	if len(it.arg) == 1 {
		v.Arg = it.arg[0]
	} else if len(it.arg) > 1 {
		v.Arg = it.arg
	}
	return Marshal(v)
}

// itemText encapsulates the copytext and largetext values for a result Item.
type itemText struct {
	// Copied to the clipboard on CMD+C
	Copy *string `json:"copy,omitempty"`
	// Shown in Alfred's Large Type window on CMD+L
	Large *string `json:"largetype,omitempty"`
}

// Modifier encapsulates alterations to Item when a modifier key is held when
// the user actions the item.
//
// Create new Modifiers via Item.NewModifier(). This binds the Modifier to the
// Item, initializes Modifier's map and inherits Item's workflow variables.
// Variables are inherited at creation time, so any Item variables you set
// after creating the Modifier are not inherited.
type Modifier struct {
	// The modifier key, e.g. "cmd", "alt".
	// With Alfred 4+, modifiers can be combined, e.g. "cmd+alt", "ctrl+shift+cmd"
	Key      string
	arg      []string
	subtitle *string
	valid    bool
	icon     *Icon
	vars     map[string]string
}

// newModifier creates a Modifier, validating key.
func newModifier(key ...string) *Modifier {
	l := []string{}
	for _, k := range key {
		s := strings.TrimSpace(strings.ToLower(k))
		if s == "opt" {
			s = "alt"
		}
		if s == "" {
			continue
		}
		if s != "alt" && s != "cmd" && s != "ctrl" && s != "fn" && s != "shift" {
			log.Printf("[warning] ignored invalid modifier %q", k)
			continue
		}
		l = append(l, s)
	}
	sort.Strings(l)
	s := strings.Join(l, "+")
	return &Modifier{Key: s, vars: map[string]string{}}
}

// Arg sets the arg for the Modifier. Multiple values are allowed in Alfred 4.1 and later.
func (m *Modifier) Arg(s ...string) *Modifier {
	m.arg = s
	return m
}

// Subtitle sets the subtitle for the Modifier.
func (m *Modifier) Subtitle(s string) *Modifier {
	m.subtitle = &s
	return m
}

// Valid sets the valid status for the Modifier.
func (m *Modifier) Valid(v bool) *Modifier {
	m.valid = v
	return m
}

// Icon sets an icon for the Modifier.
func (m *Modifier) Icon(i *Icon) *Modifier {
	m.icon = i
	return m
}

// Var sets a variable for the Modifier.
func (m *Modifier) Var(k, v string) *Modifier {
	m.vars[k] = v
	return m
}

// Vars returns all Modifier variables.
func (m *Modifier) Vars() map[string]string {
	return m.vars
}

// MarshalJSON serializes Item to Alfred 3's JSON format.
// You shouldn't need to call this directly: use SendFeedback() instead.
func (m *Modifier) MarshalJSON() ([]byte, error) {
	v := struct {
		Arg       interface{}       `json:"arg,omitempty"`
		Subtitle  *string           `json:"subtitle,omitempty"`
		Valid     bool              `json:"valid,omitempty"`
		Icon      *Icon             `json:"icon,omitempty"`
		Variables map[string]string `json:"variables,omitempty"`
	}{
		Subtitle:  m.subtitle,
		Valid:     m.valid,
		Icon:      m.icon,
		Variables: m.vars,
	}

	// serialise single arg as string
	if len(m.arg) == 1 {
		v.Arg = m.arg[0]
	} else if len(m.arg) > 1 {
		v.Arg = m.arg
	}

	return Marshal(v)
}

// Feedback represents the results for an Alfred Script Filter.
//
// Normally, you won't use this struct directly, but via the Workflow methods
// NewItem(), SendFeedback(), etc. It is important to use the constructor
// functions for Feedback, Item and Modifier structs so they are properly
// initialised and bound to their parent.
type Feedback struct {
	Items  []*Item           // The results to be sent to Alfred.
	NoUIDs bool              // If true, suppress Item UIDs.
	rerun  float64           // Tell Alfred to re-run Script Filter.
	sent   bool              // Set to true when feedback has been sent.
	vars   map[string]string // Top-level feedback variables.
}

// NewFeedback creates a new, initialised Feedback struct.
func NewFeedback() *Feedback {
	return &Feedback{Items: []*Item{}, vars: map[string]string{}}
}

// Var sets an Alfred variable for subsequent workflow elements.
func (fb *Feedback) Var(k, v string) *Feedback {
	if fb.vars == nil {
		fb.vars = make(map[string]string, 1)
	}
	fb.vars[k] = v
	return fb
}

// Rerun tells Alfred to re-run the Script Filter after `secs` seconds.
func (fb *Feedback) Rerun(secs float64) *Feedback {
	fb.rerun = secs
	return fb
}

// Vars returns the Feedback's workflow variables.
func (fb *Feedback) Vars() map[string]string {
	return fb.vars
}

// Clear removes any items.
func (fb *Feedback) Clear() {
	if len(fb.Items) > 0 {
		fb.Items = []*Item{}
	}
}

// IsEmpty returns true if Feedback contains no items.
func (fb *Feedback) IsEmpty() bool { return len(fb.Items) == 0 }

// NewItem adds a new Item and returns a pointer to it.
//
// The Item inherits any workflow variables set on the Feedback parent at
// time of creation.
func (fb *Feedback) NewItem(title string) *Item {
	it := &Item{title: title, vars: map[string]string{}, noUID: fb.NoUIDs}

	// Add top-level variables to Item. The reason for this is that
	// (older versions of) Alfred drops all item- and top-level variables
	// on the floor if a modifier has any variables set (i.e. only the
	// modifier's variables are retained). So, add top-level variables to Item
	// (and in turn to any Modifiers) to enforce more sensible behaviour.
	for k, v := range fb.vars {
		it.vars[k] = v
	}

	fb.Items = append(fb.Items, it)
	return it
}

// MarshalJSON serializes Feedback to Alfred's JSON format.
// You shouldn't need to call this: use Send() instead.
func (fb *Feedback) MarshalJSON() ([]byte, error) {
	return Marshal(&struct {
		Variables map[string]string `json:"variables,omitempty"`
		Rerun     float64           `json:"rerun,omitempty"`
		Items     []*Item           `json:"items"`
	}{
		Items:     fb.Items,
		Rerun:     fb.rerun,
		Variables: fb.vars,
	})
}

// Send generates JSON from this struct and sends it to Alfred
// (by writing the JSON to STDOUT).
//
// You shouldn't need to call this directly: use SendFeedback() instead.
func (fb *Feedback) Send() error {
	if fb.sent {
		log.Printf("Feedback already sent. Ignoring.")
		return nil
	}
	output, err := MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("Error generating JSON : %w", err)
	}

	os.Stdout.Write(output)
	fb.sent = true
	log.Printf("Sent %d result(s) to Alfred", len(fb.Items))
	return nil
}

// Sort sorts Items against query. Uses a fuzzy.Sorter with the specified
// options.
func (fb *Feedback) Sort(query string, opts ...fuzzy.Option) []*fuzzy.Result {
	s := fuzzy.New(fb, opts...)
	return s.Sort(query)
}

// Filter fuzzy-sorts Items against query and deletes Items that don't match.
// It returns a slice of Result structs, which contain the results of the
// fuzzy sorting.
func (fb *Feedback) Filter(query string, opts ...fuzzy.Option) []*fuzzy.Result {
	var (
		items []*Item
		res   []*fuzzy.Result
	)

	r := fb.Sort(query, opts...)
	for i, it := range fb.Items {
		if r[i].Match {
			items = append(items, it)
			res = append(res, r[i])
		}
	}
	fb.Items = items
	return res
}

// Keywords implements fuzzy.Sortable.
//
// Returns the match or title field for Item i.
func (fb *Feedback) Keywords(i int) string {
	it := fb.Items[i]
	// Sort on title if match isn't set
	if it.match != nil {
		return *it.match
	}
	return it.title
}

// Len implements sort.Interface.
func (fb *Feedback) Len() int { return len(fb.Items) }

// Less implements sort.Interface.
func (fb *Feedback) Less(i, j int) bool {
	// we only want to sort based on fuzzy match score
	return false
}

// Swap implements sort.Interface.
func (fb *Feedback) Swap(i, j int) { fb.Items[i], fb.Items[j] = fb.Items[j], fb.Items[i] }

// ArgVars lets you set workflow variables from Run Script actions.
// It emits the arg and variables you set in the format required by Alfred.
//
// Use ArgVars.Send() to pass variables to downstream workflow elements.
type ArgVars struct {
	arg  []string
	vars map[string]string
}

// NewArgVars returns an initialised ArgVars object.
func NewArgVars() *ArgVars {
	return &ArgVars{vars: map[string]string{}}
}

// Arg sets the arg(s)/query to be passed to the next workflow action.
// Multiple values are allowed in Alfred 4.1 and later.
func (a *ArgVars) Arg(s ...string) *ArgVars {
	a.arg = s
	return a
}

// Vars returns ArgVars' variables.
// NOTE: This function only returns variables you have set with ArgVars.Var()
// for export to Alfred during this run. To read variables from the environment,
// use Workflow.Config.
func (a *ArgVars) Vars() map[string]string {
	return a.vars
}

// Var sets the value of a workflow variable.
func (a *ArgVars) Var(k, v string) *ArgVars {
	a.vars[k] = v
	return a
}

// String returns a string representation.
//
// If any variables are set, JSON is returned. Otherwise, a plain string
// is returned.
func (a *ArgVars) String() (string, error) {
	if len(a.vars) == 0 && len(a.arg) < 2 {
		if len(a.arg) == 0 {
			return "", nil
		}
		return a.arg[0], nil
	}
	// Vars set, so return as JSON
	data, err := a.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Send outputs arg and variables to Alfred by printing a response to STDOUT.
func (a *ArgVars) Send() error {
	data, err := a.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = fmt.Print(string(data))
	return err
}

// MarshalJSON serialises ArgVars to JSON.
// You probably don't need to call this: use ArgVars.Send() instead.
func (a *ArgVars) MarshalJSON() ([]byte, error) {
	// Return arg regardless of whether it's empty or not:
	// we have to return *something*
	if len(a.vars) == 0 && len(a.arg) < 2 {
		// Want empty string, i.e. "", not null
		if len(a.arg) == 0 {
			return []byte(`""`), nil
		}
		return Marshal(a.arg[0])
	}

	v := struct {
		Arg  interface{}       `json:"arg,omitempty"`
		Vars map[string]string `json:"variables,omitempty"`
	}{
		Vars: a.vars,
	}

	if len(a.arg) == 1 {
		v.Arg = a.arg[0]
	} else if len(a.arg) > 1 {
		v.Arg = a.arg
	}

	return Marshal(&struct {
		Root interface{} `json:"alfredworkflow"`
	}{
		Root: v,
	})
}

func Marshal(v interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	return buffer.Bytes(), err
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
