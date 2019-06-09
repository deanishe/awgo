// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func mustRead(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func mustVersion(s string) SemVer {
	v, _ := NewSemVer(s)
	return v
}

// versioned is a test implementation of Versioned
type versioned struct {
	version string
}

func (v *versioned) Version() string { return v.version }

type testSource struct {
	dls []Download
}

func (src testSource) Downloads() ([]Download, error) { return src.dls, nil }

type testFailSource struct{}

func (src testFailSource) Downloads() ([]Download, error) { return nil, errors.New("fail") }

func withTempDir(fn func(dir string)) {
	dir, err := ioutil.TempDir("", "aw-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	fn(dir)
}

var (
	testSrc1 = &testSource{
		dls: []Download{
			Download{Version: mustVersion("0.5.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.1"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.4"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.2"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.3"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
		},
	}
	testSrc2 = &testSource{
		dls: []Download{
			Download{Version: mustVersion("0.5.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.4.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			Download{Version: mustVersion("0.3.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
		},
	}
)

func TestUpdater(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {

		u, err := NewUpdater(testSrc1, "0.2.2", dir)
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

	withTempDir(func(dir string) {
		u, err := NewUpdater(testSrc2, "0.2.2", dir)
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
	withTempDir(func(dir string) {
		u, err := NewUpdater(testSource{}, "0.2.2", dir)
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
		u.updateInterval = time.Duration(1 * time.Nanosecond)
		if !u.CheckDue() {
			t.Fatalf("Update is not due.")
		}
	})
}

func testUpdateIntervalFail(t *testing.T) {
	t.Parallel()
	withTempDir(func(dir string) {
		u, err := NewUpdater(testFailSource{}, "0.2.2", dir)
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
		u.updateInterval = time.Duration(1 * time.Nanosecond)
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

		data, err := getURL(ts.URL)
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

		_, err := getURL(ts.URL)
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

		f, err := ioutil.TempFile("", "awgo-*-test")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		err = download(ts.URL, f.Name())
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
