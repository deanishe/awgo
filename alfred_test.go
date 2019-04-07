// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"
)

func TestAlfred(t *testing.T) {

	var (
		a   = NewAlfred()
		x   string
		err error
	)

	defer os.Setenv("alfred_version", os.Getenv("alfred_version"))
	os.Setenv("alfred_version", "")
	a.noRunScripts = true

	x = `Application("com.runningwithcrayons.Alfred").search("");`
	if err = a.Search(""); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("com.runningwithcrayons.Alfred").search("awgo alfred");`
	if err := a.Search("awgo alfred"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("com.runningwithcrayons.Alfred").action(["/","/Volumes"]);`
	if err := a.Action("/", "/Volumes"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Action. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("com.runningwithcrayons.Alfred").browse("/Users");`
	if err := a.Browse("/Users"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("com.runningwithcrayons.Alfred").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	if err := a.RunTrigger("test", "AwGo, yo!"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Trigger. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("com.runningwithcrayons.Alfred").setTheme("Alfred Notepad");`
	if err := a.SetTheme("Alfred Notepad"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Theme. Expected=%q, Got=%q", x, a.lastScript)
	}

	os.Setenv("alfred_version", "3.8.1")
	x = `Application("Alfred 3").search("");`
	if err = a.Search(""); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("Alfred 3").search("awgo alfred");`
	if err := a.Search("awgo alfred"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("Alfred 3").action(["/","/Volumes"]);`
	if err := a.Action("/", "/Volumes"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Action. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("Alfred 3").browse("/Users");`
	if err := a.Browse("/Users"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Search. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("Alfred 3").runTrigger("test", {"inWorkflow":"net.deanishe.awgo","withArgument":"AwGo, yo!"});`
	if err := a.RunTrigger("test", "AwGo, yo!"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Trigger. Expected=%q, Got=%q", x, a.lastScript)
	}

	x = `Application("Alfred 3").setTheme("Alfred Notepad");`
	if err := a.SetTheme("Alfred Notepad"); err != nil {
		t.Error(err)
	}
	if a.lastScript != x {
		t.Errorf("Bad Theme. Expected=%q, Got=%q", x, a.lastScript)
	}

}
