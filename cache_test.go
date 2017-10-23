//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-08
//

package aw

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deanishe/awgo/util"
)

// WithTempDir creates a temporary directory, calls function fn, then deletes the directory.
func WithTempDir(fn func(dir string)) {
	root := os.TempDir()
	p := filepath.Join(root, fmt.Sprintf("awgo-%d.%d", os.Getpid(), time.Now().Nanosecond()))
	util.MustExist(p)
	defer os.RemoveAll(p)
	fn(p)
}

// TestStoreAndLoad checks that data are stored and loaded correctly
func TestStoreAndLoad(t *testing.T) {
	WithTempDir(func(dir string) {
		c := NewCache(dir)
		s := "this is a test"
		n := "test.txt"

		// Sanity checks
		p := c.path(n)
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Delete non-existant store
		if err := c.Store(n, []byte{}); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		// Non-existant cache exists
		if c.Exists(n) {
			t.Errorf("non-existant cache exists")
		}

		// Non-existant cache has expired
		if !c.Expired(n, 0) {
			t.Errorf("non-existant cache hasn't expired")
		}

		// Store data
		data := []byte(s)
		if err := c.Store(n, data); err != nil {
			t.Errorf("couldn't cache data to %s: %v", n, err)
		}
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		if c.Exists(n) != util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		// Load data
		data2, err := c.Load(n)
		if err != nil {
			t.Errorf("couldn't load cached data: %v", err)
		}
		if bytes.Compare(data, data2) != 0 {
			t.Errorf("loaded data does not match saved data: expected=%v, got=%v", data, data2)
		}

		// Data age
		age, err := c.Age(n)
		if err != nil {
			t.Errorf("couldn't get age of cache %s: %v", n, err)
		}
		if age == 0 {
			t.Errorf("age is zero")
		}

		// Delete data
		if err := c.Store(n, []byte{}); err != nil {
			t.Errorf("couldn't delete cache %s: %v", p, err)
		}

		age, err = c.Age(n)
		if err == nil {
			t.Errorf("no error getting age of non-existant cache %s: %v", n, err)
		}
		if !os.IsNotExist(err) {
			t.Errorf("deleted cache exists %s: %v", n, err)
		}

		// Load non-existant cache
		if _, err := c.Load(n); err == nil {
			t.Errorf("no error loading non-existant cache")
		}
	})
}

// TestLoadOrStore tests LoadOrStore API.
func TestLoadOrStore(t *testing.T) {
	s := "this is a test"
	var reloadCalled bool
	reload := func() ([]byte, error) {
		reloadCalled = true
		return []byte(s), nil
	}

	WithTempDir(func(dir string) {
		c := NewCache(dir)
		n := "test.txt"
		maxAge := time.Duration(time.Second * 1)

		// Sanity checks
		p := c.path(n)
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Cache empty
		data, err := c.LoadOrStore(n, maxAge, reload)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if bytes.Compare(data, []byte(s)) != 0 {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", []byte(s), data)
		}
		if !reloadCalled {
			t.Errorf("reload wasn't called")
		}

		if c.Expired(n, maxAge) {
			t.Errorf("cache expired")
		}

		// Load cached data
		reloadCalled = false
		data, err = c.LoadOrStore(n, maxAge, reload)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if bytes.Compare(data, []byte(s)) != 0 {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", []byte(s), data)
		}
		if reloadCalled {
			t.Errorf("reload was called")
		}

		// Load with 0 maxAge
		reloadCalled = false
		data, err = c.LoadOrStore(n, 0, reload)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if bytes.Compare(data, []byte(s)) != 0 {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", []byte(s), data)
		}
		if reloadCalled {
			t.Errorf("reload was called")
		}

		time.Sleep(time.Duration(time.Second * 1))

		if !c.Expired(n, maxAge) {
			t.Errorf("cache hasn't expired")
		}

		// Reload data
		reloadCalled = false
		data, err = c.LoadOrStore(n, maxAge, reload)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if bytes.Compare(data, []byte(s)) != 0 {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", []byte(s), data)
		}
		if !reloadCalled {
			t.Errorf("reload wasn't called")
		}
	})
}

// TestData is for testing JSON serialisation.
type TestData struct {
	A string
	B string
}

func (td *TestData) Eq(other *TestData) bool {
	if td.A != other.A {
		return false
	}
	if td.B != other.B {
		return false
	}
	return true
}

// TestStoreJSON round-trips data through the JSON caching API.
func TestStoreJSON(t *testing.T) {
	WithTempDir(func(dir string) {
		n := "test.json"
		c := NewCache(dir)
		p := c.path(n)

		// Delete non-existant store
		if err := c.StoreJSON(n, nil); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		a := &TestData{"one", "two"}
		if err := c.StoreJSON(n, a); err != nil {
			t.Errorf("couldn't store JSON: %v", err)
		}

		if !util.PathExists(p) {
			t.Errorf("cache doesn't exist")
		}

		b := &TestData{}
		if err := c.LoadJSON(n, b); err != nil {
			t.Errorf("couldn't load cached JSON: %v", err)
		}

		if !b.Eq(a) {
			t.Errorf("unexpected data. Expected=%+v, Got=%+v", a, b)
		}

		// Delete store
		if err := c.StoreJSON(n, nil); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		if util.PathExists(p) {
			t.Errorf("couldn't delete cache %s", p)
		}

		// Try to load non-existant cache
		b = &TestData{}
		if err := c.LoadJSON(n, b); err == nil {
			t.Errorf("no error loading non-existant cache")
		}
	})
}

