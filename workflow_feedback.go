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
func Rerun(secs float64) *Workflow { return wf.Rerun(secs) }
func (wf *Workflow) Rerun(secs float64) *Workflow {
	wf.Feedback.Rerun(secs)
	return wf
}

// Vars returns the workflow variables set on Workflow.Feedback.
// See Feedback.Vars() for more information.
func Vars() map[string]string { return wf.Vars() }
func (wf *Workflow) Vars() map[string]string {
	return wf.Feedback.Vars()
}

// Var sets the value of workflow variable k on Workflow.Feedback to v.
// See Feedback.Var() for more information.
func Var(k, v string) *Workflow { return wf.Var(k, v) }
func (wf *Workflow) Var(k, v string) *Workflow {
	wf.Feedback.Var(k, v)
	return wf
}

// NewItem adds and returns a new feedback Item.
// See Feedback.NewItem() for more information.
func NewItem(title string) *Item { return wf.NewItem(title) }
func (wf *Workflow) NewItem(title string) *Item {
	return wf.Feedback.NewItem(title)
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
// See Feedback.NewFileItem() for more information.
func NewFileItem(path string) *Item { return wf.NewFileItem(path) }
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// NewWarningItem adds and returns a new Feedback Item with the system
// warning icon (exclamation mark on yellow triangle).
func NewWarningItem(title, subtitle string) *Item { return wf.NewWarningItem(title, subtitle) }
func (wf *Workflow) NewWarningItem(title, subtitle string) *Item {
	return wf.Feedback.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)
}

// IsEmpty returns true if Workflow contains no items.
func IsEmpty() bool                { return wf.IsEmpty() }
func (wf *Workflow) IsEmpty() bool { return len(wf.Feedback.Items) == 0 }

// FatalError displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func FatalError(err error)                { wf.FatalError(err) }
func (wf *Workflow) FatalError(err error) { wf.Fatal(err.Error()) }

// Fatal displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func Fatal(msg string)                { wf.Fatal(msg) }
func (wf *Workflow) Fatal(msg string) { wf.outputErrorMsg(msg) }

// Fatalf displays an error message in Alfred, then calls log.Fatal(),
// terminating the workflow.
func Fatalf(format string, args ...interface{}) { wf.Fatalf(format, args...) }
func (wf *Workflow) Fatalf(format string, args ...interface{}) {
	wf.Fatal(fmt.Sprintf(format, args...))
}

// Warn displays a warning message in Alfred immediately. Unlike
// FatalError()/Fatal(), this does not terminate the workflow,
// but you can't send any more results to Alfred.
func Warn(title, subtitle string) *Workflow { return wf.Warn(title, subtitle) }
func (wf *Workflow) Warn(title, subtitle string) *Workflow {
	wf.Feedback.Clear()
	wf.NewItem(title).
		Subtitle(subtitle).
		Icon(IconWarning)
	return wf.SendFeedback()
}

// WarnEmpty adds a warning item to feedback if there are no other items.
func WarnEmpty(title, subtitle string) { wf.WarnEmpty(title, subtitle) }
func (wf *Workflow) WarnEmpty(title, subtitle string) {
	if wf.IsEmpty() {
		wf.Warn(title, subtitle)
	}
}

// Filter fuzzy-sorts feedback Items against query and deletes Items that don't match.
func Filter(query string) []*fuzzy.Result { return wf.Filter(query) }
func (wf *Workflow) Filter(query string) []*fuzzy.Result {
	return wf.Feedback.Filter(query, wf.SortOptions...)
}

// SendFeedback sends Script Filter results to Alfred.
//
// Results are output as JSON to STDOUT. As you can output results only once, subsequent calls to sending methods are logged and ignored.
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
func SendFeedback() { wf.SendFeedback() }
func (wf *Workflow) SendFeedback() *Workflow {
	// Set session ID
	wf.Var("AW_SESSION_ID", wf.SessionID())
	// Truncate Items if MaxResults is set
	if wf.MaxResults > 0 && len(wf.Feedback.Items) > wf.MaxResults {
		wf.Feedback.Items = wf.Feedback.Items[0:wf.MaxResults]
	}
	if err := wf.Feedback.Send(); err != nil {
		log.Fatalf("Error generating JSON : %v", err)
	}
	return wf
}
