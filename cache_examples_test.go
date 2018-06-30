//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-13
//

package aw

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func ExampleCache() {

	var (
		// Cache "key" (filename) and the value to store
		name  = "LastOpened"
		value = time.Now()
	)

	// Create a temporary directory for Cache to use
	dir, err := ioutil.TempDir("", "awgo-demo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// Create a new cache
	c := NewCache(dir)

	// Cache doesn't exist yet
	fmt.Println(c.Exists(name)) // -> false

	// The API uses bytes
	data, _ := value.MarshalText()

	if err := c.Store(name, data); err != nil {
		panic(err)
	}

	// Cache now exists
	fmt.Println(c.Exists(name)) // -> true

	// Load data from cache
	data, err = c.Load(name)
	if err != nil {
		panic(err)
	}

	v2 := time.Time{}
	v2.UnmarshalText(data)

	// Values are equal
	fmt.Println(value.Equal(v2)) // -> true

	// Output:
	// false
	// true
	// true
}

// LoadOrStore loads data from cache if they're fresh enough, otherwise it calls
// the reload function for new data (which it caches).
func ExampleCache_LoadOrStore() {

	var (
		name        = "Expiring"
		data        = []byte("test")
		maxAge      = time.Millisecond * 1000
		start       = time.Now()
		reloadCount int
	)

	// Create a temporary directory for Cache to use
	dir, err := ioutil.TempDir("", "awgo-demo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// Create a new cache
	c := NewCache(dir)

	// Called by LoadOrStore when cache is empty or has expired
	reload := func() ([]byte, error) {

		// Log call count
		reloadCount++
		fmt.Printf("reload #%d\n", reloadCount)

		return data, nil
	}

	// Cache is empty
	fmt.Println(c.Exists(name)) // -> false

	out, err := c.LoadOrStore(name, maxAge, reload)
	if err != nil {
		panic(err)
	}

	// Reload was called and cache exists
	fmt.Println(c.Exists(name)) // -> true

	// Value is the same
	fmt.Println(string(out) == string(data)) // -> true

	// Load again, this time from cache, not reload
	out, err = c.LoadOrStore(name, maxAge, reload)
	if err != nil {
		panic(err)
	}

	// Value is the same
	fmt.Println(string(out) == string(data)) // -> true

	// Wait for cache to expire, then try again
	time.Sleep(time.Millisecond + maxAge - time.Now().Sub(start))

	// reload is called again
	out, err = c.LoadOrStore(name, maxAge, reload)
	if err != nil {
		panic(err)
	}

	// Value is the same
	fmt.Println(string(out) == string(data)) // -> true

	// Output:
	// false
	// reload #1
	// true
	// true
	// true
	// reload #2
	// true
}

// LoadOrStoreJSON marshals JSON to/from the cache.
func ExampleCache_LoadOrStoreJSON() {

	var (
		name   = "Host"
		maxAge = time.Second * 5
	)

	// Create a temporary directory for Cache to use
	dir, err := ioutil.TempDir("", "awgo-demo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// struct to cache
	type host struct {
		Hostname string
		Port     int
	}

	// Called by LoadOrStoreJSON. Returns default host.
	// Normally, this function would do something that takes some time, like
	// fetch data from the web or an application.
	reload := func() (interface{}, error) {

		fmt.Println("reload")

		return &host{
			Hostname: "localhost",
			Port:     6000,
		}, nil
	}

	// Create a new cache
	c := NewCache(dir)

	// Cache is empty
	fmt.Println(c.Exists(name)) // -> false

	// Populate new host from cache/reload
	h := &host{}
	if err := c.LoadOrStoreJSON(name, maxAge, reload, h); err != nil {
		panic(err)
	}

	fmt.Println(h.Hostname)
	fmt.Println(h.Port)

	// Output:
	// false
	// reload
	// localhost
	// 6000
}
