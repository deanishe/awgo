// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlfred(t *testing.T) {
	t.Parallel()

	var (
		a = NewAlfred()
		x string
	)

	defer panicOnErr(os.Setenv("alfred_version", os.Getenv("alfred_version")))
	panicOnErr(os.Setenv("alfred_version", ""))
	a.noRunScripts = true

	x = `Application("com.runningwithcrayons.Alfred").search("");`
	assert.Nil(t, a.Search(""), "Search failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("com.runningwithcrayons.Alfred").search("awgo alfred");`
	assert.Nil(t, a.Search("awgo alfred"), "Search failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("com.runningwithcrayons.Alfred").action(["/","/Volumes"]);`
	assert.Nil(t, a.Action("/", "/Volumes"), "Action failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("com.runningwithcrayons.Alfred").browse("/Users");`
	assert.Nil(t, a.Browse("/Users"), "Browse failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "RunTrigger failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("com.runningwithcrayons.Alfred").setTheme("Alfred Notepad");`
	assert.Nil(t, a.SetTheme("Alfred Notepad"), "SetTheme failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	panicOnErr(os.Setenv("alfred_version", "3.8.1"))
	x = `Application("Alfred 3").search("");`
	assert.Nil(t, a.Search(""), "Search failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("Alfred 3").search("awgo alfred");`
	assert.Nil(t, a.Search("awgo alfred"), "Search failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("Alfred 3").action(["/","/Volumes"]);`
	assert.Nil(t, a.Action("/", "/Volumes"), "Action failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("Alfred 3").browse("/Users");`
	assert.Nil(t, a.Browse("/Users"), "Browse failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("Alfred 3").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	assert.Nil(t, a.RunTrigger("test", "AwGo, yo!"), "RunTrigger failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")

	x = `Application("Alfred 3").setTheme("Alfred Notepad");`
	assert.Nil(t, a.SetTheme("Alfred Notepad"), "SetTheme failed")
	assert.Equal(t, x, a.lastScript, "unexpected script")
}
