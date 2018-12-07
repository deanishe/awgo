// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

// testHost is the tagged struct tests load from and into.
type testHost struct {
	ID           string `env:"-"`
	Hostname     string `env:"HOST"`
	Online       bool
	Port         uint
	Score        int
	FreeSpace    int64         `env:"SPACE"`
	PingInterval time.Duration `env:"PING"`
	PingAverage  float64
}

// The default values for the bind test environment.
var (
	testID                 = "uid34"
	testHostname           = "test.example.com"
	testOnline             = true
	testPort         uint  = 3000
	testScore              = 10000
	testFreeSpace    int64 = 9876543210
	testPingInterval       = time.Second * 10
	testPingAverage        = 4.5
	testFieldCount         = 7 // Number of visible, non-ignored fields in testHost
)

// Test bindDest implementation that captures saves.
type testDest struct {
	Cfg   *Config
	Saves map[string]string
}

func (dst *testDest) setMulti(variables map[string]string, export bool) error {

	for k, v := range variables {
		dst.Saves[k] = v
	}

	return nil
}
func (dst *testDest) GetString(key string, fallback ...string) string {
	return dst.Cfg.GetString(key, fallback...)
}

// Verify checks that dst.Saves has the same content as saves.
func (dst *testDest) Verify(saves map[string]string) error {

	var err error

	for k, x := range saves {

		s, ok := dst.Saves[k]
		if !ok {
			err = fmt.Errorf("Key %s was not set", k)
			break
		}

		if s != x {
			err = fmt.Errorf("Bad %s. Expected=%v, Got=%v", k, x, s)
			break
		}

	}

	if err == nil && len(dst.Saves) != len(saves) {
		err = fmt.Errorf("Different lengths. Expected=%d, Got=%d", len(saves), len(dst.Saves))
	}

	if err != nil {
		log.Printf("Expected=\n%#v\nGot=\n%#v", saves, dst.Saves)
	}
	return err
}

// Returns a test implementation of Env
func bindTestEnv() MapEnv {

	return MapEnv{
		"ID":           "not empty",
		"HOST":         testHostname,
		"ONLINE":       fmt.Sprintf("%v", testOnline),
		"PORT":         fmt.Sprintf("%d", testPort),
		"SCORE":        fmt.Sprintf("%d", testScore),
		"SPACE":        fmt.Sprintf("%d", testFreeSpace),
		"PING":         fmt.Sprintf("%s", testPingInterval),
		"PING_AVERAGE": fmt.Sprintf("%0.1f", testPingAverage),
	}
}

// TestExtract verifies extraction of struct fields and tags.
func TestExtract(t *testing.T) {

	cfg := NewConfig()
	th := &testHost{}
	data := map[string]string{
		"Hostname":     "HOST",
		"Online":       "ONLINE",
		"Port":         "PORT",
		"Score":        "SCORE",
		"FreeSpace":    "SPACE",
		"PingInterval": "PING",
		"PingAverage":  "PING_AVERAGE",
	}

	binds, err := extract(th)
	if err != nil {
		t.Fatalf("couldn't extract testHost: %v", err)
	}

	if len(binds) != testFieldCount {
		t.Errorf("Bad Bindings count. Expected=%d, Got=%d",
			testFieldCount, len(binds))
	}

	x := map[string]string{}
	for _, bind := range binds {
		x[bind.Name] = bind.EnvVar
	}

	if err := verifyMapsEqual(x, data); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	// Field not found error
	st := struct {
		Host string
		Port uint
	}{}

	binds, err = extract(&st)
	if err != nil {
		t.Fatalf("couldn't extract struct: %v", err)
	}
	// Change field numbers
	for _, bind := range binds {
		bind.FieldNum = bind.FieldNum + 1000
	}
	// Fail to load fields
	for _, bind := range binds {
		if err := bind.Import(cfg); err == nil {
			t.Errorf("Accepted bad binding (%s)", bind.Name)
		}
	}
}

