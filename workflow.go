//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"os/exec"

	"github.com/mkrautz/plist"
)

// AwgoVersion is the semantic version number of this library.
const AwgoVersion = "0.5.0"

var (
	startTime time.Time // Time execution started

	// The workflow object operated on by top-level functions.
	// It can be retrieved/replaced with DefaultWorkflow() and
	// SetDefaultWorkflow() respectively.
	wf *Workflow

	// Flag, as we only want to set up logging once
	// TODO: Better, more pluggable logging
	logInitialized bool
)

// init creates the default Workflow.
func init() {
	startTime = time.Now()
	wf = NewWorkflow(nil)
}

// InfoPlist contains meta information extracted from info.plist.
// Use Workflow.Info() to retrieve the Info for the running
// workflow (it is lazily loaded).
//
// TODO: Do something meaningful with info.plist Variables.
type InfoPlist struct {
	BundleID    string                 `plist:"bundleid"`
	Author      string                 `plist:"createdby"`
	Description string                 `plist:"description"`
	Name        string                 `plist:"name"`
	Readme      string                 `plist:"readme"`
	Variables   map[string]interface{} `plist:"variables"`
	Version     string                 `plist:"version"`
	Website     string                 `plist:"webaddress"`
}

// Var returns the value for a variable specified in info.plist. If the
// variable is empty or unset, an empty string is returned.
//
// NOTE: This is the *default* value set in the workflow's configuration
// sheet (Workflow Environment Variables). Use os.Getenv() to get the
// current value of a variable.
func (i *InfoPlist) Var(name string) string {

	obj := i.Variables[name]
	if obj == nil {
		return ""
	}

	if s, ok := obj.(string); ok {
		return s
	}
	panic(fmt.Sprintf("Can't convert variable to string: %v", obj))
}

// Options contains the configuration options for a Workflow struct.
// Currently not a whole lot of options supported...
type Options struct {
	// GitHub is the GitHub repo the workflow should pull updates from.
	// It should have the form "username/reponame", e.g. "deanishe/alfred-ssh".
	// If set, a GitHub updater will be created and passed to SetUpdater().
	GitHub string

	// HelpURL is a link to your issues page/forum thread where users can
	// report bugs. It is shown in the debugger if the workflow crashes.
	// If no HelpURL is specified, the Website specified in the main
	// workflow setup dialog will be shown (if one is set)
	HelpURL string

	// LogPrefix is the character printed to the log at the start of each run.
	// Its purpose is to ensure the first real log message starts on its own line,
	// instead of sharing a line with Alfred's blurb in the debugger. This is only
	// printed to STDERR (i.e. Alfred's debugger), not written to the log file.
	// Default: Purple Heart (\U0001F49C)
	LogPrefix string

	// MagicPrefix overrides the default prefix for magic actions.
	MagicPrefix string

	// MaxLogSize is the size (in bytes) at which the workflow log is rotated.
	// Default: 1 MiB
	MaxLogSize int

	// MaxResults is the maximum number of results to send to Alfred.
	// 0 means send all results.
	// Default: 0
	MaxResults int

	// Fuzzy sort bonuses and penalties
	// See constants for defaults
	SortOptions *SortOptions

	// TextErrors tells Workflow to print errors as text, not JSON
	// Set to true if output goes to a Notification
	TextErrors bool
}

