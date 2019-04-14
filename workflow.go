// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
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
const AwGoVersion = "0.16.1"

// Default Workflow settings. Can be changed with the corresponding Options.
//
// See the Options and Workflow documentation for more information.
const (
	DefaultLogPrefix   = "\U0001F37A"    // Beer mug
	DefaultMaxLogSize  = 1048576         // 1 MiB
	DefaultMaxResults  = 0               // No limit, i.e. send all results to Alfred
	DefaultSessionName = "AW_SESSION_ID" // Workflow variable session ID is stored in
	DefaultMagicPrefix = "workflow:"     // Prefix to call "magic" actions
)

var (
	startTime time.Time // Time execution started

	// The workflow object operated on by top-level functions.
	// wf *Workflow

	// Flag, as we only want to set up logging once
	// TODO: Better, more pluggable logging
	logInitialized bool
)

// init creates the default Workflow.
func init() {
	startTime = time.Now()
}

// Workflow provides a consolidated API for building Script Filters.
//
// As a rule, you should create a Workflow in init or main and call your main
// entry-point via Workflow.Run(), which catches panics, and logs & shows the
// error in Alfred.
//
// Script Filter
//
// To generate feedback for a Script Filter, use Workflow.NewItem() to create
// new Items and Workflow.SendFeedback() to send the results to Alfred.
//
// Run Script
//
// Use the TextErrors option, so any rescued panics are printed as text,
// not as JSON.
//
// Use ArgVars to set workflow variables, not Workflow/Feedback.
//
// See the _examples/ subdirectory for some full examples of workflows.
type Workflow struct {
	sync.WaitGroup
	// Interface to workflow's settings.
	// Reads workflow variables by type and saves new values to info.plist.
	Config *Config

	// Call Alfred's AppleScript functions.
	Alfred *Alfred

	// Cache is a Cache pointing to the workflow's cache directory.
	Cache *Cache
	// Data is a Cache pointing to the workflow's data directory.
	Data *Cache
	// Session is a cache that stores session-scoped data. These data
	// persist until the user closes Alfred or runs a different workflow.
	Session *Session

	// The response that will be sent to Alfred. Workflow provides
	// convenience wrapper methods, so you don't normally have to
	// interact with this directly.
	Feedback *Feedback

	// Updater fetches updates for the workflow.
	Updater Updater

	// MagicActions contains the magic actions registered for this workflow.
	// Several built-in actions are registered by default. See the docs for
	// MagicAction for details.
	MagicActions *MagicActions

	logPrefix   string         // Written to debugger to force a newline
	maxLogSize  int            // Maximum size of log file in bytes
	magicPrefix string         // Overrides DefaultMagicPrefix for magic actions.
	maxResults  int            // max. results to send to Alfred. 0 means send all.
	sortOptions []fuzzy.Option // Options for fuzzy filtering
	textErrors  bool           // Show errors as plaintext, not Alfred JSON
	helpURL     string         // URL to help page (shown if there's an error)
	dir         string         // Directory workflow is in
	cacheDir    string         // Workflow's cache directory
	dataDir     string         // Workflow's data directory
	sessionName string         // Name of the variable sessionID is stored in
	sessionID   string         // Random session ID
}

// New creates and initialises a new Workflow, passing any Options to
// Workflow.Configure().
//
// For available options, see the documentation for the Option type and the
// following functions.
//
// IMPORTANT: In order to be able to initialise the Workflow correctly,
// New must be run within a valid Alfred environment; specifically
// *at least* the following environment variables must be set:
//
//     alfred_workflow_bundleid
//     alfred_workflow_cache
//     alfred_workflow_data
//
// If you aren't running from Alfred, or would like to specify a
// custom environment, use NewFromEnv().
func New(opts ...Option) *Workflow { return NewFromEnv(nil, opts...) }