// TestConfigTo verifies that a struct is correctly populated from a Config.
func TestConfigTo(t *testing.T) {

	h := &testHost{}
	e := bindTestEnv()
	cfg := NewConfig(e)

	if err := cfg.To(h); err != nil {
		t.Fatal(err)
	}

	if h.ID != "" { // ID is ignored
		t.Errorf("Bad ID. Expected=, Got=%v", h.ID)
	}

	if h.Hostname != testHostname {
		t.Errorf("Bad Hostname. Expected=%v, Got=%v", testHostname, h.Hostname)
	}

	if h.Online != testOnline {
		t.Errorf("Bad Online. Expected=%v, Got=%v", testOnline, h.Online)
	}

	if h.Port != testPort {
		t.Errorf("Bad Port. Expected=%v, Got=%v", testPort, h.Port)
	}

	if h.Score != testScore {
		t.Errorf("Bad Score. Expected=%v, Got=%v", testScore, h.Score)
	}

	if h.FreeSpace != testFreeSpace {
		t.Errorf("Bad FreeSpace. Expected=%v, Got=%v", testFreeSpace, h.FreeSpace)
	}

	if h.PingInterval != testPingInterval {
		t.Errorf("Bad PingInterval. Expected=%v, Got=%v", testPingInterval, h.PingInterval)
	}

	if h.PingAverage != testPingAverage {
		t.Errorf("Bad PingAverage. Expected=%v, Got=%v", testPingAverage, h.PingAverage)
	}

}

// TestConfigFrom verifies that a bindDest is correctly populated from a (tagged) struct.
func TestConfigFrom(t *testing.T) {

	e := MapEnv{
		"ID":           "",
		"HOST":         "",
		"ONLINE":       "true", // must be set: "" is the same as false
		"PORT":         "",
		"SCORE":        "",
		"SPACE":        "",
		"PING":         "0s",  // zero value
		"PING_AVERAGE": "0.0", // zero value
	}

	cfg := NewConfig(e)
	th := &testHost{}

	// Check Config is set up correctly
	for k, v := range e {
		s := cfg.Get(k)
		if s != v {
			t.Errorf("Bad %s. Expected=%v, Got=%v", k, v, s)
		}
	}

	var (
		newHostname           = "www.aol.com"
		newOnline             = false
		newPort         uint  = 2500
		newScore              = 7602
		newFreeSpace    int64 = 1234567890
		newPingInterval       = time.Minute * 2
		newPingAverage        = 3.3

		// How the testDest should look afterwards
		one = map[string]string{
			"HOST":   newHostname,
			"ONLINE": fmt.Sprintf("%v", newOnline),
		}
		two = map[string]string{
			"PORT":  fmt.Sprintf("%d", newPort),
			"SCORE": fmt.Sprintf("%d", newScore),
			"SPACE": fmt.Sprintf("%d", newFreeSpace),
		}
		three = map[string]string{
			"PING":         fmt.Sprintf("%s", newPingInterval),
			"PING_AVERAGE": fmt.Sprintf("%0.1f", newPingAverage),
		}
	)

	// Exports v into a testDest and verifies it against x.
	testBind := func(v interface{}, x map[string]string) {

		dst := &testDest{cfg, map[string]string{}}

		variables, err := cfg.bindVars(v)
		if err != nil {
			t.Fatal(err)
		}

		if err := dst.setMulti(variables, false); err != nil {
			t.Fatal(err)
		}

		if err := dst.Verify(x); err != nil {
			t.Errorf("%s", err)
		}
	}

	// Expected testDest value
	x := map[string]string{}

	th.Hostname = newHostname
	th.Online = newOnline

	for k, v := range one {
		x[k] = v
	}
	testBind(th, x)

	th.Port = newPort
	th.Score = newScore
	th.FreeSpace = newFreeSpace

	for k, v := range two {
		x[k] = v
	}
	testBind(th, x)

	th.PingInterval = newPingInterval
	th.PingAverage = newPingAverage

	for k, v := range three {
		x[k] = v
	}
	testBind(th, x)
}

// TestVarName tests the envvar name algorithm.
func TestVarName(t *testing.T) {
	data := []struct {
		in, out string
	}{
		{"URL", "URL"},
		{"Name", "NAME"},
		{"LastName", "LAST_NAME"},
		{"URLEncoding", "URL_ENCODING"},
		{"LongBeard", "LONG_BEARD"},
		{"HTML", "HTML"},
		{"etc", "ETC"},
	}

	for _, td := range data {
		v := EnvVarForField(td.in)
		if v != td.out {
			t.Errorf("Bad VarName (%s). Expected=%v, Got=%v",
				td.in, td.out, v)
		}
	}
}

