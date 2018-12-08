// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	tVersion                  = "0.15"
	tName                     = "AwGo"
	tBundleID                 = "net.deanishe.awgo"
	tUID                      = "user.workflow.4B0E9731-E139-4179-BC50-D7FFF82B269A"
	tDebug                    = true
	tAlfredVersion            = "3.6"
	tAlfredBuild              = "901"
	tTheme                    = "alfred.theme.custom.DE3D17CA-64A2-4B42-A3F6-C71DB1201F88"
	tThemeBackground          = "rgba(255,255,255,1.00)"
	tThemeSelectionBackground = "rgba(255,255,255,1.00)"
	tPreferences              = os.ExpandEnv("$HOME/Library/Application Support/Alfred 3/Alfred.alfredpreferences")
	tLocalhash                = "0dd4500828b5c675862eaa7786cf0f374823b965"
	tCacheDir                 = os.ExpandEnv("$HOME/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data/net.deanishe.awgo")
	tDataDir                  = os.ExpandEnv("$HOME/Library/Application Support/Alfred 3/Workflow Data/net.deanishe.awgo")

	testEnv = MapEnv{
		EnvVarVersion:          tVersion,
		EnvVarName:             tName,
		EnvVarBundleID:         tBundleID,
		EnvVarUID:              tUID,
		EnvVarDebug:            fmt.Sprintf("%v", tDebug),
		EnvVarAlfredVersion:    tAlfredVersion,
		EnvVarAlfredBuild:      tAlfredBuild,
		EnvVarTheme:            tTheme,
		EnvVarThemeBG:          tThemeBackground,
		EnvVarThemeSelectionBG: tThemeSelectionBackground,
		EnvVarPreferences:      tPreferences,
		EnvVarLocalhash:        tLocalhash,
		EnvVarCacheDir:         tCacheDir,
		EnvVarDataDir:          tDataDir,
	}

	tInfoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>bundleid</key>
	<string>net.deanishe.awgo</string>
	<key>connections</key>
	<dict/>
	<key>createdby</key>
	<string>Dean Jackson</string>
	<key>description</key>
	<string>AwGo sample info.plist</string>
	<key>disabled</key>
	<false/>
	<key>name</key>
	<string>AwGo</string>
	<key>objects</key>
	<array/>
	<key>readme</key>
	<string></string>
	<key>uidata</key>
	<dict/>
	<key>webaddress</key>
	<string>https://github.com/deanishe/awgo</string>
    <key>version</key>
    <string>0.15</string>
	<key>variables</key>
	<dict>
		<key>exported_var</key>
		<string>exported_value</string>
		<key>unexported_var</key>
		<string>unexported_value</string>
	</dict>
	<key>variablesdontexport</key>
	<array>
		<string>unexported_var</string>
	</array>
