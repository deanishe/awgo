//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import (
	"testing"
)

// TestWorkflowValues tests workflow name, bundle ID etc.
func TestWorkflowValues(t *testing.T) {

	withTestWf(func(wf *Workflow) {

		if wf.Name() != tName {
			t.Errorf("wrong name. Expected=%s, Got=%s", tName, wf.Name())
		}
		if wf.BundleID() != tBundleID {
			t.Errorf("wrong bundle ID. Expected=%s, Got=%s", tBundleID, wf.BundleID())
		}
	})
}