// Populate a struct from workflow/environment variables. See EnvVarForField
// for information on how fields are mapped to environment variables if
// no variable name is specified using an `env:"name"` tag.
func ExampleConfig_To() {

	// Set some test values
	os.Setenv("USERNAME", "dave")
	os.Setenv("API_KEY", "hunter2")
	os.Setenv("INTERVAL", "5m")
	os.Setenv("FORCE", "1")

	// Program settings to load from env
	type Settings struct {
		Username       string
		APIKey         string
		UpdateInterval time.Duration `env:"INTERVAL"`
		Force          bool
	}

	var (
		s   = &Settings{}
		cfg = NewConfig()
	)

	// Populate Settings from workflow/environment variables
	if err := cfg.To(s); err != nil {
		panic(err)
	}

	fmt.Println(s.Username)
	fmt.Println(s.APIKey)
	fmt.Println(s.UpdateInterval)
	fmt.Println(s.Force)

	// Output:
	// dave
	// hunter2
	// 5m0s
	// true

	unsetEnv(
		"USERNAME",
		"API_KEY",
		"INTERVAL",
		"FORCE",
	)
}

// Rules for generating an environment variable name from a struct field name.
func ExampleEnvVarForField() {
	// Single-case words are upper-cased
	fmt.Println(EnvVarForField("URL"))
	fmt.Println(EnvVarForField("name"))
	// Words that start with fewer than 3 uppercase chars are upper-cased
	fmt.Println(EnvVarForField("Folder"))
	fmt.Println(EnvVarForField("MTime"))
	// But with 3+ uppercase chars, the last is treated as the first
	// char of the next word
	fmt.Println(EnvVarForField("VIPath"))
	fmt.Println(EnvVarForField("URLEncoding"))
	fmt.Println(EnvVarForField("SSLPort"))
	// Camel-case words are split on case changes
	fmt.Println(EnvVarForField("LastName"))
	fmt.Println(EnvVarForField("LongHorse"))
	fmt.Println(EnvVarForField("loginURL"))
	fmt.Println(EnvVarForField("newHomeAddress"))
	fmt.Println(EnvVarForField("PointA"))
	// Digits are considered the end of a word, not the start
	fmt.Println(EnvVarForField("b2B"))

	// Output:
	// URL
	// NAME
	// FOLDER
	// MTIME
	// VI_PATH
	// URL_ENCODING
	// SSL_PORT
	// LAST_NAME
	// LONG_HORSE
	// LOGIN_URL
	// NEW_HOME_ADDRESS
	// POINT_A
	// B2_B
}

// Verify zero and non-zero values returned by isZeroValue.
func TestIsZeroValue(t *testing.T) {

	zero := struct {
		String     string
		Int        int
		Int8       int8
		Int16      int16
		Int32      int32
		Int64      int64
		Float32    float32
		Float64    float64
		Duration   time.Duration
		NilPointer *time.Time
		// struct & map not supported
		// Time       time.Time
		// Map        map[string]string
	}{}
	nonzero := struct {
		String   string
		Int      int
		Int8     int8
		Int16    int16
		Int32    int32
		Int64    int64
		Float32  float32
		Float64  float64
		Duration time.Duration
		Pointer  *time.Time
		Time     time.Time
		Map      map[string]string
	}{

		String:   "word",
		Int:      5,
		Int8:     5,
		Int16:    5,
		Int32:    5,
		Int64:    5,
		Float32:  1.5,
		Float64:  2.51,
		Duration: time.Minute * 5,
		Pointer:  &time.Time{},
		Time:     time.Now(),
		Map:      map[string]string{},
	}

	rv := reflect.ValueOf(zero)
	typ := rv.Type()

	for i := 0; i < rv.NumField(); i++ {

		field := typ.Field(i)
		value := rv.Field(i)

		v := isZeroValue(value)
		if v != true {
			t.Errorf("Bad %s. Expected=true, Got=false", field.Name)
		}
	}

	rv = reflect.ValueOf(nonzero)
	typ = rv.Type()

	for i := 0; i < rv.NumField(); i++ {

		field := typ.Field(i)
		value := rv.Field(i)

		v := isZeroValue(value)
		if v == true {
			t.Errorf("Bad %s. Expected=false, Got=true", field.Name)
		}
	}
}