</dict>
</plist>
`
)

// Call function with a test environment.
func withTestEnv(fun func(e MapEnv)) {
	e := MapEnv{
		EnvVarVersion:          tVersion,
		EnvVarName:             tName,
		EnvVarBundleID:         tBundleID,
		EnvVarUID:              tUID,
		EnvVarDebug:            fmt.Sprintf("%v", tDebug),
		EnvVarAlfredVersion:    tAlfredVersion,
		EnvVarAlfredBuild:      tAlfredBuild,
		EnvVarTheme:            tTheme,
		EnvVarThemeBG:          tThemeBackground,
		EnvVarThemeSelectionBG: tThemeSelectionBackground,
		EnvVarPreferences:      tPreferences,
		EnvVarLocalhash:        tLocalhash,
		EnvVarCacheDir:         tCacheDir,
		EnvVarDataDir:          tDataDir,
	}

	fun(e)
}

// Call function in a test workflow environment.
func withTestWf(fun func(wf *Workflow)) {

	withTestEnv(func(e MapEnv) {

		var (
			curdir, dir string
			err         error
		)

		curdir, err = os.Getwd()
		if err != nil {
			panic(err)
		}

		dir, err = ioutil.TempDir("", "awgo-")
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				panic(err)
			}
		}()

		// TempDir() returns a symlink on my macOS :(
		dir, err = filepath.EvalSymlinks(dir)
		if err != nil {
			panic(err)
		}

		var (
			wfdir    = filepath.Join(dir, "workflow")
			datadir  = filepath.Join(dir, "data")
			cachedir = filepath.Join(dir, "cache")
			ipfile   = filepath.Join(wfdir, "info.plist")
		)

		// Update env to point to cache & data dirs
		e[EnvVarCacheDir] = cachedir
		e[EnvVarDataDir] = datadir

		// Create test files & directories
		for _, p := range []string{wfdir, datadir, cachedir} {
			if err := os.MkdirAll(p, os.ModePerm); err != nil {
				panic(err)
			}
		}
		// info.plist
		if err := ioutil.WriteFile(ipfile, []byte(tInfoPlist), os.ModePerm); err != nil {
			panic(err)
		}

		// Change to workflow directory and call function from there.
		if err := os.Chdir(wfdir); err != nil {
			panic(err)
		}

		defer func() {
			if err := os.Chdir(curdir); err != nil {
				panic(err)
			}
		}()

		// Create workflow for current environment and pass it to function.
		var wf = NewFromEnv(e)
		fun(wf)
	})

}

// TestWithTestWf verifies the withTestEnv helper.
func TestWithTestWf(t *testing.T) {

	withTestWf(func(wf *Workflow) {

		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		cd := filepath.Join(wd, "../cache")
		dd := filepath.Join(wd, "../data")

		data := []struct {
			name, x, v string
		}{
			{"Workflow.Dir", wd, wf.Dir()},
			{"Cache.Dir", cd, wf.Cache.Dir},
			{"CacheDir", cd, wf.CacheDir()},
			{"Data.Dir", dd, wf.Data.Dir},
			{"DataDir", dd, wf.DataDir()},
			{"Name", tName, wf.Name()},
			{"BundleID", tBundleID, wf.BundleID()},

			{"Config.UID", tUID, wf.Config.Get(EnvVarUID)},
			{"Config.AlfredVersion", tAlfredVersion, wf.Config.Get(EnvVarAlfredVersion)},
			{"Config.AlfredBuild", tAlfredBuild, wf.Config.Get(EnvVarAlfredBuild)},
			{"Config.Theme", tTheme, wf.Config.Get(EnvVarTheme)},
			{"Config.ThemeBackground", tThemeBackground, wf.Config.Get(EnvVarThemeBG)},
			{"Config.ThemeSelectionBackground", tThemeSelectionBackground,
				wf.Config.Get(EnvVarThemeSelectionBG)},
			{"Config.Preferences", tPreferences, wf.Config.Get(EnvVarPreferences)},
			{"Config.Localhash", tLocalhash, wf.Config.Get(EnvVarLocalhash)},
			{"Config.CacheDir", cd, wf.Config.Get(EnvVarCacheDir)},
			{"Config.DataDir", dd, wf.Config.Get(EnvVarDataDir)},
		}

		if wf.Debug() != tDebug {
			t.Errorf("Bad Debug(). Expected=%v, Got=%v", tDebug, wf.Debug())
		}

		for _, td := range data {
			if td.v != td.x {
				t.Errorf("Bad %s. Expected=%#v, Got=%#v", td.name, td.x, td.v)
			}
		}

	})
}

// slicesEqual tests if 2 string slices are equal.
func slicesEqual(a, b []string) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// tests whether two maps have the same contents.
func verifyMapsEqual(a, b map[string]string) error {

	if a == nil && b == nil {
		return nil
	}

	if a == nil || b == nil {
		return fmt.Errorf("nil map. a=%v, b=%v", a, b)
	}

	if len(a) != len(b) {
		return fmt.Errorf("different lengths (%d != %d)", len(a), len(b))
	}

	for k, v := range a {

		v2, ok := b[k]
		if !ok {
			return fmt.Errorf("key %v missing in b", k)

		} else if v2 != v {
			return fmt.Errorf("%s is different: %#v != %#v", k, v, v2)
		}
	}

	return nil
}
