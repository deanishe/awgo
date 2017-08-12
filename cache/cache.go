//
// Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2017-08-08
//

// Package cache implements data caching.
package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"git.deanishe.net/deanishe/awgo/util"
)

// Cache implements a simple store/load API, saving data to specified directory.
type Cache struct {
	Dir string // Directory to save data in
}

// New creates a new Cache using given directory.
// Directory dir is created if it doesn't exist. The function will panic
// if directory can't be created.
func New(dir string) *Cache {
	util.EnsureExists(dir)
	return &Cache{dir}
}

// Store saves data under the given name. If len(data) is 0, the file is
// deleted.
func (c *Cache) Store(name string, data []byte) error {
	p := c.path(name)
	if len(data) == 0 {
		if util.PathExists(p) {
			return os.Remove(p)
		}
		return nil
	}
	return ioutil.WriteFile(p, data, 0600)
}

// StoreJSON serialises v to JSON and saves it to the cache. If v is nil,
// the cache is deleted.
func (c *Cache) StoreJSON(name string, v interface{}) error {
	p := c.path(name)
	if v == nil {
		if util.PathExists(p) {
			return os.Remove(p)
		}
		return nil
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("couldn't marshal JSON: %v", err)
	}
	return c.Store(name, data)
}

// Load reads data saved under given name.
func (c *Cache) Load(name string) ([]byte, error) {
	// TODO: load data
	p := c.path(name)
	if _, err := os.Stat(p); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(p)
}

// LoadJSON unmarshals a cache into v.
func (c *Cache) LoadJSON(name string, v interface{}) error {
	p := c.path(name)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// LoadOrStore loads data from cache if they exist and are newer than maxAge. If
// data do not exist or are older than maxAge, reload is called, and the returned
// data are cached & returned.
//
// If maxAge is 0, any cached data are always returned.
func (c *Cache) LoadOrStore(name string, maxAge time.Duration, reload func() ([]byte, error)) ([]byte, error) {
	var load bool
	age, err := c.Age(name)
	if err != nil {
		load = true
	} else if maxAge > 0 && age > maxAge {
		load = true
	}
	// log.Printf("age=%v, maxAge=%v, load=%v", age, maxAge, load)
	if load {
		data, err := reload()
		if err != nil {
			return nil, fmt.Errorf("couldn't reload data: %v", err)
		}
		if err := c.Store(name, data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return c.Load(name)
}

// LoadOrStoreJSON loads JSON-serialised data from cache if they exist and are
// newer than maxAge. If the data do not exist or are older than maxAge, reload
// is called, and the returned interface{} is cached and returned.
//
// If maxAge is 0, any cached data are always returned.
func (c *Cache) LoadOrStoreJSON(name string, maxAge time.Duration, reload func() (interface{}, error), v interface{}) error {
	var (
		load bool
		data []byte
		err  error
	)
	age, err := c.Age(name)
	if err != nil {
		load = true
	} else if maxAge > 0 && age > maxAge {
		load = true
	}

	if load {
		i, err := reload()
		if err != nil {
			return fmt.Errorf("couldn't reload data: %v", err)
		}
		data, err = json.MarshalIndent(i, "", "  ")
		if err != nil {
			return fmt.Errorf("couldn't marshal data to JSON: %v", err)
		}
		if err := c.Store(name, data); err != nil {
			return err
		}
	} else {
		data, err = c.Load(name)
		if err != nil {
			return fmt.Errorf("couldn't load cached data: %v", err)
		}
	}
	// TODO: Is there any way to directly return i without marshalling and unmarshalling it?
	return json.Unmarshal(data, v)
}

// Exists returns true if the named cache exists.
func (c *Cache) Exists(name string) bool { return util.PathExists(c.path(name)) }

// Expired returns true if the named cache does not exist or is older than maxAge.
func (c *Cache) Expired(name string, maxAge time.Duration) bool {
	age, err := c.Age(name)
	if err != nil {
		return true
	}
	return age > maxAge
}

// Age returns the age of the data cached at name.
func (c *Cache) Age(name string) (time.Duration, error) {
	p := c.path(name)
	fi, err := os.Stat(p)
	if err != nil {
		return time.Duration(0), err
	}
	return time.Now().Sub(fi.ModTime()), nil
}

// path returns the path to a named file within cache directory.
func (c *Cache) path(name string) string { return filepath.Join(c.Dir, name) }
