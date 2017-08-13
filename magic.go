//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-05
//

package aw

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// DefaultMagicPrefix is the default prefix for "magic" arguments.
// This can be overriden with the MagicPrefix value in Options.
const DefaultMagicPrefix = "workflow:"

var (
	// DefaultMagicActions are magic actions registered by default.
	DefaultMagicActions = []MagicAction{
		openLogMagic{},    // Opens log file
		openCacheMagic{},  // Opens cache directory
		clearCacheMagic{}, // Clears cache directory
		openDataMagic{},   // Opens data directory
		clearDataMagic{},  // Clears data directory
		resetMagic{},      // Clears cache and data directories
	}
)

// MagicActions contains the registered magic actions. See the MagicAction
// interface for full documentation.
type MagicActions map[string]MagicAction

// Register adds a MagicArgument to the mapping. Previous entries are overwritten.
func (ma MagicActions) Register(actions ...MagicAction) {
	for _, action := range actions {
		ma[action.Keyword()] = action
	}
}

// Args runs a magic action or returns command-line arguments.
// It parses args for magic actions. If it finds one, it takes
// control of your workflow and runs the action.
//
// If not magic actions are found, it returns args.
func (ma MagicActions) Args(args []string, prefix string) []string {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, prefix) {
			query := arg[len(prefix):]
			action := ma[query]
			if action != nil {
				log.Printf(action.RunText())
				NewItem(action.RunText()).
					Icon(IconInfo).
					Valid(false)
				SendFeedback()
				if err := action.Run(); err != nil {
					log.Printf("Error running magic arg `%s`: %s", action.Description(), err)
					finishLog(true)
				}
				finishLog(false)
				os.Exit(0)
			} else {
				for kw, action := range ma {
					NewItem(action.Keyword()).
						Subtitle(action.Description()).
						Valid(false).
						Icon(IconInfo).
						UID(action.Description()).
						Autocomplete(prefix + kw).
						SortKey(fmt.Sprintf("%s %s", action.Keyword(), action.Description()))
				}
				Filter(query)
				WarnEmpty("No matching action", "Try another query?")
				SendFeedback()
				os.Exit(0)
			}
		}
	}
	return args
}

// MagicAction is a command that can be called directly by AwGo if its
// keyword is passed in a user query. These are mainly aimed at making
// debugging and supporting users easier: the built-in actions open the
// log file and data/cache directories, and can also clear them.
// This saves users (and developers) from having to mess around in Finder
// to dig out files buried somewhere deep in ~/Library.
//
// If you call Args() or Workflow.Args(), they return os.Args[1:], but
// first check if any argument starts with the "magic" prefix ("workflow:")
// by default.
//
// If so, AwGo will take control of the workflow (i.e. your code will no
// longer be run) and run its own "magic" mode. In this mode, it checks
// if the rest of the user query matches the keyword for a registered
// MagicAction, and if so, it runs that action, displaying RunText() in
// Alfred (if it's a Script Filter) and the log & debugger.
//
// If no keyword matches, AwGo sends a list of available magic actions
// to Alfred, filtered by the user's query. Hitting TAB or RETURN on
// an item will run it.
//
// The built-in magic actions are:
//
//    Keyword           | Action
//    --------------------------------------------------------------------------------------
//    <prefix>log       | Open workflow's log file in the default app (usually Console.log).
//    <prefix>data      | Open workflow's data directory in the default app (usually Finder).
//    <prefix>cache     | Open workflow's data directory in the default app (usually Finder).
//    <prefix>deldata   | Delete everything in the workflow's data directory.
//    <prefix>delcache  | Delete everything in the workflow's cache directory.
//    <prefix>reset     | Delete everything in the workflow's data and cache directories.
//    <prefix>update    | Check for updates and install a newer version of the workflow
//                      | if available.
//                      | Only registered if you have set an Updater via SetUpdater()
//                      | or the GitHub value in Options.
//
type MagicAction interface {
	// Keyword is what the user must enter to run the action after
	// AwGo has recognised the magic prefix.
	Keyword() string
	// Description is shown when a user has entered "magic" mode, but
	// the query does not yet match a keyword.
	Description() string
	// RunText is sent to Alfred and written to the log & debugger when
	// the action is run.
	RunText() string
	// Run executes the magic action.
	Run() error
}

// Opens workflow's log file.
type openLogMagic struct{}

func (a openLogMagic) Keyword() string     { return "log" }
func (a openLogMagic) Description() string { return "Open workflow's log file" }
func (a openLogMagic) RunText() string     { return "Opening log file…" }
func (a openLogMagic) Run() error          { return OpenLog() }

// Opens workflow's data directory.
type openDataMagic struct{}

func (a openDataMagic) Keyword() string     { return "data" }
func (a openDataMagic) Description() string { return "Open workflow's data directory" }
func (a openDataMagic) RunText() string     { return "Opening data directory…" }
func (a openDataMagic) Run() error          { return OpenData() }

// Opens workflow's cache directory.
type openCacheMagic struct{}

func (a openCacheMagic) Keyword() string     { return "cache" }
func (a openCacheMagic) Description() string { return "Open workflow's cache directory" }
func (a openCacheMagic) RunText() string     { return "Opening cache directory…" }
func (a openCacheMagic) Run() error          { return OpenCache() }

// Deletes the contents of the workflow's cache directory.
type clearCacheMagic struct{}

func (a clearCacheMagic) Keyword() string     { return "delcache" }
func (a clearCacheMagic) Description() string { return "Delete workflow's cached data" }
func (a clearCacheMagic) RunText() string     { return "Deleted workflow's cached data" }
func (a clearCacheMagic) Run() error          { return ClearCache() }

// Deletes the contents of the workflow's data directory.
type clearDataMagic struct{}

func (a clearDataMagic) Keyword() string     { return "deldata" }
func (a clearDataMagic) Description() string { return "Delete workflow's saved data" }
func (a clearDataMagic) RunText() string     { return "Deleted workflow's saved data" }
func (a clearDataMagic) Run() error          { return ClearData() }

// Deletes the contents of the workflow's cache & data directories.
type resetMagic struct{}

func (a resetMagic) Keyword() string     { return "reset" }
func (a resetMagic) Description() string { return "Delete all saved and cached workflow data" }
func (a resetMagic) RunText() string     { return "Deleted workflow saved and cached data" }
func (a resetMagic) Run() error          { return Reset() }

// Updates the workflow if a newer release is available.
type updateMagic struct {
	updater Updater
}

func (a updateMagic) Keyword() string     { return "update" }
func (a updateMagic) Description() string { return "Check for updates, and install if one is available" }
func (a updateMagic) RunText() string     { return "Fetching update…" }
func (a updateMagic) Run() error {
	if err := a.updater.CheckForUpdate(); err != nil {
		return err
	}
	if a.updater.UpdateAvailable() {
		return a.updater.Install()
	}
	log.Println("No update available")
	return nil
}
