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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock exec.Command
type mockExec struct {
	name string
	args []string
}

func (me *mockExec) Run(name string, arg ...string) error {
	me.name = name
	me.args = append([]string{name}, arg...)
	return nil
}

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
	defer panicOnError(os.RemoveAll(dir))
	fn(dir)
}

var (
	testSrc1 = &testSource{
		dls: []Download{
			{Version: mustVersion("0.5.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.1"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.4"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.2"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.3"), Prerelease: false, Filename: "Dummy.alfredworkflow"},
		},
	}
	testSrc2 = &testSource{
		dls: []Download{
			{Version: mustVersion("0.5.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.4.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
			{Version: mustVersion("0.3.0-beta"), Prerelease: true, Filename: "Dummy.alfredworkflow"},
		},
	}
)

func TestUpdater(t *testing.T) {
	withTempDir(func(dir string) {
		vStr := "4.0.4"
		oldVal := os.Getenv("alfred_version")
		defer panicOnError(os.Setenv("alfred_version", oldVal))

		panicOnError(os.Setenv("alfred_version", vStr))

		u, err := NewUpdater(testSrc1, "0.2.2", dir)
		assert.Nil(t, err, "create updater failed")
		assert.Nil(t, u.CheckForUpdate(), "get releases failed")

		u.CurrentVersion = mustVersion("1")
		assert.False(t, u.UpdateAvailable(), "unexpected update")
		u.CurrentVersion = mustVersion("0.5")
		assert.False(t, u.UpdateAvailable(), "unexpected update")
		u.CurrentVersion = mustVersion("0.4.5")
		assert.False(t, u.UpdateAvailable(), "unexpected update")

		u.Prereleases = true
		assert.True(t, u.UpdateAvailable(), "unexpected update")

		sv, _ := NewSemVer(vStr)
		assert.True(t, sv.Eq(u.AlfredVersion), "unexpected Alfred version")

		// Empty cache directory
		_, err = NewUpdater(testSrc1, "0.2.2", "")
		assert.NotNil(t, err, "Updater accepted empty cacheDir")
	})
}

// TestUpdaterPreOnly tests that updater works with only pre-releases available
func TestUpdaterPreOnly(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {
		u, err := NewUpdater(testSrc2, "0.2.2", dir)
		require.Nil(t, err, "create updater failed")
		require.Nil(t, u.CheckForUpdate(), "get releases failed")

		u.CurrentVersion = mustVersion("1")
		assert.False(t, u.UpdateAvailable(), "unexpected update")

		u.Prereleases = true
		u.CurrentVersion = mustVersion("0.4.5")
		assert.True(t, u.UpdateAvailable(), "unexpected update")
	})
}

// TestUpdateInterval tests caching of LastCheck.
func TestUpdateInterval(t *testing.T) {
	t.Parallel()
	t.Run("UpdateIntervalOnSuccess", func(t *testing.T) {
		t.Parallel()
		testUpdateInterval(testSource{}, false, t)
	})
	t.Run("UpdateIntervalOnFailure", func(t *testing.T) {
		t.Parallel()
		testUpdateInterval(testFailSource{}, true, t)
	})
}

func testUpdateInterval(src Source, fail bool, t *testing.T) {
	withTempDir(func(dir string) {
		u, err := NewUpdater(src, "0.2.2", dir)
		require.Nil(t, err, "create updater failed")

		// UpdateInterval is set
		assert.True(t, u.LastCheck.IsZero(), "LastCheck is not zero")
		assert.True(t, u.CheckDue(), "update check is not due")

		// LastCheck is updated
		if fail {
			assert.NotNil(t, u.CheckForUpdate(), "fetch releases succeeded")
		} else {
			assert.Nil(t, u.CheckForUpdate(), "fetch releases failed")
		}
		assert.False(t, u.LastCheck.IsZero(), "LastCheck is zero")
		assert.False(t, u.CheckDue(), "update check is due")

		// Changing UpdateInterval
		u.updateInterval = time.Nanosecond
		assert.True(t, u.CheckDue(), "update check is not due")
	})
}

func TestUpdater_Install(t *testing.T) {
	origRun := runCommand
	origDownload := download
	defer func() {
		runCommand = origRun
		download = origDownload
	}()

	me := &mockExec{}
	runCommand = me.Run
	download = func(URL, path string) error { return nil }

	withTempDir(func(dir string) {
		u, err := NewUpdater(testSrc1, "0.2.2", dir)
		require.Nil(t, err, "create updater failed")

		assert.False(t, u.UpdateAvailable(), "empty updater has update")
		assert.NotNil(t, u.Install(), "empty updater installed")
		assert.Nil(t, u.CheckForUpdate(), "get releases failed")
		assert.Nil(t, u.Install(), "install failed")
		assert.Equal(t, "open", me.name, "wrong command called")
	})
}

func TestHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("HTTP(hello)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := fmt.Fprintln(w, "hello"); err != nil {
				panic(err)
			}
		}))
		defer ts.Close()

		data, err := getURL(ts.URL)
		require.Nil(t, err, "getURL failed")
		ts.Close()

		assert.Equal(t, "hello\n", string(data), "unexpected response")
	})

	t.Run("HTTP(404)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		defer ts.Close()

		_, err := getURL(ts.URL)
		assert.NotNil(t, err, "404 request succeeded")
		ts.Close()
	})

	t.Run("HTTP(fail)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}))
		URL := ts.URL
		ts.Close()

		_, err := getURL(URL)
		assert.NotNil(t, err, "bad request succeeded")
		ts.Close()
	})

	t.Run("HTTP(download)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := fmt.Fprintln(w, "contents"); err != nil {
				panic(err)
			}
		}))
		defer ts.Close()

		f, err := ioutil.TempFile("", "awgo-*-test")
		require.Nil(t, err, "create tempfile failed")
		defer panicOnError(f.Close())

		err = download(ts.URL, f.Name())
		require.Nil(t, err, "download failed")

		data, err := ioutil.ReadFile(f.Name())
		require.Nil(t, err, "read file failed")
		assert.Equal(t, "contents\n", string(data), "unexpected file contents")
	})

	t.Run("HTTP(download fails)", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := fmt.Fprintln(w, "contents"); err != nil {
				panic(err)
			}
		}))
		URL := ts.URL
		ts.Close()

		err := download(URL, "")
		require.NotNil(t, err, "bad download succeeded")
	})
}

func TestRunCommand(t *testing.T) {
	assert.Nil(t, runCommand("/usr/bin/true"), `exec "/usr/bin/true" returned error`)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
