// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deanishe/awgo/util"
)

var (
	// Filenames of session cache files are prefixed with this string
	sessionPrefix = "_aw_session"
	sidLength     = 24
	letters       = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Cache implements a simple store/load API, saving data to specified directory.
//
// There are two APIs, one for storing/loading bytes and one for
// marshalling and storing/loading and unmarshalling JSON.
//
// Each API has basic Store/Load functions plus a LoadOrStore function which
// loads cached data if these exist and aren't too old, or retrieves new data
// via the provided function, then caches and returns these.
//
// The `name` parameter passed to Load*/Store* methods is used as the filename
// for the on-disk cache, so make sure it's filesystem-safe, and consider
// adding an appropriate extension to the name, e.g. use "name.txt" (or
// "name.json" with LoadOrStoreJSON).
type Cache struct {
	Dir string // Directory to save data in
}

// NewCache creates a new Cache using given directory.
// Directory is created if it doesn't exist. Panics if directory can't be created.
func NewCache(dir string) *Cache {
	util.MustExist(dir)
	return &Cache{dir}
}

// Store saves data under the given name. If data is nil, the cache is deleted.
func (c Cache) Store(name string, data []byte) error {
	p := c.path(name)
	if data == nil {
		if util.PathExists(p) {
			return os.Remove(p)
		}
		return nil
	}
	return util.WriteFile(p, data, 0600)
}

// StoreJSON serialises v to JSON and saves it to the cache. If v is nil,
// the cache is deleted.
func (c Cache) StoreJSON(name string, v interface{}) error {
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
func (c Cache) Load(name string) ([]byte, error) {
	p := c.path(name)
	if _, err := os.Stat(p); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(p)
}

// LoadJSON unmarshals named cache into v.
func (c Cache) LoadJSON(name string, v interface{}) error {
	p := c.path(name)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// LoadOrStore loads data from cache if they exist and are newer than maxAge.
// If data do not exist or are older than maxAge, the reload function is
// called, and the returned data are save to the cache and also returned.
//
// If maxAge is 0, any cached data are always returned.
func (c Cache) LoadOrStore(name string, maxAge time.Duration, reload func() ([]byte, error)) ([]byte, error) {
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
// newer than maxAge. If the data do not exist or are older than maxAge, the
// reload function is called, and the data it returns are marshalled to JSON &
// cached, and also unmarshalled into v.
//
// If maxAge is 0, any cached data are loaded regardless of age.
func (c Cache) LoadOrStoreJSON(name string, maxAge time.Duration, reload func() (interface{}, error), v interface{}) error {
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
func (c Cache) Exists(name string) bool { return util.PathExists(c.path(name)) }

// Expired returns true if the named cache does not exist or is older than maxAge.
func (c Cache) Expired(name string, maxAge time.Duration) bool {
	age, err := c.Age(name)
	if err != nil {
		return true
	}
	return age > maxAge
}

// Age returns the age of the data cached at name.
func (c Cache) Age(name string) (time.Duration, error) {
	p := c.path(name)
	fi, err := os.Stat(p)
	if err != nil {
		return time.Duration(0), err
	}
	return time.Now().Sub(fi.ModTime()), nil
}

// path returns the path to a named file within cache directory.
func (c Cache) path(name string) string { return filepath.Join(c.Dir, name) }

// Session is a Cache that is tied to the `sessionID` value passed to NewSession().
//
// All cached data are stored under the sessionID. NewSessionID() creates
// a pseudo-random string based on the current UNIX time (in nanoseconds).
// The Workflow struct persists this value as a session ID as long as the
// user is using the current workflow via the `AW_SESSION_ID` top-level
// workflow variable.
//
// As soon as Alfred closes or the user calls another workflow, this variable
// is lost and the data are "hidden". Session.Clear(false) must be called to
// actually remove the data from the cache directory, which Workflow.Run() does.
//
// In contrast to the Cache API, Session methods lack an explicit `maxAge`
// parameter. It is always `0`, i.e. cached data are always loaded regardless
// of age as long as the session is valid.
//
// TODO: Embed Cache rather than wrapping it?
type Session struct {
	SessionID string
	cache     *Cache
}

// NewSession creates and initialises a Session.
func NewSession(dir, sessionID string) *Session {
	s := &Session{sessionID, NewCache(dir)}
	return s
}

// NewSessionID returns a pseudo-random string based on the current UNIX time
// in nanoseconds.
func NewSessionID() string {
	b := make([]rune, sidLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Clear removes session-scoped cache data. If current is true, it also removes
// data cached for the current session.
func (s Session) Clear(current bool) error {
	prefix := sessionPrefix + "."
	curPrefix := fmt.Sprintf("%s.%s.", sessionPrefix, s.SessionID)

	files, err := ioutil.ReadDir(s.cache.Dir)
	if err != nil {
		return fmt.Errorf("couldn't read directory (%s): %v", s.cache.Dir, err)
	}
	for _, fi := range files {
		if !strings.HasPrefix(fi.Name(), prefix) {
			continue
		}
		if !current && strings.HasPrefix(fi.Name(), curPrefix) {
			continue
		}
		p := filepath.Join(s.cache.Dir, fi.Name())
		os.RemoveAll(p)
		log.Printf("deleted %s", p)
	}
	return nil
}

// Store saves data under the given name. If len(data) is 0, the file is
// deleted.
func (s Session) Store(name string, data []byte) error {
	return s.cache.Store(s.name(name), data)
}

// StoreJSON serialises v to JSON and saves it to the cache. If v is nil,
// the cache is deleted.
func (s Session) StoreJSON(name string, v interface{}) error {
	return s.cache.StoreJSON(s.name(name), v)
}

// Load reads data saved under given name.
func (s Session) Load(name string) ([]byte, error) {
	return s.cache.Load(s.name(name))
}

// LoadJSON unmarshals a cache into v.
func (s Session) LoadJSON(name string, v interface{}) error {
	return s.cache.LoadJSON(s.name(name), v)
}

// LoadOrStore loads data from cache if they exist. If data do not exist,
// reload is called, and the resulting data are cached & returned.
func (s Session) LoadOrStore(name string, reload func() ([]byte, error)) ([]byte, error) {
	return s.cache.LoadOrStore(s.name(name), 0, reload)
}

// LoadOrStoreJSON loads JSON-serialised data from cache if they exist.
// If the data do not exist, reload is called, and the resulting interface{}
// is cached and returned.
func (s Session) LoadOrStoreJSON(name string, reload func() (interface{}, error), v interface{}) error {
	return s.cache.LoadOrStoreJSON(s.name(name), 0, reload, v)
}

// Exists returns true if the named cache exists.
func (s Session) Exists(name string) bool {
	return s.cache.Exists(s.name(name))
}

// name prefixes name with session prefix and session ID.
func (s Session) name(name string) string {
	return fmt.Sprintf("%s.%s.%s", sessionPrefix, s.SessionID, name)
}
