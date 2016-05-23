package workflow

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/mkrautz/plist"
)

const (
	// LibVersion is the semantic version number of the Awgo library,
	// *not* the workflow.
	LibVersion = "0.2.1"
)

// The workflow object operated on by top-level functions.
// It can be retrieved/replaced with GetDefaultWorkflow() and
// SetDefaultWorkflow() respectively.
var wf *Workflow

// Info contains meta information extracted from info.plist.
// Use Workflow.Info() to retrieve the Info for the running
// workflow (it is lazily loaded).
//
// TODO: Do something meaningful with Variables.
type Info struct {
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
func (i *Info) Var(name string) string {

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
	// The version of your workflow. Use semver. The version string is
	// read from info.plist by default. This overrides that.
	Version string
}

// Workflow provides a simple, consolidated API for building Script
// Filters and talking to Alfred.
type Workflow struct {
	// The response that will be sent to Alfred. Workflow provides
	// convenience wrapper methods, so you don't have to interact
	// with this directly.
	Feedback *Feedback

	// Alfred-specific environmental variables, without the 'alfred_'
	// prefix. The following variables are present:
	//
	//	   version                 Alfred version number, e.g. "2.7"
	//     version_build           Alfred build, e.g. "277"
	//     theme                   ID of current theme, e.g.
	//                             "alfred.theme.custom.UUID-UUID-UUID"
	//     theme_background        Theme background colour in rgba format,
	//                             e.g. "rgba(255,255,255,1.00)"
	//     theme_subtext           User's subtext setting.
	//                                 "0" = Always show
	//                                 "1" = Show only for alternate actions
	//                                 "2" = Never show
	//     preferences             Path to "Alfred.alfredpreferences" file
	//     preferences_localhash   Machine-specific hash. Machine preferences
	//                             are stored in
	//                             Alfred.alfredpreferences/preferences/local/<hash>
	//     workflow_cache          Path to workflow's cache directory. Use
	//                             Workflow.GetCacheDir() instead to ensure
	//                             directory exists.
	//     workflow_data           Path to workflow's data directory. Use
	//                             Workflow.GetDataDir() instead to ensure
	//                             directory exists.
	//     workflow_name           Name of workflow, e.g. "Fast Translator"
	//     workflow_uid            Random UID assigned to workflow by Alfred
	//     workflow_bundleid       Workflow's bundle ID from info.plist
	Env map[string]string

	// Set this to your workflow's version (used in logging)
	version string

	// Populated by readInfoPlist()
	info       *Info
	infoLoaded bool

	// Set from environment or info.plist
	bundleID    string
	name        string
	cacheDir    string
	dataDir     string
	workflowDir string
}

// readInfoPlist loads the data in `info.plist`
func (wf *Workflow) readInfoPlist() error {
	if wf.infoLoaded {
		return nil
	}

	p := path.Join(wf.Dir(), "info.plist")
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("Couldn't open `info.plist` (%s) :  %v", p, err)
	}

	if wf.info == nil {
		wf.info = &Info{}
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
		"version",
		"version_build",
		"theme",
		"theme_background",
		"theme_subtext",
		"preferences",
		"preferences_localhash",
		"workflow_cache",
		"workflow_data",
		"workflow_name",
		"workflow_uid",
		"workflow_bundleid",
	}

	var val, envkey string

	for _, key := range keys {
		envkey = fmt.Sprintf("alfred_%s", key)
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
		}
	}
}

// initializeLogging ensures future log messages are written to
// workflow's log file.
func (wf *Workflow) initializeLogging() {
	// TODO: Rotate log file
	file, err := os.OpenFile(wf.LogFile(),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		wf.Fatal(fmt.Sprintf("Couldn't open log file %s : %v",
			wf.LogFile(), err))
	}

	multi := io.MultiWriter(file, os.Stderr)
	log.SetOutput(multi)
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetFlags(log.Lshortfile)
	// log.New(multi, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info returns the metadata read from the workflow's info.plist.
func (wf *Workflow) Info() *Info {
	if err := wf.readInfoPlist(); err != nil {
		wf.FatalError(err)
	}
	return wf.info
}

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
func (wf *Workflow) Name() string {
	if wf.name == "" { // Really old version of Alfred with no envvars?
		if err := wf.readInfoPlist(); err != nil {
			wf.FatalError(err)
		}
	}
	return wf.name
}

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
func (wf *Workflow) SetVersion(v string) {
	wf.version = v
}

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

// LogFile returns the path to the workflow's log file.
func (wf *Workflow) LogFile() string {
	return path.Join(wf.CacheDir(), fmt.Sprintf("%s.log", wf.BundleID()))
}

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func (wf *Workflow) NewItem() *Item {
	return wf.Feedback.NewItem()
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// NewWarningItem adds and returns a new Feedback Item with the system
// warning icon (exclamation mark on yellow triangle).
func (wf *Workflow) NewWarningItem(title, subtitle string) *Item {
	it := wf.Feedback.NewItem()
	it.Title = title
	it.Subtitle = subtitle
	it.Icon = IconWarning
	return it
}

// Run runs your workflow function, catching any errors.
func (wf *Workflow) Run(fn func()) {
	var vstr string
	startTime := time.Now()
	if wf.Version() != "" {
		vstr = fmt.Sprintf("%s/%v", wf.Name(), wf.Version())
	} else {
		vstr = wf.Name()
	}
	vstr = fmt.Sprintf(" %s (awgo/%v) ", vstr, LibVersion)

	// Print an underscore, so the log starts on the line following Alfred's
	// introductory blurb in the debugger. Alfred strips whitespace.
	fmt.Fprintln(os.Stderr, "_")
	log.Println(Pad(vstr, "-", 50))
	// log.Printf("-------- %s (awgo/%v) --------", vstr, LibVersion)

	// Catch any `panic` and display an error in Alfred.
	// SendError(Msg) will terminate the process (via log.Fatal).
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%s : %s", r, debug.Stack())
			// log.Printf("Recovered : %x", r)
			err, ok := r.(error)
			if ok {
				wf.FatalError(err)
			}
			wf.Fatal(fmt.Sprintf("%v", err))
		}
	}()

	// Call the workflow's main function.
	fn()

	elapsed := time.Now().Sub(startTime)
	log.Println(Pad(fmt.Sprintf(" %v ", elapsed), "-", 50))
}

