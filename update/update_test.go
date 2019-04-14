// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
		var err error
		if v.dir, err = ioutil.TempDir("", "aw-"); err != nil {
			panic(err)
		}
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

type testFailReleaser struct{}

func (r testFailReleaser) Releases() ([]*Release, error) {
	return nil, errors.New("failed")
}

var (
	tr, trPre *testReleaser
)

func withVersioned(version string, fn func(v *versioned)) {
	v := &versioned{version: version}
	defer v.Clean()
	fn(v)
}

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

func TestUpdater(t *testing.T) {
	t.Parallel()

	withVersioned("0.2.2", func(v *versioned) {
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
	})
}

// TestUpdaterPreOnly tests that updater works with only pre-releases available
func TestUpdaterPreOnly(t *testing.T) {
	t.Parallel()

	withVersioned("0.2.2", func(v *versioned) {
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
	})
}

// TestUpdateInterval tests caching of LastCheck.
func TestUpdateInterval(t *testing.T) {
	t.Parallel()
	t.Run("UpdateIntervalOnSuccess", testUpdateIntervalSuccess)
	t.Run("UpdateIntervalOnFailure", testUpdateIntervalFail)
}

func testUpdateIntervalSuccess(t *testing.T) {
	t.Parallel()
	withVersioned("0.2.2", func(v *versioned) {
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
	})
}

func testUpdateIntervalFail(t *testing.T) {
	t.Parallel()
	withVersioned("0.2.2", func(v *versioned) {
		u, err := New(v, testFailReleaser{})
		if err != nil {
			t.Fatalf("Error creating updater: %s", err)
		}

		// UpdateInterval is set
		if !u.LastCheck.IsZero() {
			t.Fatal("LastCheck is not zero.")
		}
		if !u.CheckDue() {
			t.Fatal("Update is not due.")
		}
		// LastCheck is updated
		if err := u.CheckForUpdate(); err == nil {
			t.Fatal("Fetch succeeded")
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
	})
}

func TestHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("HTTP(hello)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello")
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		data, err := getURL(u)
		if err != nil {
			t.Fatal(err)
		}
		ts.Close()

		s := string(data)
		if s != "hello\n" {
			t.Errorf("Expected=%q, Got=%q", "hello", s)
		}
	})

	t.Run("HTTP(404)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		_, err := getURL(u)
		if err == nil {
			t.Errorf("404 request succeeded")
		}
		ts.Close()
	})

	t.Run("HTTP(download)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "contents")
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)
		f, err := ioutil.TempFile("", "awgo-*-test")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		err = download(u, f.Name())
		if err != nil {
			t.Fatalf("[ERROR] download: %v", err)
		}

		data, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("[ERROR] open file: %v", err)
		}
		s := string(data)
		if s != "contents\n" {
			t.Errorf("Expected=%q, Got=%q", "contents\n", s)
		}
	})
}
