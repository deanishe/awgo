// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package build

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	rootDirV3      = "./testdata/v3"
	rootDirV4      = "./testdata/v4"
	rootDirInvalid = "./testdata/invalid"
	syncDirV3      = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3")
	prefsBundleV3  = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Alfred.alfredpreferences")
	wfDirV3        = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Alfred.alfredpreferences/workflows")
	cacheDirV3     = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data")
	dataDirV3      = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Workflow Data")
	syncDirV4      = os.ExpandEnv("${HOME}/Library/Application Support/Alfred")
	prefsBundleV4  = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Alfred.alfredpreferences")
	wfDirV4        = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Alfred.alfredpreferences/workflows")
	cacheDirV4     = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred/Workflow Data")
	dataDirV4      = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Workflow Data")

	testPlist = InfoPlist("./testdata/info.plist")
)

func withEnv(env map[string]string, fn func()) {
	prev := map[string]string{}
	prevSet := map[string]bool{}
	for key, value := range env {
		prev[key], prevSet[key] = os.LookupEnv(key)
		panicOnError(os.Setenv(key, value))
	}

	fn()

	for key, value := range prev {
		if prevSet[key] {
			panicOnError(os.Setenv(key, value))
		} else {
			panicOnError(os.Unsetenv(key))
		}
	}
}

func TestWorkflowInfo(t *testing.T) {
	var (
		name     = "AwGo"
		bundleID = "net.deanishe.awgo"
		version  = "1.2.0"
	)

	testInfo := func(t *testing.T) {
		info, err := NewInfo(LibDir(rootDirV4), testPlist)
		require.Nil(t, err, "NewInfo failed")
		assert.Equal(t, name, info.Name, "unexpected name")
		assert.Equal(t, bundleID, info.BundleID, "unexpected bundle ID")
		assert.Equal(t, version, info.Version, "unexpected version")
	}

	testInfo(t)

	t.Run("read info.plist", func(t *testing.T) {
		// Read workflow data from info.plist
		env := map[string]string{
			"alfred_workflow_bundleid": "",
			"alfred_workflow_name":     "",
			"alfred_workflow_version":  "",
		}
		withEnv(env, func() {
			testInfo(t)
		})
	})

	t.Run("info.plist has priority over env", func(t *testing.T) {
		env := map[string]string{
			"alfred_workflow_bundleid": "net.deanishe.wrong-bundleid",
			"alfred_workflow_name":     "Wrong Name",
			"alfred_workflow_version":  "0.0.1",
		}
		withEnv(env, func() {
			testInfo(t)
		})
	})
}

// Read Alfred version number from environment or based on
// presence of configuration files.
func TestAlfredVersion(t *testing.T) {
	tests := []struct {
		dir    string
		envvar string
		x      int
		err    bool
	}{
		{rootDirV3, "", 3, false},
		{rootDirV3, "3", 3, false},
		{rootDirV3, "4", 4, false},
		{rootDirV4, "", 4, false},
		{rootDirV4, "4", 4, false},
		{".", "", 0, true},
		{".", "four", 0, true},
	}

	for _, td := range tests {
		td := td // pin variable
		t.Run(fmt.Sprintf("dir=%q, env=%q", td.dir, td.envvar), func(t *testing.T) {
			withEnv(map[string]string{
				"alfred_version": td.envvar,
				// ensure defaults
				"alfred_workflow_data":  "",
				"alfred_workflow_cache": "",
			}, func() {
				info, err := NewInfo(LibDir(td.dir), testPlist)
				if td.err {
					assert.NotNil(t, err, "expected error")
					return
				}
				require.Nil(t, err, "unexpected error")
				assert.Equal(t, td.x, info.AlfredMajorVersion, "unexpected version")
			})
		})
	}
}

func TestDirs(t *testing.T) {
	tests := []struct {
		name    string
		version string
		dir     string
		plist   Option
		x       int
		fail    bool
	}{
		{"default", "", rootDirV4, testPlist, 4, false},
		{"v4", "4", rootDirV4, testPlist, 4, false},
		{"v4 (version=0)", "", rootDirV4, testPlist, 4, false},
		{"v3", "3", rootDirV3, testPlist, 3, false},
		{"v3 (v4 dir)", "3", rootDirV4, testPlist, 3, false},
		// invalid input
		{"non-existent info.plist", "", rootDirV4, InfoPlist("./invalid"), 0, true},
		{"invalid info.plist", "", rootDirV4, InfoPlist("./testdata/invalid.plist"), 0, true},
		{"non-existent info.plist", "", "./invalid", testPlist, 0, true},
		{"invalid prefs.json", "", rootDirInvalid, testPlist, 0, true},
		{"invalid Alfred Preferences prefs", "3", rootDirInvalid, testPlist, 0, true},
	}

	for _, td := range tests {
		td := td // pin variable
		t.Run(td.name, func(t *testing.T) {
			withEnv(map[string]string{
				"alfred_version": td.version,
				// ensure defaults
				"alfred_workflow_data":  "",
				"alfred_workflow_cache": "",
			}, func() {
				info, err := NewInfo(LibDir(td.dir), td.plist)

				if td.fail {
					assert.NotNil(t, err, td.name)
					return
				}

				require.Nil(t, err, "NewInfo failed")

				if td.x == 3 {
					assert.Equal(t, syncDirV3, info.AlfredSyncDir, "unexpected AlfredSyncDir")
					assert.Equal(t, prefsBundleV3, info.AlfredPrefsBundle, "unexpected PrefsBundle")
					assert.Equal(t, wfDirV3, info.AlfredWorkflowDir, "unexpected AlfredWorkflowDir")
					assert.Equal(t, cacheDirV3, info.AlfredCacheDir, "unexpected AlfredCacheDir")
					assert.Equal(t, dataDirV3, info.AlfredDataDir, "unexpected AlfredDataDir")
				} else {
					assert.Equal(t, syncDirV4, info.AlfredSyncDir, "unexpected AlfredSyncDir")
					assert.Equal(t, prefsBundleV4, info.AlfredPrefsBundle, "unexpected PrefsBundle")
					assert.Equal(t, wfDirV4, info.AlfredWorkflowDir, "unexpected AlfredWorkflowDir")
					assert.Equal(t, cacheDirV4, info.AlfredCacheDir, "unexpected AlfredCacheDir")
					assert.Equal(t, dataDirV4, info.AlfredDataDir, "unexpected AlfredDataDir")
				}
			})
		})
	}
}

func TestEnv(t *testing.T) {
	t.Parallel()

	info, err := NewInfo(LibDir(rootDirV4), testPlist)
	require.Nil(t, err, "NewInfo failed")

	tests := []struct {
		key, x string
	}{
		{"alfred_workflow_name", info.Name},
		{"alfred_workflow_version", info.Version},
		{"alfred_workflow_bundleid", info.BundleID},
		{"alfred_workflow_uid", info.BundleID},
		{"alfred_workflow_cache", info.CacheDir},
		{"alfred_workflow_data", info.DataDir},
		{"alfred_preferences", info.AlfredPrefsBundle},
		{"alfred_version", fmt.Sprintf("%d", info.AlfredMajorVersion)},
	}
	env := info.Env()
	for _, td := range tests {
		td := td // pin variable
		t.Run(td.key, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, env[td.key], "unexpected value")
		})
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