// FatalError displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) FatalError(err error) {
	msg := fmt.Sprintf("%v", err)
	wf.Fatal(msg)
}

// Fatal displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) Fatal(errMsg string) {
	wf.Feedback.Clear()
	it := wf.NewItem()
	it.Title = errMsg
	it.Icon = IconError
	wf.SendFeedback()
	log.Fatal(errMsg)
}

// Warn displays a warning message in Alfred immediately. Unlike
// FatalError()/Fatal(), this does not terminate the workflow,
// but you can't send any more results to Alfred.
func (wf *Workflow) Warn(title, subtitle string) {
	wf.Feedback.Clear()
	it := wf.NewItem()
	it.Title = title
	it.Subtitle = subtitle
	it.Icon = IconWarning
	wf.SendFeedback()
}

// SendFeedback generates and sends the XML response to Alfred.
func (wf *Workflow) SendFeedback() {
	if err := wf.Feedback.Send(); err != nil {
		log.Fatalf("Error generating JSON : %v", err)
	}
}

// NewWorkflow creates and initialises a new Workflow.
func NewWorkflow(opts *Options) *Workflow {
	w := &Workflow{}
	// Configure workflow
	w.version = opts.Version
	w.Feedback = &Feedback{}
	w.info = &Info{}
	w.loadEnv()
	w.initializeLogging()
	return w
}

func init() {
	wf = NewWorkflow(&Options{})
}

// GetInfo returns the metadata read from the workflow's info.plist.
func GetInfo() *Info {
	return wf.Info()
}

// DefaultWorkflow returns the Workflow object used by the
// package-level functions.
func DefaultWorkflow() *Workflow {
	return wf
}

// SetDefaultWorkflow changes the Workflow object used by the
// package-level functions.
func SetDefaultWorkflow(w *Workflow) {
	wf = w
}

// Version returns the version of the workflow parsed from info.plist.
func Version() string {
	return wf.Version()
}

// SetVersion sets the version of your workflow. This is only
// used for logging, but is helpful for bug reports.
func SetVersion(v string) {
	wf.SetVersion(v)
}

// BundleID returns the bundle ID of the workflow.
// It is retrieved from Alfred's environmental variables.
func BundleID() string {
	return wf.BundleID()
}

// Name returns the name of the workflow.
func Name() string {
	return wf.Name()
}

// CacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func CacheDir() string {
	return wf.CacheDir()
}

// LogFile returns the path to the workflow's log file.
// The file may or may not exist.
func LogFile() string {
	return wf.LogFile()
}

// DataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func DataDir() string {
	return wf.DataDir()
}

// Dir returns the path to the workflow's root directory.
func Dir() string {
	return wf.Dir()
}

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func NewItem() *Item {
	return wf.NewItem()
}

// NewFileItem adds and returns an Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func NewFileItem(path string) *Item {
	return wf.NewFileItem(path)
}

// NewWarningItem adds and returns an Item with a warning icon.
func NewWarningItem(title, subtitle string) *Item {
	return wf.NewWarningItem(title, subtitle)
}

// FatalError sends an error message to Alfred as JSON feedback and
// terminates the workflow via log.Fatal().
func FatalError(err error) {
	wf.FatalError(err)
}

// Fatal sends an error message to Alfred as JSON feedback and
// terminates the workflow via log.Fatal().
func Fatal(msg string) {
	wf.Fatal(msg)
}

// Warn sends a warning message to Alfred as JSON feedback. This
// does not terminate the workflow process, but it sends the feedback
// to Alfred, so you can't send any more data to Alfred after calling
// this.
func Warn(title, subtitle string) {
	wf.Warn(title, subtitle)
}

// SendFeedback generates and sends the JSON response to Alfred.
// The JSON is output to STDOUT. At this point, Alfred considers your
// workflow complete; sending further responses will have no effect.
func SendFeedback() {
	wf.SendFeedback()
}

// Run runs your workflow function, catching any errors.
// If the workflow panics, Run rescues and displays an error
// message in Alfred.
func Run(fn func()) {
	wf.Run(fn)
}
