// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReset(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		s := wf.Dir()
		x, err := os.Getwd()
		require.Nil(t, err, "Getwd failed")
		assert.Equal(t, x, s, "unexpected dir")

		name := "xyz.json"
		data := []byte("muh bytes")
		err = wf.Cache.Store(name, data)
		assert.Nil(t, err, "cache store failed")
		err = wf.Data.Store(name, data)
		assert.Nil(t, err, "data store failed")
		err = wf.Session.Store(name, data)
		assert.Nil(t, err, "session store failed")

		assert.True(t, wf.Cache.Exists(name), "cache data do not exist")
		assert.True(t, wf.Data.Exists(name), "data do not exist")
		assert.True(t, wf.Session.Exists(name), "session data do not exist")

		require.Nil(t, wf.Reset(), "reset failed")

		assert.False(t, wf.Cache.Exists(name), "cache data exist")
		assert.False(t, wf.Data.Exists(name), "data exist")
		assert.False(t, wf.Session.Exists(name), "session data exist")
	})
}

func TestWorkflowRoot(t *testing.T) {
	withTestWf(func(wf *Workflow) {
		wd, err := os.Getwd()
		require.Nil(t, err, "Getwd failed")

		p := findWorkflowRoot(wd)
		assert.Equal(t, wd, p, "unexpected workflow directory")
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
			assert.Nil(t, td.fn(), "test command failed")
			assert.Equal(t, td.name, me.name, "Wrong command name")
			assert.Equal(t, td.args, me.args, "Wrong command args")
		}
	})
}
