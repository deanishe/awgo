//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-09
//

package aw

import "github.com/deanishe/awgo/fuzzy"

// Option is a configuration option for Workflow.
// Pass one or more Options to New() or Workflow.Configure().
//
// An Option returns its inverse (i.e. an Option that restores the
// previous value).
//
// You can apply Options at any time, so you can, e.g. suppress UIDs
// if you need to for items to be in a particular order.
type Option func(wf *Workflow) Option

// options combines Options, allowing the application of their inverse
// via a single call to options.apply().
type options []Option

// apply configures Workflow with all options and returns a single Option
// to reverse all changes.
func (opts options) apply(wf *Workflow) Option {
	previous := make(options, len(opts))
	for i, opt := range opts {
		previous[i] = wf.Configure(opt)
	}
	return previous.apply
}

// HelpURL sets link shown in debugger & log if Run() catches a panic
// ("Get help at http://…").
// Set this to the URL of an issue tracker/forum thread where users can
// ask for help.
func HelpURL(URL string) Option {
	return func(wf *Workflow) Option {
		prev := wf.HelpURL
		ma := &helpMA{URL}
		if URL != "" {
			wf.MagicActions.Register(ma)
		} else {
			wf.MagicActions.Unregister(ma)
		}
		wf.HelpURL = URL
		return HelpURL(prev)
	}
}

// LogPrefix is the printed to debugger at the start of each run.
// Its purpose is to ensure that the first real log message is shown
// on its own line.
// It is only sent to Alfred's debugger, not the log file.
//
// Default: Purple Heart (\U0001F49C)
func LogPrefix(prefix string) Option {
	return func(wf *Workflow) Option {
		prev := wf.LogPrefix
		wf.LogPrefix = prefix
		return LogPrefix(prev)
	}
}

// MagicPrefix sets the prefix for "magic" commands.
// If a user enters this prefix, AwGo takes control of the workflow and
// shows a list of matching magic commands to the user.
//
// Default: workflow:
func MagicPrefix(prefix string) Option {
	return func(wf *Workflow) Option {
		prev := wf.magicPrefix
		wf.magicPrefix = prefix
		return MagicPrefix(prev)
	}
}

// MaxLogSize sets the size (in bytes) when workflow log is rotated.
// Default: 1 MiB
func MaxLogSize(bytes int) Option {
	return func(wf *Workflow) Option {
		prev := wf.MaxLogSize
		wf.MaxLogSize = bytes
		return MaxLogSize(prev)
	}
}

// MaxResults is the maximum number of results to send to Alfred.
// 0 means send all results.
// Default: 0
func MaxResults(num int) Option {
	return func(wf *Workflow) Option {
		prev := wf.MaxResults
		wf.MaxResults = num
		return MaxResults(prev)
	}
}

// TextErrors tells Workflow to print errors as text, not JSON.
// Messages are still sent to STDOUT. Set to true if error
// should be captured by Alfred, e.g. if output goes to a Notification.
func TextErrors(on bool) Option {
	return func(wf *Workflow) Option {
		prev := wf.TextErrors
		wf.TextErrors = on
		return TextErrors(prev)
	}
}

// SortOptions sets the fuzzy sorting options for Workflow.Filter().
// See fuzzy and fuzzy.Option for info on (configuring) the sorting
// algorithm.
//
// _examples/fuzzy contains an example workflow using fuzzy sort.
func SortOptions(opts ...fuzzy.Option) Option {

	return func(wf *Workflow) Option {

		prev := wf.SortOptions
		wf.SortOptions = opts

		return SortOptions(prev...)
	}
}

// SessionName changes the name of the variable used to store the session ID.
//
// This is useful if you have multiple Script Filters chained together that
// you don't want to use the same cache.
func SessionName(name string) Option {

	return func(wf *Workflow) Option {

		prev := wf.sessionName
		wf.sessionName = name

		return SessionName(prev)
	}
}

// SuppressUIDs prevents UIDs from being set on feedback Items.
//
// This turns off Alfred's knowledge, i.e. prevents Alfred from
// applying its own sort, so items will be shown in the
// order you add them.
//
// Useful if you need to force a particular item to the top/bottom.
//
// This setting only applies to Items created *after* it has been
// set.
func SuppressUIDs(on bool) Option {
	return func(wf *Workflow) Option {
		prev := wf.Feedback.NoUIDs
		wf.Feedback.NoUIDs = on
		return SuppressUIDs(prev)
	}
}

// Update sets the updater for the Workflow.
// Panics if a version number isn't set (in Alfred Preferences).
//
// See Updater interface and subpackage update for more documentation.
func Update(updater Updater) Option {
	return func(wf *Workflow) Option {
		if updater != nil && wf.Version() == "" {
			panic("can't set Updater as workflow has no version number")
		}
		prev := wf.Updater
		wf.setUpdater(updater)
		return Update(prev)
	}
}

// AddMagic registers Magic Actions with the Workflow.
// Magic Actions connect special keywords/queries to callback functions.
// See the MagicAction interface for more information.
func AddMagic(actions ...MagicAction) Option {
	return func(wf *Workflow) Option {
		for _, action := range actions {
			wf.MagicActions.Register(action)
		}
		return RemoveMagic(actions...)
	}
}

// RemoveMagic unregisters Magic Actions with Workflow.
// Magic Actions connect special keywords/queries to callback functions.
// See the MagicAction interface for more information.
func RemoveMagic(actions ...MagicAction) Option {
	return func(wf *Workflow) Option {
		for _, action := range actions {
			delete(wf.MagicActions, action.Keyword())
		}
		return AddMagic(actions...)
	}
}

// withEnv provides an alternative Env to load settings from.
func withEnv(e Env) Option {

	return func(wf *Workflow) Option {

		prev := wf.Conf
		wf.Conf = NewConfig(e)

		return withEnv(prev)
	}
}
