// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"
)

func TestReset(t *testing.T) {
	t.Parallel()

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
