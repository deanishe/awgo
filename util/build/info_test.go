// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package build

import (
	"fmt"
	"os"
	"testing"
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
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != name {
		t.Errorf("Bad Name. Expected=%q, Got=%q", name, info.Name)
	}
	if info.BundleID != bundleID {
		t.Errorf("Bad BundleID. Expected=%q, Got=%q", bundleID, info.BundleID)
	}
	if info.Version != version {
		t.Errorf("Bad Version. Expected=%q, Got=%q", version, info.Version)
	}

	// Read workflow data from info.plist
	env := map[string]string{
		"alfred_workflow_bundleid": "",
		"alfred_workflow_name":     "",
		"alfred_workflow_version":  "",
	}
	withEnv(env, func() {
		info, err := NewInfo(LibDir(rootDirV4), testPlist)
		if err != nil {
			t.Fatal(err)
		}
		if info.Name != name {
			t.Errorf("Bad Name. Expected=%q, Got=%q", name, info.Name)
		}
		if info.BundleID != bundleID {
			t.Errorf("Bad BundleID. Expected=%q, Got=%q", bundleID, info.BundleID)
		}
		if info.Version != version {
			t.Errorf("Bad Version. Expected=%q, Got=%q", version, info.Version)
		}
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
					if err == nil {
						t.Error("Error expected")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				// info := &Info{dir: td.dir}
				if info.AlfredMajorVersion != td.x {
					t.Errorf("Bad Version. Expected=%d, Got=%d", td.x, info.AlfredMajorVersion)
				}
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
				if err != nil {
					t.Fatal(err)
				}
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

				if info.AlfredSyncDir != syncX {
					t.Errorf("Bad SyncDir. Expected=%q, Got=%q", syncX, info.AlfredSyncDir)
				}
				if info.AlfredPrefsBundle != prefsX {
					t.Errorf("Bad PrefsBundle. Expected=%q, Got=%q", prefsX, info.AlfredPrefsBundle)
				}
				if info.AlfredWorkflowDir != wfDirX {
					t.Errorf("Bad WorkflowsDir. Expected=%q, Got=%q", wfDirX, info.AlfredWorkflowDir)
				}
				if info.AlfredCacheDir != cacheX {
					t.Errorf("Bad AlfredCacheDir. Expected=%q, Got=%q", cacheX, info.AlfredCacheDir)
				}
				if info.AlfredDataDir != dataX {
					t.Errorf("Bad AlfredDataDir. Expected=%q, Got=%q", dataX, info.AlfredDataDir)
				}
			})
		})
	}
}

func TestEnv(t *testing.T) {
	t.Parallel()

	info, err := NewInfo(LibDir(rootDirV4), testPlist)
	if err != nil {
		t.Fatal(err)
	}
	env := info.Env()
	if env["alfred_workflow_name"] != info.Name {
		t.Errorf("Bad Name. Expected=%q, Got=%q", info.Name, env["alfred_workflow_name"])
	}
	if env["alfred_workflow_version"] != info.Version {
		t.Errorf("Bad Version. Expected=%q, Got=%q", info.Version, env["alfred_workflow_version"])
	}
	if env["alfred_workflow_bundleid"] != info.BundleID {
		t.Errorf("Bad BundleID. Expected=%q, Got=%q", info.BundleID, env["alfred_workflow_bundleid"])
	}
	if env["alfred_workflow_uid"] != info.BundleID {
		t.Errorf("Bad UID. Expected=%q, Got=%q", info.BundleID, env["alfred_workflow_uid"])
	}
	if env["alfred_workflow_cache"] != info.CacheDir {
		t.Errorf("Bad CacheDir. Expected=%q, Got=%q", info.CacheDir, env["alfred_workflow_cache"])
	}
	if env["alfred_workflow_data"] != info.DataDir {
		t.Errorf("Bad DataDir. Expected=%q, Got=%q", info.DataDir, env["alfred_workflow_data"])
	}
	if env["alfred_preferences"] != info.AlfredPrefsBundle {
		t.Errorf("Bad PrefsBundle. Expected=%q, Got=%q", info.AlfredPrefsBundle, env["alfred_preferences"])
	}
	if env["alfred_version"] != fmt.Sprintf("%d", info.AlfredMajorVersion) {
		t.Errorf("Bad Version. Expected=%q, Got=%q", fmt.Sprintf("%d", info.AlfredMajorVersion), env["alfred_version"])
	}
}
