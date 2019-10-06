// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockExec struct {
	name string
	args []string
}

func (me *mockExec) Run(name string, arg ...string) error {
	me.name = name
	me.args = append([]string{name}, arg...)
	return nil
}

func TestReset(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		s := wf.Dir()
		x, err := os.Getwd()
		if err != nil {
			t.Fatalf("[ERROR] %v", err)
		}
		if s != x {
			t.Errorf("Bad Dir. Expected=%v, Got=%v", x, s)
		}

		name := "xyz.json"
		data := []byte("muh bytes")
		if err := wf.Cache.Store(name, data); err != nil {
			t.Fatal(err)
		}
		if err := wf.Data.Store(name, data); err != nil {
			t.Fatal(err)
		}
		if err := wf.Session.Store(name, data); err != nil {
			t.Fatal(err)
		}

		if !wf.Cache.Exists(name) {
			t.Fatal("Cache does not exist")
		}
		if !wf.Data.Exists(name) {
			t.Fatal("Data do not exist")
		}
		if !wf.Session.Exists(name) {
			t.Fatal("Session cache does not exist")
		}

		if err := wf.Reset(); err != nil {
			t.Fatal(err)
		}

		if wf.Cache.Exists(name) {
			t.Fatal("Cache exists")
		}
		if wf.Data.Exists(name) {
			t.Fatal("Data exist")
		}
		if wf.Session.Exists(name) {
			t.Fatal("Session cache exists")
		}
	})
}

func TestWorkflowRoot(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		p := findWorkflowRoot(wd)
		if p != wd {
			t.Errorf("Bad Workflow directory. Expected=%q, Got=%q", wd, p)
		}
	})
}

func TestOpen(t *testing.T) {
	helpURL := "https://github.com/deanishe/awgo"

	withTestWf(func(wf *Workflow) {
		wf.Configure(HelpURL(helpURL))
		tests := []struct {
			fn   func() error
			name string
			args []string
		}{
			{wf.OpenLog, "open",
				[]string{"open", wf.LogFile()},
			},
			{wf.OpenHelp, "open",
				[]string{"open", helpURL},
			},
			{wf.OpenCache, "open",
				[]string{"open", wf.CacheDir()},
			},
			{wf.OpenData, "open",
				[]string{"open", wf.DataDir()},
			},
		}

		for _, td := range tests {
			me := &mockExec{}
			wf.execFunc = me.Run
			td.fn()
			assert.Equal(t, td.name, me.name, "Wrong command name")
			assert.Equal(t, td.args, me.args, "Wrong command args")
		}
	})
}
