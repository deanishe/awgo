//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-09
//

package aw

import (
	"fmt"
	"log"

	"github.com/deanishe/awgo/fuzzy"
)

// --------------------------------------------------------------------
// Feedback

// Rerun tells Alfred to re-run the Script Filter after `secs` seconds.
func (wf *Workflow) Rerun(secs float64) *Workflow {
	wf.Feedback.Rerun(secs)
	return wf
}

// Vars returns the workflow variables set on Workflow.Feedback.
// See Feedback.Vars() for more information.
func (wf *Workflow) Vars() map[string]string {
	return wf.Feedback.Vars()
}

// Var sets the value of workflow variable k on Workflow.Feedback to v.
// See Feedback.Var() for more information.
func (wf *Workflow) Var(k, v string) *Workflow {
	wf.Feedback.Var(k, v)
	return wf
}

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func (wf *Workflow) NewItem(title string) *Item {
	return wf.Feedback.NewItem(title)
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// NewWarningItem adds and returns a new Feedback Item with the system
// warning icon (exclamation mark on yellow triangle).
func (wf *Workflow) NewWarningItem(title, subtitle string) *Item {

	return wf.Feedback.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)
}

// IsEmpty returns true if Workflow contains no items.
func (wf *Workflow) IsEmpty() bool { return len(wf.Feedback.Items) == 0 }

// FatalError displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) FatalError(err error) { wf.Fatal(err.Error()) }

// Fatal displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) Fatal(msg string) { wf.outputErrorMsg(msg) }

// Fatalf displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func (wf *Workflow) Fatalf(format string, args ...interface{}) {
	wf.Fatal(fmt.Sprintf(format, args...))
}

// Warn displays a warning message in Alfred immediately. Unlike
// FatalError()/Fatal(), this does not terminate the workflow,
// but you can't send any more results to Alfred.
func (wf *Workflow) Warn(title, subtitle string) *Workflow {

	// Remove any existing items
	wf.Feedback.Clear()

	wf.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)

	return wf.SendFeedback()
}

// WarnEmpty adds a warning item to feedback if there are no other items.
func (wf *Workflow) WarnEmpty(title, subtitle string) {
	if wf.IsEmpty() {
		wf.Warn(title, subtitle)
	}
}

// Filter fuzzy-sorts feedback Items against query and deletes Items that don't match.
func (wf *Workflow) Filter(query string) []*fuzzy.Result {
	return wf.Feedback.Filter(query, wf.sortOptions...)
}

// SendFeedback sends Script Filter results to Alfred.
//
// Results are output as JSON to STDOUT. As you can output results only once,
// subsequent calls to sending methods are logged and ignored.
//
// The sending methods are:
//
//     SendFeedback()
//     Fatal()
//     Fatalf()
//     FatalError()
//     Warn()
//     WarnEmpty()  // only sends if there are no items
//
func (wf *Workflow) SendFeedback() *Workflow {

	// Set session ID
	wf.Var("AW_SESSION_ID", wf.SessionID())

	// Truncate Items if maxResults is set
	if wf.maxResults > 0 && len(wf.Feedback.Items) > wf.maxResults {
		wf.Feedback.Items = wf.Feedback.Items[0:wf.maxResults]
	}

	if err := wf.Feedback.Send(); err != nil {
		log.Fatalf("Error generating JSON : %v", err)
	}

	return wf
}