// Workflow provides a simple, consolidated API for building Script
// Filters and talking to Alfred.
//
// As a rule, you should create a Workflow in main() and call your main
// entry-point via Workflow.Run(). Use Workflow.NewItem() to create new
// feedback Items and Workflow.SendFeedback() to send the results to
// Alfred.
//
// If you don't need to customise Workflow's behaviour in any way, you
// can use the package-level functions, which call the corresponding
// methods on the default Workflow object.
//
// See the examples/ subdirectory for some full examples of workflows.
type Workflow struct {
	// The response that will be sent to Alfred. Workflow provides
	// convenience wrapper methods, so you don't normally have to
	// interact with this directly.
	Feedback *Feedback

	// Alfred-specific environmental variables, without the 'alfred_'
	// prefix. The following variables are present:
	//
	//     debug                        Set to "1" if Alfred's debugger is open.
	//                                  Generally, you should call Debug()/Workflow.Debug() instead.
	//     version                      Alfred version number, e.g. "2.7"
	//     version_build                Alfred build, e.g. "277"
	//     theme                        ID of current theme, e.g. "alfred.theme.custom.UUID-UUID-UUID"
	//     theme_background             Theme background colour in rgba format, e.g. "rgba(255,255,255,1.00)"
	//     theme_selection_background   Theme selection background colour in rgba format, e.g. "rgba(255,255,255,1.00)"
	//     theme_subtext                User's subtext setting.
	//                                      "0" = Always show
	//                                      "1" = Show only for alternate actions
	//                                      "2" = Never show
	//     preferences                  Path to "Alfred.alfredpreferences" file
	//     preferences_localhash        Machine-specific hash. Machine preferences are stored in
	//                                  Alfred.alfredpreferences/preferences/local/<hash>
	//     workflow_cache               Path to workflow's cache directory. Use Workflow.CacheDir()
	//                                  instead to ensure directory exists.
	//     workflow_data                Path to workflow's data directory. Use Workflow.DataDir()
	//                                  instead to ensure directory exists.
	//     workflow_name                Name of workflow, e.g. "Fast Translator"
	//     workflow_uid                 Random UID assigned to workflow by Alfred
	//     workflow_bundleid            Workflow's bundle ID from info.plist
	//     workflow_version             Workflow's version number from info.plist
	//
	// TODO: Replace Env with something better (Context object?)
	Env map[string]string

	// HelpURL is a link to your issues page/forum thread where users can
	// report bugs. It is shown in the debugger if the workflow crashes.
	// If no HelpURL is specified, the Website specified in the main
	// workflow setup dialog will be shown (if one is set)
	HelpURL string

	// LogPrefix is the character printed to the log at the start of each run.
	// Its purpose is to ensure the first real log message starts on its own line,
	// instead of sharing a line with Alfred's blurb in the debugger. This is only
	// printed to STDERR (i.e. Alfred's debugger), not written to the log file.
	// Default: Purple Heart (\U0001F49C)
	LogPrefix string

	// MaxLogSize is the size (in bytes) at which the workflow log is rotated.
	// Default: 1 MiB
	MaxLogSize int

	// MaxResults is the maximum number of results to send to Alfred.
	// 0 means send all results.
	// Default: 0
	MaxResults int

	// Fuzzy search bonuses and penalties
	// See constants for defaults
	SortOptions *SortOptions

	// TextErrors tells Workflow to print errors as text, not JSON
	// Set to true if output goes to a Notification
	TextErrors bool

	// debug is set from Alfred's `alfred_debug` environment variable.
	debug bool

	magicPrefix string // Overrides DefaultMagicPrefix for magic actions.

	// version holds value set by user or read from environment variable or info.plist
	version string

	// updater is configured if GitHub is specified in Options.
	updater *Updater

	// Populated by readInfoPlist()
	info       *InfoPlist
	infoLoaded bool

	// Set from environment or info.plist
	bundleID    string
	name        string
	cacheDir    string
	dataDir     string
	workflowDir string
}

