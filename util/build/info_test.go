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
	rootDirV3     = "./testdata/v3"
	rootDirV4     = "./testdata/v4"
	syncDirV3     = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3")
	prefsBundleV3 = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Alfred.alfredpreferences")
	wfDirV3       = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Alfred.alfredpreferences/workflows")
	cacheDirV3    = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data")
	dataDirV3     = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Workflow Data")
	syncDirV4     = os.ExpandEnv("${HOME}/Library/Application Support/Alfred")
	prefsBundleV4 = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Alfred.alfredpreferences")
	wfDirV4       = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Alfred.alfredpreferences/workflows")
	cacheDirV4    = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred/Workflow Data")
	dataDirV4     = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Workflow Data")

	testPlist = InfoPlist("./testdata/info.plist")
)

func withEnv(env map[string]string, fn func()) {
	prev := map[string]string{}
	prevSet := map[string]bool{}
	for key, value := range env {
		prev[key], prevSet[key] = os.LookupEnv(key)
		os.Setenv(key, value)
	}

	fn()

	for key, value := range prev {
		if prevSet[key] {
			os.Setenv(key, value)
		} else {
			os.Unsetenv(key)
		}
	}
}

func TestWorkflowInfo(t *testing.T) {
	var (
		name     = "AwGo"
		bundleID = "net.deanishe.awgo"
		version  = "1.2.0"
	)
	info, err := NewInfo(LibDir(rootDirV4), testPlist)
	require.Nil(t, err, "NewInfo failed")
	assert.Equal(t, name, info.Name, "unexpected name")
	assert.Equal(t, bundleID, info.BundleID, "unexpected bundle ID")
	assert.Equal(t, version, info.Version, "unexpected version")

	// Read workflow data from info.plist
	env := map[string]string{
		"alfred_workflow_bundleid": "",
		"alfred_workflow_name":     "",
		"alfred_workflow_version":  "",
	}
	withEnv(env, func() {
		info, err := NewInfo(LibDir(rootDirV4), testPlist)
		require.Nil(t, err, "NewInfo failed")
		assert.Equal(t, name, info.Name, "unexpected name")
		assert.Equal(t, bundleID, info.BundleID, "unexpected bundle ID")
		assert.Equal(t, version, info.Version, "unexpected version")
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
		td := td
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
		x       int
	}{
		{"default", "", rootDirV4, 4},
		{"v4", "4", rootDirV4, 4},
		{"v3", "3", rootDirV3, 3},
		{"v3 (v4 dir)", "3", rootDirV4, 3},
	}

	for _, td := range tests {
		td := td
		t.Run(td.name, func(t *testing.T) {
			withEnv(map[string]string{
				"alfred_version": td.version,
				// ensure defaults
				"alfred_workflow_data":  "",
				"alfred_workflow_cache": "",
			}, func() {
				info, err := NewInfo(LibDir(td.dir), testPlist)
				require.Nil(t, err, "NewInfo failed")

				var (
					syncX  = syncDirV4
					prefsX = prefsBundleV4
					wfDirX = wfDirV4
					cacheX = cacheDirV4
					dataX  = dataDirV4
				)
				if td.x == 3 {
					syncX = syncDirV3
					prefsX = prefsBundleV3
					wfDirX = wfDirV3
					cacheX = cacheDirV3
					dataX = dataDirV3
				}

				assert.Equal(t, syncX, info.AlfredSyncDir, "unexpected AlfredSyncDir")
				assert.Equal(t, prefsX, info.AlfredPrefsBundle, "unexpected PrefsBundle")
				assert.Equal(t, wfDirX, info.AlfredWorkflowDir, "unexpected AlfredWorkflowDir")
				assert.Equal(t, cacheX, info.AlfredCacheDir, "unexpected AlfredCacheDir")
				assert.Equal(t, dataX, info.AlfredDataDir, "unexpected AlfredDataDir")
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
		td := td
		t.Run(td.key, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, env[td.key], "unexpected value")
		})
	}
}