// TestLoadOrStoreJSON tests JSON serialisation.
func TestLoadOrStoreJSON(t *testing.T) {
	var reloadCalled bool
	var a, b *TestData

	reload := func() (interface{}, error) {
		reloadCalled = true
		return &TestData{"one", "two"}, nil
	}

	WithTempDir(func(dir string) {
		n := "test.json"
		c := NewCache(dir)
		maxAge := time.Duration(time.Second * 1)

		// Sanity checks
		p := c.path(n)
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		a = &TestData{"one", "two"}
		b = &TestData{}
		// Cache empty
		err := c.LoadOrStoreJSON(n, maxAge, reload, b)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if !a.Eq(b) {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", a, b)
		}
		if !reloadCalled {
			t.Errorf("reload wasn't called")
		}

		if c.Expired(n, maxAge) {
			t.Errorf("cache expired")
		}

		// Load cached data
		reloadCalled = false
		a = &TestData{"one", "two"}
		b = &TestData{}
		err = c.LoadOrStoreJSON(n, maxAge, reload, b)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if !b.Eq(a) {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", a, b)
		}
		if reloadCalled {
			t.Errorf("reload was called")
		}

		// Load with 0 maxAge
		reloadCalled = false
		a = &TestData{"one", "two"}
		b = &TestData{}
		err = c.LoadOrStoreJSON(n, 0, reload, b)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if !b.Eq(a) {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", a, b)
		}
		if reloadCalled {
			t.Errorf("reload was called")
		}

		time.Sleep(time.Duration(time.Second * 1))

		if !c.Expired(n, maxAge) {
			t.Errorf("cache hasn't expired")
		}

		// Reload data
		reloadCalled = false
		a = &TestData{"one", "two"}
		b = &TestData{}
		err = c.LoadOrStoreJSON(n, maxAge, reload, b)
		if err != nil {
			t.Errorf("couldn't load/store cached data: %v", err)
		}
		if !b.Eq(a) {
			t.Errorf("unexpected cache data. Expected=%v, Got=%v", a, b)
		}
		if !reloadCalled {
			t.Errorf("reload wasn't called")
		}
	})
}

// TestBadReloadError checks reload funcs that return errors
func TestBadReloadError(t *testing.T) {
	reloadB := func() ([]byte, error) {
		return nil, fmt.Errorf("an error")
	}

	reloadJSON := func() (interface{}, error) {
		return nil, fmt.Errorf("an error")
	}

	WithTempDir(func(dir string) {
		c := NewCache(dir)
		n := "test"
		if _, err := c.LoadOrStore(n, 0, reloadB); err == nil {
			t.Error("no error returned by reloadB")
		}
		v := &TestData{}
		if err := c.LoadOrStoreJSON(n, 0, reloadJSON, v); err == nil {
			t.Error("no error returned by reloadJSON")
		}
	})
}

// TestSession tests session-scoped caching.
func TestSession(t *testing.T) {
	WithTempDir(func(dir string) {
		sid := NewSessionID()
		s := NewSession(dir, sid)
		data := []byte("this is a test")
		n := "test.txt"

		// Sanity checks
		p := s.cache.path(s.name(n))
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Delete non-existant store
		if err := s.Store(n, []byte{}); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		// Non-existant cache exists
		if s.Exists(n) {
			t.Errorf("non-existant cache exists")
		}

		// Store data
		if err := s.Store(n, data); err != nil {
			t.Errorf("couldn't cache data to %s: %v", n, err)
		}
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		if s.Exists(n) != util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		// Load data
		data2, err := s.Load(n)
		if err != nil {
			t.Errorf("couldn't load cached data: %v", err)
		}
		if bytes.Compare(data, data2) != 0 {
			t.Errorf("loaded data does not match saved data: expected=%v, got=%v", data, data2)
		}

		// Clear session
		s.Clear(false) // Leave current session data
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}
		// Clear this session's data, too
		s.Clear(true)
		if util.PathExists(p) {
			t.Errorf("cache file exists: %s", p)
		}

		// Load non-existant cache
		if _, err := s.Load(n); err == nil {
			t.Errorf("no error loading non-existant cache")
		}

		// Clear old sessions
		sid1 := NewSessionID()
		sid2 := NewSessionID()
		s = NewSession(dir, sid1)
		s.Store(n, data)

		if !s.Exists(n) {
			t.Errorf("cached data do not exist: %s", n)
		}

		s = NewSession(dir, sid2)
		s.Clear(false)

		if s.Exists(n) {
			t.Errorf("expired data still exist: %s", n)
		}
	})
}
