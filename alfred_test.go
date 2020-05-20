// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlfred(t *testing.T) {
	var (
		a = NewAlfred()
		x string
	)

	defer func() {
		if err := os.Setenv("alfred_version", ""); err != nil {
			panic(fmt.Sprintf("setenv failed: %v", err))
		}
	}()
	assert.Nil(t, os.Setenv("alfred_version", ""), "setenv failed")
	a.noRunScripts = true

	// run this first because a.lastScript won't be set
	assert.Nil(t, a.Action(), "call empty action failed")
	assert.Equal(t, "", a.lastScript, "run empty action failed")

	x = `Application("com.runningwithcrayons.Alfred").search("");`
	assert.Nil(t, a.Search(""), "call empty search failed")
	assert.Equal(t, x, a.lastScript, "empty search failed")

	x = `Application("com.runningwithcrayons.Alfred").search("awgo alfred");`
	assert.Nil(t, a.Search("awgo alfred"), "call search failed")
	assert.Equal(t, x, a.lastScript, "search failed")

	x = `Application("com.runningwithcrayons.Alfred").action(["/","/Volumes"]);`
	assert.Nil(t, a.Action("/", "/Volumes"), "call action failed")
	assert.Equal(t, x, a.lastScript, "action failed")

	x = `Application("com.runningwithcrayons.Alfred").browse("/Users");`
	assert.Nil(t, a.Browse("/Users"), "call browse failed")
	assert.Equal(t, x, a.lastScript, "browse failed")

	x = `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "call trigger failed")
	assert.Equal(t, x, a.lastScript, "run trigger failed")

	x = `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"com.example.workflow","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!", "com.example.workflow"), "call 3rd-party trigger failed")
	assert.Equal(t, x, a.lastScript, "run trigger in other workflow failed")

	x = `Application("com.runningwithcrayons.Alfred").setTheme("Alfred Notepad");`
	assert.Nil(t, a.SetTheme("Alfred Notepad"), "call set theme failed")
	assert.Equal(t, x, a.lastScript, "set theme failed")

	// run a do-nothing script
	t.Run("do-nothing script", func(t *testing.T) {
		a.noRunScripts = false
		js := `function run(argv) { return %s }`
		assert.Nil(t, a.runScript(js), "run script failed")
	})
}

func TestAlfred3(t *testing.T) {
	var (
		a = NewAlfred()
		x string
	)

	defer func() {
		if err := os.Setenv("alfred_version", ""); err != nil {
			panic(fmt.Sprintf("setenv failed: %v", err))
		}
	}()
	a.noRunScripts = true

	assert.Nil(t, os.Setenv("alfred_version", "3.8.1"), "setenv failed")
	x = `Application("Alfred 3").search("");`
	assert.Nil(t, a.Search(""), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "search failed")

	x = `Application("Alfred 3").search("awgo alfred");`
	assert.Nil(t, a.Search("awgo alfred"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "search failed")

	x = `Application("Alfred 3").action(["/","/Volumes"]);`
	assert.Nil(t, a.Action("/", "/Volumes"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "action failed")

	x = `Application("Alfred 3").browse("/Users");`
	assert.Nil(t, a.Browse("/Users"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "browse failed")

	x = `Application("Alfred 3").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "run trigger failed")

	x = `Application("Alfred 3").setTheme("Alfred Notepad");`
	assert.Nil(t, a.SetTheme("Alfred Notepad"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "run trigger failed")
}