// NewWorkflow creates and initialises a new Workflow. Use NewWorkflow to avoid
// uninitialised maps.
func NewWorkflow(o *Options) *Workflow {
	w := &Workflow{
		Env:         map[string]string{},
		Feedback:    &Feedback{},
		info:        &InfoPlist{},
		LogPrefix:   "\U0001F49C", // Purple heart
		MaxLogSize:  1048576,      // 1 MiB
		MaxResults:  0,            // Send all results to Alfred
		SortOptions: NewSortOptions(),
	}

	// Configure workflow from options
	if o != nil {
		if o.HelpURL != "" {
			w.HelpURL = o.HelpURL
		}
		// if o.Version != "" {
		// 	w.version = o.Version
		// }
		if o.LogPrefix != "" {
			w.LogPrefix = o.LogPrefix
		}
		if o.MagicPrefix != "" {
			w.magicPrefix = o.MagicPrefix
		}
		if o.MaxLogSize > 0 {
			w.MaxLogSize = o.MaxLogSize
		}
		if o.MaxResults > 0 {
			w.MaxResults = o.MaxResults
		}
		if o.SortOptions != nil {
			w.SortOptions = o.SortOptions
		}
		if o.GitHub != "" {
			var err error
			u, err := NewUpdater(&GitHub{Repo: o.GitHub})
			if err != nil {
				log.Printf("Error configuring updater: %s", err)
			} else {
				w.SetUpdater(u)
			}
		}
	}

	w.loadEnv()
	w.initializeLogging()
	EnsureExists(w.DataDir())
	EnsureExists(w.CacheDir())
	return w
}

// --------------------------------------------------------------------
// Initialisation methods

// readInfoPlist loads the data in `info.plist`
func (wf *Workflow) readInfoPlist() error {
	if wf.infoLoaded {
		return nil
	}

	p := filepath.Join(wf.Dir(), "info.plist")
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("Couldn't open `info.plist` (%s) :  %v", p, err)
	}

	if wf.info == nil {
		wf.info = &InfoPlist{}
	}
	err = plist.Unmarshal(buf, wf.info)
	if err != nil {
		return fmt.Errorf("Error parsing `info.plist` (%s) : %v", p, err)
	}

	wf.bundleID = wf.info.BundleID
	wf.name = wf.info.Name
	if wf.version == "" { // Other options override info.plist
		wf.version = wf.info.Version
	}
	wf.infoLoaded = true
	return nil
}

// loadEnv reads Alfred's variables from the environment.
func (wf *Workflow) loadEnv() {
	wf.Env = make(map[string]string)
	// Variables currently exported by Alfred. These actual names
	// are prefixed with `alfred_`.
	keys := []string{
		"debug",
		"version",
		"version_build",
		"theme",
		"theme_background",
		"theme_selection_background",
		"theme_subtext",
		"preferences",
		"preferences_localhash",
		"workflow_cache",
		"workflow_data",
		"workflow_name",
		"workflow_uid",
		"workflow_bundleid",
		"workflow_version",
	}

	var val, envkey string

	for _, key := range keys {
		envkey = "alfred_" + key
		val = os.Getenv(envkey)
		wf.Env[key] = val

		// Some special keys
		if key == "workflow_cache" {
			wf.cacheDir = val
		} else if key == "workflow_data" {
			wf.dataDir = val
		} else if key == "workflow_bundleid" {
			wf.bundleID = val
		} else if key == "workflow_name" {
			wf.name = val
		} else if key == "debug" && val == "1" {
			wf.debug = true
		} else if key == "workflow_version" && wf.version == "" {
			wf.version = val
		}
	}
}

