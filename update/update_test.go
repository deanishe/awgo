//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-03
//

package update

import "testing"
import "time"
import "fmt"

func mustVersion(s string) SemVer {
	v, _ := NewSemVer(s)
	return v
}

type testReleaser struct {
	releases []*Release
}

func (r testReleaser) Releases() ([]*Release, error) {
	return []*Release{
		&Release{"workflow.alfredworkflow", nil, true, mustVersion("0.5.0-beta")},
		&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.1")},
		&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.4")},
		&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.2")},
		&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.3")},
	}, nil
}

func clearUpdateCache() error {
	u, err := New(testReleaser{})
	if err != nil {
		return fmt.Errorf("Error creating updater: %s", err)
	}
	u.clearCache()
	return nil
}

func TestUpdater(t *testing.T) {
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
	rl := testReleaser{}
	u, err := New(rl)
	if err != nil {
		t.Fatalf("Error creating updater: %s", err)
	}
	if err := u.CheckForUpdate(); err != nil {
		t.Fatalf("Error getting releases: %s", err)
	}
	u.CurrentVersion = mustVersion("1")
	if u.UpdateAvailable() {
		t.Fatal("Bad update #1")
	}
	u.CurrentVersion = mustVersion("0.5")
	if u.UpdateAvailable() {
		t.Fatal("Bad update #2")
	}
	u.CurrentVersion = mustVersion("0.4.5")
	if u.UpdateAvailable() {
		t.Fatal("Bad update #3")
	}
	u.Prereleases = true
	u.CurrentVersion = mustVersion("0.4.5")
	if !u.UpdateAvailable() {
		t.Fatal("Bad update #4")
	}
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
}

// TestUpdateInterval tests caching of LastCheck.
func TestUpdateInterval(t *testing.T) {
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
	u, err := New(testReleaser{})
	if err != nil {
		t.Fatalf("Error creating updater: %s", err)
	}

	// UpdateInterval is set
	if !u.LastCheck.IsZero() {
		t.Fatalf("LastCheck is not zero.")
	}
	if !u.CheckDue() {
		t.Fatalf("Update is not due.")
	}
	// LastCheck is updated
	if err := u.CheckForUpdate(); err != nil {
		t.Fatalf("Error fetching releases: %s", err)
	}
	if u.LastCheck.IsZero() {
		t.Fatalf("LastCheck is zero.")
	}
	if u.CheckDue() {
		t.Fatalf("Update is due.")
	}
	// Changing UpdateInterval
	u.UpdateInterval = time.Duration(1 * time.Nanosecond)
	if !u.CheckDue() {
		t.Fatalf("Update is not due.")
	}
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
}
