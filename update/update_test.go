// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mustVersion(s string) SemVer {
	v, _ := NewSemVer(s)
	return v
}

// versioned is a test implementation of Versioned
type versioned struct {
	version string
	dir     string
}

func (v *versioned) Version() string { return v.version }
func (v *versioned) CacheDir() string {
	if v.dir == "" {
		v.dir = filepath.Join(os.TempDir(), fmt.Sprintf("aw-%d", os.Getpid()))
		os.MkdirAll(v.dir, 0700)
	}
	return v.dir
}
func (v *versioned) Clean() { os.RemoveAll(v.dir) }

// testReleaser is a test implementation of Releaser
type testReleaser struct {
	releases []*Release
}

func (r testReleaser) Releases() ([]*Release, error) {
	return r.releases, nil
}

var (
	tr, trPre *testReleaser
)

func init() {
	tr = &testReleaser{
		releases: []*Release{
			&Release{"workflow.alfredworkflow", nil, true, mustVersion("0.5.0-beta")},
			&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.1")},
			&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.4")},
			&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.2")},
			&Release{"workflow.alfredworkflow", nil, false, mustVersion("0.3")},
		},
	}
	trPre = &testReleaser{
		releases: []*Release{
			&Release{"workflow.alfredworkflow", nil, true, mustVersion("0.5.0-beta")},
			&Release{"workflow.alfredworkflow", nil, true, mustVersion("0.4.0-beta")},
			&Release{"workflow.alfredworkflow", nil, true, mustVersion("0.3.0-beta")},
		},
	}
}

func clearUpdateCache() error {
	v := &versioned{version: "0.2.2"}
	defer v.Clean()
	u, err := New(v, testReleaser{})
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
	v := &versioned{version: "0.2.2"}
	defer v.Clean()
	u, err := New(v, tr)
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

// TestUpdaterPreOnly tests that updater works with only pre-releases available
func TestUpdaterPreOnly(t *testing.T) {
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
	v := &versioned{version: "0.2.2"}
	defer v.Clean()
	u, err := New(v, trPre)
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
	u.Prereleases = true
	u.CurrentVersion = mustVersion("0.4.5")
	if !u.UpdateAvailable() {
		t.Fatal("Bad update #2")
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
	v := &versioned{version: "0.2.2"}
	defer v.Clean()
	u, err := New(v, testReleaser{})
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
	u.UpdateInterval(time.Duration(1 * time.Nanosecond))
	if !u.CheckDue() {
		t.Fatalf("Update is not due.")
	}
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
}
