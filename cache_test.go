// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

// Session-scoped caching.
func TestSession_Load(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {
		var (
			sid   = NewSessionID()
			s     = NewSession(dir, sid)
			data  = []byte("this is a test")
			data2 []byte
			n     = "test.txt"
			err   error
		)

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
		data2, err = s.Load(n)
		assert.Nil(t, err, "load cached data failed")
		assert.Equal(t, data, data2, "loaded data != saved data")

		// Clear session
		_ = s.Clear(false) // Leave current session data
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}
		// Clear this session's data, too
		_ = s.Clear(true)
		if util.PathExists(p) {
			t.Errorf("cleared cache file still exists: %s", p)
		}

		// Load non-existent cache
		if _, err := s.Load(n); err == nil {
			t.Errorf("no error loading non-existent cache")
		}
	})
}

func TestSession_LoadOrStore(t *testing.T) {
	withTempDir(func(dir string) {
		var (
			sid    = NewSessionID()
			s      = NewSession(dir, sid)
			data   = []byte("this is a test")
			data2  []byte
			n      = "test.txt"
			called bool
			err    error
		)

		reload := func() ([]byte, error) {
			called = true
			return data, nil
		}

		// Sanity checks
		p := s.cache.path(s.name(n))
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// LoadOrStore API
		data2, err = s.LoadOrStore(n, reload)
		assert.Nil(t, err, "LoadOrStore return error")
		assert.Equal(t, data, data2, "returned data != reload data")
		assert.True(t, called, "reload not called")

		called = false
		data2, err = s.LoadOrStore(n, reload)
		assert.Nil(t, err, "LoadOrStore return error")
		assert.Equal(t, data, data2, "returned data != reload data")
		assert.False(t, called, "reload called")
	})
}

func TestSession_LoadJSON(t *testing.T) {
	t.Parallel()

	withTempDir(func(dir string) {
		var (
			sid   = NewSessionID()
			s     = NewSession(dir, sid)
			data  = map[string]string{"foo": "bar"}
			data2 map[string]string
			n     = "test.txt"
			err   error
		)

		// Sanity checks
		p := s.cache.path(s.name(n))
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// Delete non-existent store
		if err := s.StoreJSON(n, nil); err != nil {
			t.Errorf("unexpected error clearing cache: %v", err)
		}

		// Non-existent cache exists
		if s.Exists(n) {
			t.Errorf("non-existent cache exists")
		}

		// Store data
		if err := s.StoreJSON(n, data); err != nil {
			t.Errorf("couldn't cache data to %s: %v", n, err)
		}
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		if s.Exists(n) != util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}

		// Load data
		err = s.LoadJSON(n, &data2)
		assert.Nil(t, err, "load cached data failed")
		assert.Equal(t, data, data2, "loaded data != saved data")

		// Clear session
		_ = s.Clear(false) // Leave current session data
		if !util.PathExists(p) {
			t.Errorf("cache file does not exist: %s", p)
		}
		// Clear this session's data, too
		_ = s.Clear(true)
		if util.PathExists(p) {
			t.Errorf("cleared cache file still exists: %s", p)
		}

		// Load non-existent cache
		if err := s.LoadJSON(n, &data2); err == nil {
			t.Errorf("no error loading non-existent cache")
		}
	})
}

func TestSession_LoadOrStoreJSON(t *testing.T) {
	withTempDir(func(dir string) {
		var (
			sid    = NewSessionID()
			s      = NewSession(dir, sid)
			data   = map[string]string{"foo": "bar"}
			data2  map[string]string
			n      = "test.txt"
			called bool
			err    error
		)

		reload := func() (interface{}, error) {
			called = true
			return data, nil
		}

		// Sanity checks
		p := s.cache.path(s.name(n))
		if util.PathExists(p) {
			t.Errorf("cache file already exists: %s", p)
		}

		// LoadOrStore API
		err = s.LoadOrStoreJSON(n, reload, &data2)
		assert.Nil(t, err, "LoadOrStore return error")
		assert.Equal(t, data, data2, "returned data != reload data")
		assert.True(t, called, "reload not called")

		called = false
		err = s.LoadOrStoreJSON(n, reload, &data2)
		assert.Nil(t, err, "LoadOrStore return error")
		assert.Equal(t, data, data2, "returned data != reload data")
		assert.False(t, called, "reload called")
	})
}

func TestSession_Clear(t *testing.T) {
	withTempDir(func(dir string) {
		var (
			sid1 = NewSessionID()
			sid2 = NewSessionID()
			data = []byte("this is a test")
			n    = "test.txt"
		)

		// "old" session
		s := NewSession(dir, sid1)
		if err := s.Store(n, data); err != nil {
			t.Fatalf("store failed: %v", err)
		}

		if !s.Exists(n) {
			t.Errorf("cached data do not exist: %s", n)
		}

		// "new" session
		s = NewSession(dir, sid2)
		if err := s.Clear(false); err != nil {
			t.Fatalf("clear failed: %v", err)
		}

		if s.Exists(n) {
			t.Errorf("expired data still exist: %s", n)
		}
	})
}