// initializeLogging ensures future log messages are written to
// workflow's log file.
func (wf *Workflow) initializeLogging() {

	if logInitialized { // All Workflows use the same global logger
		return
	}

	// Rotate log file if larger than MaxLogSize
	fi, err := os.Stat(wf.LogFile())
	if err == nil {
		if fi.Size() >= int64(wf.MaxLogSize) {
			new := wf.LogFile() + ".1"
			if err := os.Rename(wf.LogFile(), new); err != nil {
				fmt.Fprintf(os.Stderr, "Error rotating log: %v", err)
			}
			fmt.Fprintln(os.Stderr, "Rotated log")
		}
	}

	// Open log file
	file, err := os.OpenFile(wf.LogFile(),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		wf.Fatal(fmt.Sprintf("Couldn't open log file %s : %v",
			wf.LogFile(), err))
	}

	// Attach logger to file
	multi := io.MultiWriter(file, os.Stderr)
	log.SetOutput(multi)
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if wf.Env["debug"] == "1" {
		log.SetFlags(log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(log.Ltime)
	}

	logInitialized = true
}

// --------------------------------------------------------------------
// API methods

// Debug returns true if Alfred's debugger is open.
func Debug() bool { return wf.debug }

// Debug returns true if Alfred's debugger is open.
func (wf *Workflow) Debug() bool { return wf.debug }

// Info returns the metadata read from the workflow's info.plist.
func Info() *InfoPlist { return wf.Info() }

// Info returns the metadata read from the workflow's info.plist.
func (wf *Workflow) Info() *InfoPlist {
	if err := wf.readInfoPlist(); err != nil {
		wf.FatalError(err)
	}
	return wf.info
}

// BundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in info.plist.
func BundleID() string { return wf.BundleID() }

// BundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in info.plist.
func (wf *Workflow) BundleID() string {
	if wf.bundleID == "" { // Really old version of Alfred with no envvars?
		if err := wf.readInfoPlist(); err != nil {
			wf.FatalError(err)
		}
		if wf.bundleID == "" {
			wf.Fatal("No bundle ID set in info.plist. You *must* set a bundle ID to use awgo.")
		}
	}
	return wf.bundleID
}

// Name returns the workflow's name as specified in info.plist.
func Name() string { return wf.Name() }

// Name returns the workflow's name as specified in info.plist.
func (wf *Workflow) Name() string {
	if wf.name == "" { // Really old version of Alfred with no envvars?
		if err := wf.readInfoPlist(); err != nil {
			wf.FatalError(err)
		}
	}
	return wf.name
}

// Version returns the workflow's version from info.plist.
func Version() string { return wf.Version() }

// Version returns the workflow's version from info.plist.
func (wf *Workflow) Version() string {
	if wf.version == "" {
		if err := wf.readInfoPlist(); err != nil {
			wf.FatalError(err)
		}
	}
	return wf.version
}

// SetVersion sets the workflow's version string.
func SetVersion(v string) { wf.SetVersion(v) }

// SetVersion sets the workflow's version string.
func (wf *Workflow) SetVersion(v string) { wf.version = v }

// Dir returns the path to the workflow's root directory.
func Dir() string { return wf.Dir() }

// Dir returns the path to the workflow's root directory.
func (wf *Workflow) Dir() string {
	if wf.workflowDir == "" {
		dir, err := FindWorkflowRoot()
		if err != nil {
			wf.FatalError(err)
		}
		wf.workflowDir = dir
	}
	return wf.workflowDir
}

// Args returns command-line arguments passed to the program.
// It intercepts "magic args" and runs the corresponding actions, terminating
// the workflow.
// See MagicAction for full documentation.
func (wf *Workflow) Args() []string {
	if wf.magicPrefix != "" {
		return parseArgs(os.Args[1:], wf.magicPrefix)
	}
	return Args()
}

// --------------------------------------------------------------------
// Cache & Data

// CacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func CacheDir() string { return wf.CacheDir() }

// CacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func (wf *Workflow) CacheDir() string {
	if wf.cacheDir == "" { // Really old version of Alfred with no envvars?
		wf.cacheDir = os.ExpandEnv(fmt.Sprintf(
			"$HOME/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data/%s",
			wf.BundleID()))
	}
	return EnsureExists(wf.cacheDir)
}

// OpenCache opens the workflow's cache directory in the default application (usually Finder).
func OpenCache() error { return wf.OpenCache() }

// OpenCache opens the workflow's data directory in the default application (usually Finder).
func (wf *Workflow) OpenCache() error {
	EnsureExists(wf.DataDir())
	cmd := exec.Command("open", wf.CacheDir())
	return cmd.Run()
}

// ClearCache deletes all files from the workflow's cache directory.
func ClearCache() error { return wf.ClearCache() }

// ClearCache deletes all files from the workflow's cache directory.
func (wf *Workflow) ClearCache() error {
	return clearDirectory(wf.CacheDir())
}

// DataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func DataDir() string { return wf.DataDir() }

// DataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func (wf *Workflow) DataDir() string {
	if wf.dataDir == "" { // Really old version of Alfred with no envvars?
		wf.dataDir = os.ExpandEnv(fmt.Sprintf(
			"$HOME/Library/Application Support/Alfred 3/Workflow Data/%s",
			wf.BundleID()))
	}
	return EnsureExists(wf.dataDir)
}

// OpenData opens the workflow's data directory in the default application (usually Finder).
func OpenData() error { return wf.OpenData() }

// OpenData opens the workflow's data directory in the default application (usually Finder).
func (wf *Workflow) OpenData() error {
	EnsureExists(wf.DataDir())
	cmd := exec.Command("open", wf.DataDir())
	return cmd.Run()
}

// ClearData deletes all files from the workflow's cache directory.
func ClearData() error { return wf.ClearData() }

// ClearData deletes all files from the workflow's cache directory.
func (wf *Workflow) ClearData() error {
	return clearDirectory(wf.DataDir())
}

// Reset deletes all workflow data (cache and data directories).
func Reset() error { return wf.Reset() }

// Reset deletes all workflow data (cache and data directories).
func (wf *Workflow) Reset() error {
	errs := []error{}
	if err := wf.ClearCache(); err != nil {
		errs = append(errs, err)
	}
	if err := wf.ClearData(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// LogFile returns the path to the workflow's log file.
func LogFile() string { return wf.LogFile() }

// LogFile returns the path to the workflow's log file.
func (wf *Workflow) LogFile() string {
	return filepath.Join(wf.CacheDir(), fmt.Sprintf("%s.log", wf.BundleID()))
}

// OpenLog opens the workflow's logfile in the default application (usually Console.app).
func OpenLog() error { return wf.OpenLog() }

// OpenLog opens the workflow's logfile in the default application (usually Console.app).
func (wf *Workflow) OpenLog() error {
	if !PathExists(wf.LogFile()) {
		log.Println("Creating log file...")
	}
	cmd := exec.Command("open", wf.LogFile())
	return cmd.Run()
}

// --------------------------------------------------------------------
// Feedback

// Vars returns the workflow variables set on Workflow.Feedback.
// See Feedback.Vars() for more information.
func Vars() map[string]string { return wf.Feedback.Vars() }

// Vars returns the workflow variables set on Workflow.Feedback.
// See Feedback.Vars() for more information.
func (wf *Workflow) Vars() map[string]string {
	return wf.Feedback.Vars()
}

// Var sets the value of workflow variable k on Workflow.Feedback to v.
// See Feedback.Var() for more information.
func Var(k, v string) *Workflow { return wf.Var(k, v) }

// Var sets the value of workflow variable k on Workflow.Feedback to v.
// See Feedback.Var() for more information.
func (wf *Workflow) Var(k, v string) *Workflow {
	wf.Feedback.Var(k, v)
	return wf
}

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func NewItem(title string) *Item { return wf.NewItem(title) }

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func (wf *Workflow) NewItem(title string) *Item {
	return wf.Feedback.NewItem(title)
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func NewFileItem(path string) *Item { return wf.NewFileItem(path) }

// NewFileItem adds and returns a new feedback Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// NewWarningItem adds and returns a new Feedback Item with the system
// warning icon (exclamation mark on yellow triangle).
func NewWarningItem(title, subtitle string) *Item { return wf.NewWarningItem(title, subtitle) }

// NewWarningItem adds and returns a new Feedback Item with the system
// warning icon (exclamation mark on yellow triangle).
func (wf *Workflow) NewWarningItem(title, subtitle string) *Item {
	return wf.Feedback.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)
}

// IsEmpty returns true if Workflow contains no items.
func IsEmpty() bool { return wf.IsEmpty() }

// IsEmpty returns true if Workflow contains no items.
func (wf *Workflow) IsEmpty() bool { return len(wf.Feedback.Items) == 0 }

// Filter fuzzy-sorts feedback Items against query and deletes Items that
// don't match.
func Filter(query string) []*Result { return wf.Filter(query) }

// Filter fuzzy-sorts feedback Items against query and deletes Items that
// don't match.
func (wf *Workflow) Filter(query string) []*Result {
	return wf.Feedback.Filter(query, wf.SortOptions)
}

// Run runs your workflow function, catching any errors.
// If the workflow panics, Run rescues and displays an error
// message in Alfred.
func Run(fn func()) { wf.Run(fn) }

// Run runs your workflow function, catching any errors.
// If the workflow panics, Run rescues and displays an error
// message in Alfred.
func (wf *Workflow) Run(fn func()) {
	var vstr string
	if wf.Version() != "" {
		vstr = fmt.Sprintf("%s/%v", wf.Name(), wf.Version())
	} else {
		vstr = wf.Name()
	}
	vstr = fmt.Sprintf(" %s (awgo/%v) ", vstr, AwgoVersion)

	// Print right after Alfred's introductory blurb in the debugger.
	// Alfred strips whitespace.
	if wf.LogPrefix != "" {
		fmt.Fprintln(os.Stderr, wf.LogPrefix)
	}
	log.Println(Pad(vstr, "-", 50))

	// Catch any `panic` and display an error in Alfred.
	// Fatal(msg) will terminate the process (via log.Fatal).
	defer func() {
		if r := recover(); r != nil {
			log.Println(Pad(" FATAL ERROR ", "-", 50))
			log.Printf("%s : %s", r, debug.Stack())
			log.Println(Pad(" END STACK TRACE ", "-", 50))
			// log.Printf("Recovered : %x", r)
			err, ok := r.(error)
			if ok {
				wf.outputErrorMsg(err.Error())
			}
			wf.outputErrorMsg(fmt.Sprintf("%v", r))
		}
	}()

	// Call the workflow's main function.
	fn()

	finishLog(false)
}

// FatalError displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func FatalError(err error) { wf.FatalError(err) }

// FatalError displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) FatalError(err error) { wf.Fatal(err.Error()) }

// Fatal displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func Fatal(msg string) { wf.Fatal(msg) }

// Fatal displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) Fatal(msg string) { wf.outputErrorMsg(msg) }

// Fatalf displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func Fatalf(format string, args ...interface{}) { wf.Fatalf(format, args...) }

// Fatalf displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) Fatalf(format string, args ...interface{}) {
	wf.Fatal(fmt.Sprintf(format, args...))
}

// Warn displays a warning message in Alfred immediately. Unlike
// FatalError()/Fatal(), this does not terminate the workflow,
// but you can't send any more results to Alfred.
func Warn(title, subtitle string) *Workflow { return wf.Warn(title, subtitle) }

// Warn displays a warning message in Alfred immediately. Unlike
// FatalError()/Fatal(), this does not terminate the workflow,
// but you can't send any more results to Alfred.
func (wf *Workflow) Warn(title, subtitle string) *Workflow {
	wf.Feedback.Clear()
	wf.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)
	return wf.SendFeedback()
}

