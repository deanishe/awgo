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
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/deanishe/awgo/fuzzy"
	"github.com/deanishe/awgo/util"
)

// AwGoVersion is the semantic version number of this library.
const AwGoVersion = "0.14.0"

var (
	startTime time.Time // Time execution started

	// The workflow object operated on by top-level functions.
	wf *Workflow

	// Flag, as we only want to set up logging once
	// TODO: Better, more pluggable logging
	logInitialized bool
)

// init creates the default Workflow.
func init() {
	startTime = time.Now()
	wf = New()
}

// Updater can check for and download & install newer versions of the workflow.
// There is a concrete implementation and documentation in subpackage "update".
type Updater interface {
	UpdateInterval(time.Duration) // Set interval between checks
	UpdateAvailable() bool        // Return true if a newer version is available
	CheckDue() bool               // Return true if a check for a newer version is due
	CheckForUpdate() error        // Retrieve available releases, e.g. from a URL
	Install() error               // Install the latest version
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
	sync.WaitGroup
	// The response that will be sent to Alfred. Workflow provides
	// convenience wrapper methods, so you don't normally have to
	// interact with this directly.
	Feedback *Feedback

	// Workflow configuration.
	// Access workflow variables by type and save settings to info.plist.
	// See Config for documentation.
	Conf *Config

	// HelpURL is a link to your issues page/forum thread where users can
	// report bugs. It is shown in the debugger if the workflow crashes.
	HelpURL string

	// LogPrefix is the character printed to the log at the start of each run.
	// Its purpose is to ensure the first real log message starts on its own
	// line, instead of sharing a line with Alfred's blurb in the debugger.
	// This is only printed to STDERR (i.e. Alfred's debugger), not written to
	// the log file. Default: Purple Heart (\U0001F49C)
	LogPrefix string

	// MaxLogSize is the size (in bytes) at which the workflow log is rotated.
	// Default: 1 MiB
	MaxLogSize int

	// MaxResults is the maximum number of results to send to Alfred.
	// 0 means send all results.
	// Default: 0
	MaxResults int

	// SortOptions are options for fuzzy sorting.
	SortOptions []fuzzy.Option

	// TextErrors tells Workflow to print errors as text, not JSON
	// Set to true if output goes to a Notification.
	TextErrors bool

	// Cache is a Cache pointing to the workflow's cache directory.
	Cache *Cache
	// Data is a Cache pointing to the workflow's data directory.
	Data *Cache
	// Session is a cache that stores session-scoped data. These data
	// persist until the user closes Alfred or runs a different workflow.
	Session *Session

	// Updater fetches updates for the workflow.
	Updater Updater

	magicPrefix string // Overrides DefaultMagicPrefix for magic actions.

	// MagicActions contains the magic actions registered for this workflow.
	// It is set to DefaultMagicActions by default.
	MagicActions MagicActions

	// Set from environment
	// bundleID    string
	// name        string
	// cacheDir    string
	// dataDir     string
	// workflowDir string
	sessionID string
}

// New creates and initialises a new Workflow, passing any Options to Workflow.Configure().
//
// For available options, see the documentation for the Option type and the
// following functions.
func New(opts ...Option) *Workflow {

	c := NewConfig()
	if err := validateConfig(c); err != nil {
		panic(err)
	}

	wf := &Workflow{
		Conf:         c,
		Feedback:     &Feedback{},
		LogPrefix:    "\U0001F49C", // Purple heart
		MaxLogSize:   1048576,      // 1 MiB
		MaxResults:   0,            // Send all results to Alfred
		MagicActions: MagicActions{},
		SortOptions:  []fuzzy.Option{},
	}

	wf.MagicActions.Register(defaultMagicActions...)

	wf.Configure(opts...)

	wf.Cache = NewCache(wf.CacheDir())
	wf.Data = NewCache(wf.DataDir())
	wf.Session = NewSession(wf.CacheDir(), wf.SessionID())
	wf.initializeLogging()
	return wf
}

// --------------------------------------------------------------------
// Initialisation methods

