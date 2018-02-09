//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-13
//

package aw

import (
	"os"
	"testing"
)

// TestContext verifies that Context holds the expected values.
func TestContext(t *testing.T) {
	var (
		version                  = "0.13"
		name                     = "AwGo"
		bundleID                 = "net.deanishe.awgo"
		uid                      = "user.workflow.4B0E9731-E139-4179-BC50-D7FFF82B269A"
		debug                    = true
		alfredVersion            = "3.5.1"
		alfredBuild              = "883"
		theme                    = "alfred.theme.custom.DE3D17CA-64A2-4B42-A3F6-C71DB1201F88"
		themeBackground          = "rgba(255,255,255,1.00)"
		themeSelectionBackground = "rgba(255,255,255,1.00)"
		preferences              = os.ExpandEnv("$HOME/Dropbox/Config/Alfred 3/Alfred.alfredpreferences")
		localhash                = "0dd4500828b5c675862eaa7786cf0f374823b965"
		cacheDir                 = os.ExpandEnv("$HOME/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data/net.deanishe.awgo")
		dataDir                  = os.ExpandEnv("$HOME/Library/Application Support/Alfred 3/Workflow Data/net.deanishe.awgo")
	)

	ctx := NewContext()

	if ctx.WorkflowVersion != version {
		t.Errorf("bad version. Expected=%s, Got=%s", version, ctx.WorkflowVersion)
	}
	if ctx.Name != name {
		t.Errorf("bad name. Expected=%s, Got=%s", name, ctx.Name)
	}
	if ctx.BundleID != bundleID {
		t.Errorf("bad bundle ID. Expected=%s, Got=%s", bundleID, ctx.BundleID)
	}
	if ctx.UID != uid {
		t.Errorf("bad UID. Expected=%s, Got=%s", uid, ctx.UID)
	}
	if ctx.Debug != debug {
		t.Errorf("bad debug. Expected=%v, Got=%v", debug, ctx.Debug)
	}
	if ctx.AlfredVersion != alfredVersion {
		t.Errorf("bad Alfred version. Expected=%s, Got=%s", alfredVersion, ctx.AlfredVersion)
	}
	if ctx.AlfredBuild != alfredBuild {
		t.Errorf("bad Alfred build. Expected=%s, Got=%s", alfredBuild, ctx.AlfredBuild)
	}
	if ctx.Theme != theme {
		t.Errorf("bad theme. Expected=%s, Got=%s", theme, ctx.Theme)
	}
	if ctx.ThemeBackground != themeBackground {
		t.Errorf("bad background. Expected=%s, Got=%s", themeBackground, ctx.ThemeBackground)
	}
	if ctx.ThemeSelectionBackground != themeSelectionBackground {
		t.Errorf("bad selection background. Expected=%s, Got=%s", themeSelectionBackground, ctx.ThemeSelectionBackground)
	}
	if ctx.Preferences != preferences {
		t.Errorf("bad preferences. Expected=%s, Got=%s", preferences, ctx.Preferences)
	}
	if ctx.Localhash != localhash {
		t.Errorf("bad localhash. Expected=%s, Got=%s", localhash, ctx.Localhash)
	}
	if ctx.CacheDir != cacheDir {
		t.Errorf("bad cache dir. Expected=%s, Got=%s", cacheDir, ctx.CacheDir)
	}
	if ctx.DataDir != dataDir {
		t.Errorf("bad data dir. Expected=%s, Got=%s", dataDir, ctx.DataDir)
	}
}
