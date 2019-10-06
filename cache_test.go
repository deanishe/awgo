// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/deanishe/awgo/util"
)

// Data are stored and loaded correctly
func TestCache_Store(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {
		c := NewCache(dir)
		s := "this is a test"
		n := "test.txt"

		// Sanity checks
		p := c.path(n)
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Delete non-existent store
		if err := c.Store(n, nil); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		// Non-existent cache exists
		if c.Exists(n) {
			t.Errorf("non-existent cache exists")
		}

		// Non-existent cache has expired
		if !c.Expired(n, 0) {
			t.Errorf("non-existent cache hasn't expired")
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
		if err := c.Store(n, nil); err != nil {
			t.Errorf("couldn't delete cache %s: %v", p, err)
		}

		_, err = c.Age(n)
		if err == nil {
			t.Errorf("no error getting age of non-existent cache %s: %v", n, err)
		}
		if !os.IsNotExist(err) {
			t.Errorf("deleted cache exists %s: %v", n, err)
		}

		// Load non-existent cache
		if _, err := c.Load(n); err == nil {
			t.Errorf("no error loading non-existent cache")
		}
	})
}

// LoadOrStore API.
func TestCache_LoadOrStore(t *testing.T) {
	t.Parallel()

	s := "this is a test"
	var reloadCalled bool
	reload := func() ([]byte, error) {
		reloadCalled = true
		return []byte(s), nil
	}

	withTempDir(func(dir string) {
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

// Round-trip data through the JSON caching API.
func TestCache_StoreJSON(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {
		n := "test.json"
		c := NewCache(dir)
		p := c.path(n)

		// Delete non-existent store
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

		// Try to load non-existent cache
		b = &TestData{}
		if err := c.LoadJSON(n, b); err == nil {
			t.Errorf("no error loading non-existent cache")
		}
	})
}

// TestLoadOrStoreJSON tests JSON serialisation.
func TestCache_LoadOrStoreJSON(t *testing.T) {
	t.Parallel()

	var reloadCalled bool
	var a, b *TestData

	reload := func() (interface{}, error) {
		reloadCalled = true
		return &TestData{"one", "two"}, nil
	}

	withTempDir(func(dir string) {
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

// Reload funcs that return errors
func TestCache_reloadError(t *testing.T) {
	t.Parallel()

	reloadB := func() ([]byte, error) {
		return nil, fmt.Errorf("an error")
	}

	reloadJSON := func() (interface{}, error) {
		return nil, fmt.Errorf("an error")
	}

	withTempDir(func(dir string) {
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
	t.Parallel()

	withTempDir(func(dir string) {
		sid := NewSessionID()
		s := NewSession(dir, sid)
		data := []byte("this is a test")
		n := "test.txt"

		// Sanity checks
		p := s.cache.path(s.name(n))
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Delete non-existent store
		if err := s.Store(n, nil); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		// Non-existent cache exists
		if s.Exists(n) {
			t.Errorf("non-existent cache exists")
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
		_ = s.Clear(false) // Leave current session data
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}
		// Clear this session's data, too
		_ = s.Clear(true)
		if util.PathExists(p) {
			t.Errorf("cache file exists: %s", p)
		}

		// Load non-existent cache
		if _, err := s.Load(n); err == nil {
			t.Errorf("no error loading non-existent cache")
		}

		// Clear old sessions
		sid1 := NewSessionID()
		sid2 := NewSessionID()
		s = NewSession(dir, sid1)
		_ = s.Store(n, data)

		if !s.Exists(n) {
			t.Errorf("cached data do not exist: %s", n)
		}

		s = NewSession(dir, sid2)
		_ = s.Clear(false)

		if s.Exists(n) {
			t.Errorf("expired data still exist: %s", n)
		}
	})
}