// Configure applies one or more Options to Workflow. The returned Option reverts
// all Options passed to Configure.
func Configure(opts ...Option) (previous Option) { return wf.Configure(opts...) }
func (wf *Workflow) Configure(opts ...Option) (previous Option) {
	prev := make(options, len(opts))
	for i, opt := range opts {
		prev[i] = opt(wf)
	}
	return prev.apply
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

	// Show filenames and line numbers if Alfred's debugger is open
	if wf.Debug() {
		log.SetFlags(log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(log.Ltime)
	}

	logInitialized = true
}

// --------------------------------------------------------------------
// API methods

// Debug returns true if Alfred's debugger is open.
func Debug() bool                { return wf.Debug() }
func (wf *Workflow) Debug() bool { return wf.Conf.GetBool(EnvVarDebug) }

// BundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in the workflow's main
// setup sheet in Alfred Preferences.
func BundleID() string { return wf.BundleID() }
func (wf *Workflow) BundleID() string {

	s := wf.Conf.Get(EnvVarBundleID)
	if s == "" {
		wf.Fatal("No bundle ID set. You *must* set a bundle ID to use AwGo.")
	}
	return s
}

// Name returns the workflow's name as specified in the workflow's main
// setup sheet in Alfred Preferences.
func Name() string                { return wf.Name() }
func (wf *Workflow) Name() string { return wf.Conf.Get(EnvVarName) }

// Version returns the workflow's version set in the workflow's configuration
// sheet in Alfred Preferences.
func Version() string                { return wf.Version() }
func (wf *Workflow) Version() string { return wf.Conf.Get(EnvVarVersion) }

// SessionID returns the session ID for this run of the workflow.
// This is used internally for session-scoped caching.
//
// The session ID is persisted as a workflow variable. It and the session
// persist as long as the user is using the workflow in Alfred. That
// means that the session expires as soon as Alfred closes or the user
// runs a different workflow.
func SessionID() string { return wf.SessionID() }
func (wf *Workflow) SessionID() string {
	if wf.sessionID == "" {
		ev := os.Getenv("AW_SESSION_ID")
		if ev != "" {
			wf.sessionID = ev
		} else {
			wf.sessionID = NewSessionID()
		}
	}
	return wf.sessionID
}

// Args returns command-line arguments passed to the program.
// It intercepts "magic args" and runs the corresponding actions, terminating
// the workflow. See MagicAction for full documentation.
func (wf *Workflow) Args() []string {
	prefix := DefaultMagicPrefix
	if wf.magicPrefix != "" {
		prefix = wf.magicPrefix
	}
	return wf.MagicActions.Args(os.Args[1:], prefix)
}

// Run runs your workflow function, catching any errors.
// If the workflow panics, Run rescues and displays an error message in Alfred.
func Run(fn func()) { wf.Run(fn) }
func (wf *Workflow) Run(fn func()) {

	var vstr string

	if wf.Version() != "" {
		vstr = fmt.Sprintf("%s/%v", wf.Name(), wf.Version())
	} else {
		vstr = wf.Name()
	}
	vstr = fmt.Sprintf(" %s (AwGo/%v) ", vstr, AwGoVersion)

	// Print right after Alfred's introductory blurb in the debugger.
	// Alfred strips whitespace.
	if wf.LogPrefix != "" {
		fmt.Fprintln(os.Stderr, wf.LogPrefix)
	}

	log.Println(util.Pad(vstr, "-", 50))

	// Clear expired session data
	wf.Add(1)
	go func() {
		defer wf.Done()
		wf.Session.Clear(false)
	}()

	// Catch any `panic` and display an error in Alfred.
	// Fatal(msg) will terminate the process (via log.Fatal).
	defer func() {

		if r := recover(); r != nil {

			log.Println(util.Pad(" FATAL ERROR ", "-", 50))
			log.Printf("%s : %s", r, debug.Stack())
			log.Println(util.Pad(" END STACK TRACE ", "-", 50))

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

	wf.Wait()
	finishLog(false)
}

// --------------------------------------------------------------------
// Updating

// SetUpdater sets an updater for the workflow.
func SetUpdater(u Updater) { wf.SetUpdater(u) }
func (wf *Workflow) SetUpdater(u Updater) {
	wf.Updater = u
	wf.MagicActions.Register(&updateMA{wf.Updater})
}

// UpdateCheckDue returns true if an update is available.
func UpdateCheckDue() bool { return wf.UpdateCheckDue() }
func (wf *Workflow) UpdateCheckDue() bool {
	if wf.Updater == nil {
		log.Println("No updater configured")
		return false
	}
	return wf.Updater.CheckDue()
}

// CheckForUpdate retrieves and caches the list of available releases.
func CheckForUpdate() error { return wf.CheckForUpdate() }
func (wf *Workflow) CheckForUpdate() error {
	if wf.Updater == nil {
		return errors.New("No updater configured")
	}
	return wf.Updater.CheckForUpdate()
}

// UpdateAvailable returns true if a newer version is available to install.
func UpdateAvailable() bool { return wf.UpdateAvailable() }
func (wf *Workflow) UpdateAvailable() bool {
	if wf.Updater == nil {
		log.Println("No updater configured")
		return false
	}
	return wf.Updater.UpdateAvailable()
}

// InstallUpdate downloads and installs the latest version of the workflow.
func InstallUpdate() error { return wf.InstallUpdate() }
func (wf *Workflow) InstallUpdate() error {
	if wf.Updater == nil {
		return errors.New("No updater configured")
	}
	return wf.Updater.Install()
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
	if wf.HelpURL != "" {
		log.Printf("Get help at %s", wf.HelpURL)
	}
	finishLog(true)
}

// awDataDir is the directory for AwGo's own data.
func awDataDir() string { return wf.awDataDir() }
func (wf *Workflow) awDataDir() string {
	return util.MustExist(filepath.Join(wf.DataDir(), "_aw"))
}

// awCacheDir is the directory for AwGo's own cache.
func awCacheDir() string { return wf.awCacheDir() }
func (wf *Workflow) awCacheDir() string {
	return util.MustExist(filepath.Join(wf.CacheDir(), "_aw"))
}

// --------------------------------------------------------------------
// Package-level only

// finishLog outputs the workflow duration
func finishLog(fatal bool) {
	elapsed := time.Now().Sub(startTime)
	s := util.Pad(fmt.Sprintf(" %s ", util.HumanDuration(elapsed)), "-", 50)
	if fatal {
		log.Fatalln(s)
	} else {
		log.Println(s)
	}
}