// NewFromEnv creates a new Workflows from the specified Env.
// If env is nil, the system environment is used.
func NewFromEnv(env Env, opts ...Option) *Workflow {

	if env == nil {
		env = sysEnv{}
	}

	if err := validateEnv(env); err != nil {
		panic(err)
	}

	wf := &Workflow{
		Config:      NewConfig(env),
		Alfred:      NewAlfred(env),
		Feedback:    &Feedback{},
		logPrefix:   DefaultLogPrefix,
		maxLogSize:  DefaultMaxLogSize,
		maxResults:  DefaultMaxResults,
		sessionName: DefaultSessionName,
		sortOptions: []fuzzy.Option{},
	}

	wf.MagicActions = defaultMagicActions(wf)

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

		if fi.Size() >= int64(wf.maxLogSize) {

			new := wf.LogFile() + ".1"
			if err := os.Rename(wf.LogFile(), new); err != nil {
				fmt.Fprintf(os.Stderr, "Error rotating log: %v\n", err)
			}

			fmt.Fprintln(os.Stderr, "Rotated log")
		}
	}

	// Open log file
	file, err := os.OpenFile(wf.LogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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

// BundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in the workflow's main
// setup sheet in Alfred Preferences.
func (wf *Workflow) BundleID() string {

	s := wf.Config.Get(EnvVarBundleID)
	if s == "" {
		wf.Fatal("No bundle ID set. You *must* set a bundle ID to use AwGo.")
	}
	return s
}

// Name returns the workflow's name as specified in the workflow's main
// setup sheet in Alfred Preferences.
func (wf *Workflow) Name() string { return wf.Config.Get(EnvVarName) }

// Version returns the workflow's version set in the workflow's configuration
// sheet in Alfred Preferences.
func (wf *Workflow) Version() string { return wf.Config.Get(EnvVarVersion) }

// SessionID returns the session ID for this run of the workflow.
// This is used internally for session-scoped caching.
//
// The session ID is persisted as a workflow variable. It and the session
// persist as long as the user is using the workflow in Alfred. That
// means that the session expires as soon as Alfred closes or the user
// runs a different workflow.
func (wf *Workflow) SessionID() string {

	if wf.sessionID == "" {

		ev := os.Getenv(wf.sessionName)

		if ev != "" {
			wf.sessionID = ev
		} else {
			wf.sessionID = NewSessionID()
		}
	}

	return wf.sessionID
}

// Debug returns true if Alfred's debugger is open.
func (wf *Workflow) Debug() bool { return wf.Config.GetBool(EnvVarDebug) }

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
func (wf *Workflow) Run(fn func()) {

	vstr := wf.Name()

	if wf.Version() != "" {
		vstr += "/" + wf.Version()
	}

	vstr = fmt.Sprintf(" %s (AwGo/%v) ", vstr, AwGoVersion)

	// Print right after Alfred's introductory blurb in the debugger.
	// Alfred strips whitespace.
	if wf.logPrefix != "" {
		fmt.Fprintln(os.Stderr, wf.logPrefix)
	}

	log.Println(util.Pad(vstr, "-", 50))

	// Clear expired session data
	wf.Add(1)
	go func() {
		defer wf.Done()
		if err := wf.Session.Clear(false); err != nil {
			log.Printf("[ERROR] clear session: %v", err)
		}
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
// Helper methods

// outputErrorMsg prints and logs error, then exits process.
func (wf *Workflow) outputErrorMsg(msg string) {
	if wf.textErrors {
		fmt.Print(msg)
	} else {
		wf.Feedback.Clear()
		wf.NewItem(msg).Icon(IconError)
		wf.SendFeedback()
	}
	log.Printf("[ERROR] %s", msg)
	// Show help URL or website URL
	if wf.helpURL != "" {
		log.Printf("Get help at %s", wf.helpURL)
	}
	finishLog(true)
}

// awDataDir is the directory for AwGo's own data.
func (wf *Workflow) awDataDir() string {
	return util.MustExist(filepath.Join(wf.DataDir(), "_aw"))
}

// awCacheDir is the directory for AwGo's own cache.
func (wf *Workflow) awCacheDir() string {
	return util.MustExist(filepath.Join(wf.CacheDir(), "_aw"))
}

// --------------------------------------------------------------------
// Package-level only

// finishLog outputs the workflow duration
func finishLog(fatal bool) {

	elapsed := time.Now().Sub(startTime)
	s := util.Pad(fmt.Sprintf(" %v ", elapsed), "-", 50)

	if fatal {
		log.Fatalln(s)
	} else {
		log.Println(s)
	}
}