// WarnEmpty adds a warning item to feedback if there are no other items.
func WarnEmpty(title, subtitle string) { wf.WarnEmpty(title, subtitle) }

// WarnEmpty adds a warning item to feedback if there are no other items.
func (wf *Workflow) WarnEmpty(title, subtitle string) {
	if wf.IsEmpty() {
		wf.Warn(title, subtitle)
	}
}

// SendFeedback generates and sends the JSON response to Alfred.
// The JSON is output to STDOUT. At this point, Alfred considers your
// workflow complete; sending further responses will have no effect.
func SendFeedback() { wf.SendFeedback() }

// SendFeedback generates and sends the JSON response to Alfred.
// The JSON is output to STDOUT. At this point, Alfred considers your
// workflow complete; sending further responses will have no effect.
func (wf *Workflow) SendFeedback() *Workflow {
	// Truncate Items if MaxResults is set
	if wf.MaxResults > 0 && len(wf.Feedback.Items) > wf.MaxResults {
		wf.Feedback.Items = wf.Feedback.Items[0:wf.MaxResults]
	}
	if err := wf.Feedback.Send(); err != nil {
		log.Fatalf("Error generating JSON : %v", err)
	}
	return wf
}

// --------------------------------------------------------------------
// Updating

// SetUpdater sets an updater for the workflow.
func SetUpdater(u *Updater) { wf.SetUpdater(u) }

