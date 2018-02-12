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
	"os/exec"
	"strings"
)

// Magic actions registered by default.
var (
	defaultMagicActions = []MagicAction{
		logMA{},        // Opens log file
		cacheMA{},      // Opens cache directory
		clearCacheMA{}, // Clears cache directory
		dataMA{},       // Opens data directory
		clearDataMA{},  // Clears data directory
		resetMA{},      // Clears cache and data directories
	}
)

/*
MagicAction is a command that is called directly by AwGo (i.e.  your workflow
code is not run) if its keyword is passed in a user query.

To use Magic Actions, it's imperative that your workflow call
Args()/Workflow.Args().

Calls to Args()/Workflow.Args() check the workflow's arguments (os.Args[1:])
for the magic prefix ("workflow:" by default), and hijack control
of the workflow if found.

If an exact keyword match is found (e.g. "workflow:log"), the corresponding
action is executed, and the workflow exits.

If no exact match is found, AwGo runs a Script Filter for the user to
select an action. Hitting TAB or RETURN on an item will run it.

Magic Actions are mainly aimed at making debugging and supporting users easier
(via the built-in actions), but they also provide a simple way to integrate
your own commands that don't need a "real" UI.

For example, setting an Updater on Workflow adds an "update" command that
checks for & installs a new version of the workflow.


Defaults

There are several built-in magic actions, which are registered by
default:

	<prefix>log         Open workflow's log file in the default app (usually
	                    Console).
	<prefix>data        Open workflow's data directory in the default app
	                    (usually Finder).
	<prefix>cache       Open workflow's data directory in the default app
	                    (usually Finder).
	<prefix>deldata     Delete everything in the workflow's data directory.
	<prefix>delcache    Delete everything in the workflow's cache directory.
	<prefix>reset       Delete everything in the workflow's data and cache directories.
	<prefix>help        Open help URL in default browser.
	                    Only registered if you have set a HelpURL.
	<prefix>update      Check for updates and install a newer version of the
	                    workflow if available.
	                    Only registered if you have configured an Updater.


Custom Actions

To add custom MagicActions, you must register them with your Workflow
*before* you call Workflow.Args()

To do this, pass MagicAction implementors to Workflow.MagicActions.Register()

*/
type MagicAction interface {
	// Keyword is what the user must enter to run the action after
	// AwGo has recognised the magic prefix. So if the prefix is
	// "workflow:" (the default), a user must enter the query
	// "workflow:<keyword>" to execute this action.
	Keyword() string

	// Description is shown when a user has entered "magic" mode, but
	// the query does not yet match a keyword.
	Description() string

	// RunText is sent to Alfred and written to the log file &
	// debugger when the action is run.
	RunText() string

	// Run is called when the Magic Action is triggered.
	Run() error
}

// MagicActions contains the registered magic actions. See the MagicAction
// interface for full documentation.
type MagicActions map[string]MagicAction

// Register adds a MagicAction to the mapping. Previous entries are overwritten.
func (ma MagicActions) Register(actions ...MagicAction) {
	for _, action := range actions {
		ma[action.Keyword()] = action
	}
}

// Unregister removes a MagicAction from the mapping (based on its keyword).
func (ma MagicActions) Unregister(actions ...MagicAction) {
	for _, action := range actions {
		delete(ma, action.Keyword())
	}
}

// Args runs a magic action or returns command-line arguments.
// It parses args for magic actions. If it finds one, it takes
// control of your workflow and runs the action.
//
// If not magic actions are found, it returns args.
func (ma MagicActions) Args(args []string, prefix string) []string {

	args, handled := ma.handleArgs(args, prefix)

	if handled {
		finishLog(false)
		os.Exit(0)
	}

	return args

}

// handleArgs checks args for the magic prefix. Returns args and true if
// it found and handled a magic argument.
func (ma MagicActions) handleArgs(args []string, prefix string) ([]string, bool) {

	var handled bool

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

				handled = true

			} else {
				for kw, action := range ma {

					NewItem(action.Keyword()).
						Subtitle(action.Description()).
						Valid(false).
						Icon(IconInfo).
						UID(action.Description()).
						Autocomplete(prefix + kw).
						Match(fmt.Sprintf("%s %s", action.Keyword(), action.Description()))
				}

				Filter(query)
				WarnEmpty("No matching action", "Try another query?")
				SendFeedback()

				handled = true
			}
		}
	}

	return args, handled
}

// Opens workflow's log file.
type logMA struct{}

func (a logMA) Keyword() string     { return "log" }
func (a logMA) Description() string { return "Open workflow's log file" }
func (a logMA) RunText() string     { return "Opening log file…" }
func (a logMA) Run() error          { return OpenLog() }

// Opens workflow's data directory.
type dataMA struct{}

func (a dataMA) Keyword() string     { return "data" }
func (a dataMA) Description() string { return "Open workflow's data directory" }
func (a dataMA) RunText() string     { return "Opening data directory…" }
func (a dataMA) Run() error          { return OpenData() }

// Opens workflow's cache directory.
type cacheMA struct{}

func (a cacheMA) Keyword() string     { return "cache" }
func (a cacheMA) Description() string { return "Open workflow's cache directory" }
func (a cacheMA) RunText() string     { return "Opening cache directory…" }
func (a cacheMA) Run() error          { return OpenCache() }

// Deletes the contents of the workflow's cache directory.
type clearCacheMA struct{}

func (a clearCacheMA) Keyword() string     { return "delcache" }
func (a clearCacheMA) Description() string { return "Delete workflow's cached data" }
func (a clearCacheMA) RunText() string     { return "Deleted workflow's cached data" }
func (a clearCacheMA) Run() error          { return ClearCache() }

// Deletes the contents of the workflow's data directory.
type clearDataMA struct{}

func (a clearDataMA) Keyword() string     { return "deldata" }
func (a clearDataMA) Description() string { return "Delete workflow's saved data" }
func (a clearDataMA) RunText() string     { return "Deleted workflow's saved data" }
func (a clearDataMA) Run() error          { return ClearData() }

// Deletes the contents of the workflow's cache & data directories.
type resetMA struct{}

func (a resetMA) Keyword() string     { return "reset" }
func (a resetMA) Description() string { return "Delete all saved and cached workflow data" }
func (a resetMA) RunText() string     { return "Deleted workflow saved and cached data" }
func (a resetMA) Run() error          { return Reset() }

// Opens URL in default browser.
type helpMA struct {
	URL string
}

func (a helpMA) Keyword() string     { return "help" }
func (a helpMA) Description() string { return "Open workflow help URL in default browser" }
func (a helpMA) RunText() string     { return "Opening help in your browser…" }
func (a helpMA) Run() error {
	cmd := exec.Command("open", a.URL)
	return cmd.Run()
}

// Updates the workflow if a newer release is available.
type updateMA struct {
	updater Updater
}

func (a updateMA) Keyword() string     { return "update" }
func (a updateMA) Description() string { return "Check for updates, and install if one is available" }
func (a updateMA) RunText() string     { return "Fetching update…" }
func (a updateMA) Run() error {
	if err := a.updater.CheckForUpdate(); err != nil {
		return err
	}
	if a.updater.UpdateAvailable() {
		return a.updater.Install()
	}
	log.Println("No update available")
	return nil
}
