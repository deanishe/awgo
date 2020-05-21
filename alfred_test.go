// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlfred(t *testing.T) {
	a := NewAlfred()
	defer func() {
		if err := os.Setenv("alfred_version", ""); err != nil {
			panic(fmt.Sprintf("setenv failed: %v", err))
		}
	}()

	t.Run("setup", func(t *testing.T) {
		require.Nil(t, os.Setenv("alfred_version", ""), "setenv failed")
		a.noRunScripts = true
	})

	t.Run("empty action", func(t *testing.T) {
		// run this first because a.lastScript won't be set
		assert.Nil(t, a.Action(), "call empty action failed")
		assert.Equal(t, "", a.lastScript, "run empty action failed")
	})

	t.Run("empty search", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").search("");`
		assert.Nil(t, a.Search(""), "call empty search failed")
		assert.Equal(t, x, a.lastScript, "empty search failed")
	})

	t.Run("search", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").search("awgo alfred");`
		assert.Nil(t, a.Search("awgo alfred"), "call search failed")
		assert.Equal(t, x, a.lastScript, "search failed")
	})

	t.Run("action", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").action(["/","/Volumes"]);`
		assert.Nil(t, a.Action("/", "/Volumes"), "call action failed")
		assert.Equal(t, x, a.lastScript, "action failed")
	})

	t.Run("browse", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").browse("/Users");`
		assert.Nil(t, a.Browse("/Users"), "call browse failed")
		assert.Equal(t, x, a.lastScript, "browse failed")
	})

	t.Run("run trigger", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
		assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "call trigger failed")
		assert.Equal(t, x, a.lastScript, "run trigger failed")
	})
	t.Run("run 3rd-party trigger", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"com.example.workflow","withArgument":"AwGo, yo!"});`
		assert.Nil(t, a.RunTrigger("test", "AwGo, yo!", "com.example.workflow"), "call 3rd-party trigger failed")
		assert.Equal(t, x, a.lastScript, "run trigger in other workflow failed")
	})

	t.Run("set theme", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").setTheme("Alfred Notepad");`
		assert.Nil(t, a.SetTheme("Alfred Notepad"), "call set theme failed")
		assert.Equal(t, x, a.lastScript, "set theme failed")
	})

	t.Run("reload workflow", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").reloadWorkflow("net.deanishe.awgo");`
		assert.Nil(t, a.ReloadWorkflow(), "call reload workflow failed")
		assert.Equal(t, x, a.lastScript, "reload workflow failed")
	})

	t.Run("reload 3rd-party workflow", func(t *testing.T) {
		x := `Application("com.runningwithcrayons.Alfred").reloadWorkflow("com.example.workflow");`
		assert.Nil(t, a.ReloadWorkflow("com.example.workflow"), "call reload workflow failed")
		assert.Equal(t, x, a.lastScript, "reload workflow failed")
	})

	// run a do-nothing script
	t.Run("do-nothing script", func(t *testing.T) {
		a.noRunScripts = false
		js := `function run(argv) { return %s }`
		assert.Nil(t, a.runScript(js), "run script failed")
	})
}

func TestAlfred3(t *testing.T) {
	a := NewAlfred()

	defer func() {
		if err := os.Setenv("alfred_version", ""); err != nil {
			panic(fmt.Sprintf("setenv failed: %v", err))
		}
	}()
	a.noRunScripts = true

	t.Run("set env", func(t *testing.T) {
		require.Nil(t, os.Setenv("alfred_version", "3.8.1"), "setenv failed")
	})

	t.Run("empty search", func(t *testing.T) {
		x := `Application("Alfred 3").search("");`
		assert.Nil(t, a.Search(""), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "search failed")
	})

	t.Run("search", func(t *testing.T) {
		x := `Application("Alfred 3").search("awgo alfred");`
		assert.Nil(t, a.Search("awgo alfred"), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "search failed")
	})

	t.Run("action", func(t *testing.T) {
		x := `Application("Alfred 3").action(["/","/Volumes"]);`
		assert.Nil(t, a.Action("/", "/Volumes"), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "action failed")
	})

	t.Run("browser", func(t *testing.T) {
		x := `Application("Alfred 3").browse("/Users");`
		assert.Nil(t, a.Browse("/Users"), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "browse failed")
	})

	t.Run("run trigger", func(t *testing.T) {
		x := `Application("Alfred 3").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
		assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "run trigger failed")
	})

	t.Run("set theme", func(t *testing.T) {
		x := `Application("Alfred 3").setTheme("Alfred Notepad");`
		assert.Nil(t, a.SetTheme("Alfred Notepad"), "Alfred call failed")
		assert.Equal(t, x, a.lastScript, "run trigger failed")
	})
}