// SetUpdater sets an updater for the workflow.
func (wf *Workflow) SetUpdater(u *Updater) {
	wf.updater = u
	RegisterMagic(&updateMagic{wf.updater})
}

// UpdateCheckDue returns true if an update is available.
func UpdateCheckDue() bool { return wf.UpdateCheckDue() }

// UpdateCheckDue returns true if an update is available.
func (wf *Workflow) UpdateCheckDue() bool {
	if wf.updater == nil {
		log.Println("No GitHub repo configured")
		return false
	}
	return wf.updater.CheckDue()
}

// CheckForUpdate retrieves and caches the list of available releases.
func CheckForUpdate() error { return wf.CheckForUpdate() }

// CheckForUpdate retrieves and caches the list of available releases.
func (wf *Workflow) CheckForUpdate() error {
	if wf.updater == nil {
		return errors.New("No GitHub repo configured")
	}
	return wf.updater.CheckForUpdate()
}

// UpdateAvailable returns true if a newer version is available to install.
func UpdateAvailable() bool { return wf.UpdateAvailable() }

// UpdateAvailable returns true if a newer version is available to install.
func (wf *Workflow) UpdateAvailable() bool {
	if wf.updater == nil {
		log.Println("No GitHub repo configured")
		return false
	}
	return wf.updater.UpdateAvailable()
}

// InstallUpdate downloads and installs the latest version of the workflow.
func InstallUpdate() error { return wf.InstallUpdate() }

// InstallUpdate downloads and installs the latest version of the workflow.
func (wf *Workflow) InstallUpdate() error {
	if wf.updater == nil {
		return errors.New("No GitHub repo configured")
	}
	return wf.updater.Install()
}

// --------------------------------------------------------------------
// Helper methods

// outputErrorMsg prints and logs error, then exits process.
func (wf *Workflow) outputErrorMsg(msg string) {
	if wf.TextErrors {
		fmt.Print(msg)
	} else {
		wf.Feedback.Clear()
		wf.NewItem(msg).Icon(IconError)
		wf.SendFeedback()
	}
	log.Printf("[ERROR] %s", msg)
	// Show help URL or website URL
	if u := wf.helpURL(); u != "" {
		log.Printf("Get help at %s", u)
	}
	finishLog(true)
}

func (wf *Workflow) helpURL() string {
	if wf.HelpURL != "" {
		return wf.HelpURL
	}
	if wf.Info().Website != "" {
		return wf.Info().Website
	}
	return ""
}

// awDataDir is the directory for awgo's own data.
func awDataDir() string { return wf.awDataDir() }
func (wf *Workflow) awDataDir() string {
	return EnsureExists(filepath.Join(wf.DataDir(), "_aw"))
}

// awCacheDir is the directory for awgo's own cache.
func awCacheDir() string { return wf.awCacheDir() }
func (wf *Workflow) awCacheDir() string {
	return EnsureExists(filepath.Join(wf.CacheDir(), "_aw"))
}

// --------------------------------------------------------------------
// Package-level only

// finishLog outputs the workflow duration
func finishLog(fatal bool) {
	elapsed := time.Now().Sub(startTime)
	s := Pad(fmt.Sprintf(" %v ", elapsed), "-", 50)
	if fatal {
		log.Fatalln(s)
	} else {
		log.Println(s)
	}
}

// DefaultWorkflow returns the Workflow object used by the
// package-level functions.
func DefaultWorkflow() *Workflow { return wf }

// SetDefaultWorkflow changes the Workflow object used by the
// package-level functions.
func SetDefaultWorkflow(w *Workflow) { wf = w }