// Verify *string* zero values for other types, e.g. "0" is zero value for int.
func TestIsZeroString(t *testing.T) {
	data := []struct {
		in   string
		kind reflect.Kind
		x    bool
	}{
		{"", reflect.String, true},
		{" ", reflect.String, false},
		{"test", reflect.String, false},
		// Ints
		{"", reflect.Int, true},
		{"0", reflect.Int, true},
		{"0000", reflect.Int, true},
		{"", reflect.Int8, true},
		{"0", reflect.Int8, true},
		{"0000", reflect.Int8, true},
		{"", reflect.Int16, true},
		{"0", reflect.Int16, true},
		{"0000", reflect.Int16, true},
		{"", reflect.Int32, true},
		{"0", reflect.Int32, true},
		{"0000", reflect.Int32, true},
		{"", reflect.Int64, true},
		{"0", reflect.Int64, true},
		{"0000", reflect.Int64, true},
		{"1,23", reflect.Int64, false},
		{"test", reflect.Int64, false},
		// Floats
		{"", reflect.Float32, true},
		{"0", reflect.Float32, true},
		{"0.0", reflect.Float32, true},
		{"00.00", reflect.Float32, true},
		{"1,23", reflect.Float32, false},
		{"test", reflect.Float32, false},
		{"", reflect.Float64, true},
		{"0", reflect.Float64, true},
		{"0.0", reflect.Float64, true},
		{"00.00", reflect.Float64, true},
		{"1,23", reflect.Float64, false},
		{"test", reflect.Float64, false},
		// Booleans
		{"", reflect.Bool, true},
		{"0", reflect.Bool, true},
		{"false", reflect.Bool, true},
		{"FALSE", reflect.Bool, true},
		{"f", reflect.Bool, true},
		{"F", reflect.Bool, true},
		{"False", reflect.Bool, true},
		// Durations
		{"0s", reflect.Int64, true},
		{"0m0s", reflect.Int64, true},
		{"0h0m", reflect.Int64, true},
		{"0h0m0s", reflect.Int64, true},
		{"1s", reflect.Int64, false},
		{"2ms", reflect.Int64, false},
		{"1h5m", reflect.Int64, false},
		{"96h", reflect.Int64, false},
		{"12h", reflect.Int64, false},
	}

	for _, td := range data {

		v := isZeroString(td.in, td.kind)
		if v != td.x {
			t.Errorf("Bad ZeroString (%s). Expected=%v, Got=%v", td.in, td.x, v)
		}
	}
}

func TestIsCamelCase(t *testing.T) {
	data := []struct {
		s string
		v bool
	}{
		{"", false},
		{"URL", false},
		{"url", false},
		{"Url", false},
		{"HomeAddress", true},
		{"myHomeAddress", true},
		{"PlaceA", true},
		{"myPlaceB", true},
		{"myB", true},
		{"my2B", true},
		{"B2B", false},
		{"SSLPort", true},
	}

	for _, td := range data {
		b := isCamelCase(td.s)
		if b != td.v {
			t.Errorf("Bad CamelCase (%s). Expected=%v, Got=%v", td.s, td.v, b)
		}

	}
}

func TestSplitCamelCase(t *testing.T) {
	data := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"HomeAddress", "HOME_ADDRESS"},
		{"homeAddress", "HOME_ADDRESS"},
		{"loginURL", "LOGIN_URL"},
		{"SSLPort", "SSL_PORT"},
		{"HomeAddress", "HOME_ADDRESS"},
		{"myHomeAddress", "MY_HOME_ADDRESS"},
		{"PlaceA", "PLACE_A"},
		{"myPlaceB", "MY_PLACE_B"},
		{"myB", "MY_B"},
		{"my2B", "MY2_B"},
		{"URLEncoding", "URL_ENCODING"},
	}

	for _, td := range data {
		s := splitCamelCase(td.in)
		if s != td.out {
			t.Errorf("Bad SplitCamel (%s). Expected=%v, Got=%v", td.in, td.out, s)
		}

	}
}

func TestIsBindable(t *testing.T) {

	data := []struct {
		k reflect.Kind
		x bool
	}{
		{reflect.Ptr, false},
		{reflect.Map, false},
		{reflect.Slice, false},
		{reflect.Struct, false},
		{reflect.String, true},
		{reflect.Bool, true},
		{reflect.Int, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Uint, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Float32, true},
		{reflect.Float64, true},
	}

	for _, td := range data {
		v := isBindable(td.k)
		if v != td.x {
			t.Errorf("Bad Bindable for %v. Expected=%v, Got=%v", td.k, td.x, v)
		}
	}
}
