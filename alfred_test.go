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

	x = `Application("com.runningwithcrayons.Alfred").search("");`
	assert.Nil(t, a.Search(""), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "search failed")

	x = `Application("com.runningwithcrayons.Alfred").search("awgo alfred");`
	assert.Nil(t, a.Search("awgo alfred"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "search failed")

	x = `Application("com.runningwithcrayons.Alfred").action(["/","/Volumes"]);`
	assert.Nil(t, a.Action("/", "/Volumes"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "action failed")

	x = `Application("com.runningwithcrayons.Alfred").browse("/Users");`
	assert.Nil(t, a.Browse("/Users"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "browse failed")

	x = `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "run trigger failed")

	x = `Application("com.runningwithcrayons.Alfred").setTheme("Alfred Notepad");`
	assert.Nil(t, a.SetTheme("Alfred Notepad"), "Alfred call failed")
	assert.Equal(t, x, a.lastScript, "run trigger failed")
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
