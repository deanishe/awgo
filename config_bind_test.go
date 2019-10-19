// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	testHostname           = "test.example.com"
	testOnline             = true
	testPort         uint  = 3000
	testScore              = 10000
	testFreeSpace    int64 = 9876543210
	testPingInterval       = time.Second * 10
	testPingAverage        = 4.5
	testFieldCount         = 7 // Number of visible, non-ignored fields in testHost
)

var (
	privTestName     = "Hello World"
	privTestQuoted   = `"QUOTED"`
	privTestEmpty    = ""
	privTestBool     = true
	privTestDuration = time.Minute * 5
	privTestInt      = 10
	privTestFloat    = 6.6

	privTestEnv = MapEnv{
		"AWGO_TEST_NAME":     privTestName,
		"AWGO_TEST_QUOTED":   privTestQuoted,
		"AWGO_TEST_EMPTY":    privTestEmpty,
		"AWGO_TEST_BOOL":     fmt.Sprintf("%v", privTestBool),
		"AWGO_TEST_DURATION": fmt.Sprintf("%v", privTestDuration),
		"AWGO_TEST_INT":      fmt.Sprintf("%d", privTestInt),
		"AWGO_TEST_FLOAT":    fmt.Sprintf("%f", privTestFloat),
	}

	privTestSrc = struct {
		TestName     string
		TestQuoted   string
		TestEmpty    string
		TestBool     bool
		TestDuration time.Duration
		TestInt      int
		TestFloat    float64
	}{
		TestName:     privTestName,
		TestQuoted:   privTestQuoted,
		TestEmpty:    privTestEmpty,
		TestBool:     privTestBool,
		TestDuration: privTestDuration,
		TestInt:      privTestInt,
		TestFloat:    privTestFloat,
	}
)

type mockJSRunner struct {
	script string
}

func (mj *mockJSRunner) Run(script string) error {
	mj.script = script
	return nil
}

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

// Returns a test implementation of Env
func bindTestEnv() MapEnv {
	return MapEnv{
		"ID":           "not empty",
		"HOST":         testHostname,
		"ONLINE":       fmt.Sprintf("%v", testOnline),
		"PORT":         fmt.Sprintf("%d", testPort),
		"SCORE":        fmt.Sprintf("%d", testScore),
		"SPACE":        fmt.Sprintf("%d", testFreeSpace),
		"PING":         testPingInterval.String(),
		"PING_AVERAGE": fmt.Sprintf("%0.1f", testPingAverage),
	}
}

// TestExtract verifies extraction of struct fields and tags.
func TestExtract(t *testing.T) {
	t.Parallel()

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
	assert.Nil(t, err, "extract testHost failed")
	assert.Equal(t, testFieldCount, len(binds), "unexpected binding count")

	x := map[string]string{}
	for _, bind := range binds {
		x[bind.Name] = bind.EnvVar
	}

	assert.Equal(t, x, data, "extract failed")

	// Field not found error
	st := struct {
		Host string
		Port uint
	}{}

	binds, err = extract(&st)
	assert.Nil(t, err, "extract failed")

	// Change field numbers
	for _, bind := range binds {
		bind.FieldNum += 1000
	}
	// Fail to load fields
	for _, bind := range binds {
		assert.NotNil(t, bind.Import(cfg), "accepted bad binding")
	}
}

// TestConfigTo verifies that a struct is correctly populated from a Config.
func TestConfig_To(t *testing.T) {
	t.Parallel()

	h := &testHost{}
	e := bindTestEnv()
	cfg := NewConfig(e)

	require.Nil(t, cfg.To(h), "cfg.To failed")
	assert.Equal(t, "", h.ID, "unexpected ID") // ID is ignored
	assert.Equal(t, testHostname, h.Hostname, "unexpected Hostname")
	assert.Equal(t, testOnline, h.Online, "unexpected Online")
	assert.Equal(t, testPort, h.Port, "unexpected Port")
	assert.Equal(t, testScore, h.Score, "unexpected Score")
	assert.Equal(t, testFreeSpace, h.FreeSpace, "unexpected FreeSpace")
	assert.Equal(t, testPingInterval, h.PingInterval, "unexpected PingInterval")
	assert.Equal(t, testPingAverage, h.PingAverage, "unexpected PingAverage")
}

// generated script
func TestConfig_Do(t *testing.T) {
	orig := runJS
	defer func() { runJS = orig }()
	mj := &mockJSRunner{}
	runJS = mj.Run

	cfg := NewConfig(MapEnv{
		EnvVarAlfredVersion: "4.0.4",
		EnvVarBundleID:      "net.deanishe.awgo",
	})
	keys := []string{}
	for k := range privTestEnv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cfg.Set(k, privTestEnv[k], false)
	}
	assert.Nil(t, cfg.Do(), "create test env failed")

	x := `Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_BOOL", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"true"});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_DURATION", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"5m0s"});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_EMPTY", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":""});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_FLOAT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"6.600000"});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_INT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"10"});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_NAME", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"Hello World"});
Application("com.runningwithcrayons.Alfred").setConfiguration("AWGO_TEST_QUOTED", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"\"QUOTED\""});`
	assert.Equal(t, x, mj.script, "bad script")

	// no scripts, should return error
	assert.NotNil(t, cfg.Do(), "empty scripts did not return error")

	for _, k := range keys {
		cfg.Unset(k)
	}
	assert.Nil(t, cfg.Do(), "delete test env failed")

	x = `Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_BOOL", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_DURATION", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_EMPTY", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_FLOAT", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_INT", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_NAME", {"inWorkflow":"net.deanishe.awgo"});
Application("com.runningwithcrayons.Alfred").removeConfiguration("AWGO_TEST_QUOTED", {"inWorkflow":"net.deanishe.awgo"});`
	assert.Equal(t, x, mj.script, "bad script")
}

// verify that a bindDest is correctly populated from a (tagged) struct.
func TestConfig_From(t *testing.T) {
	t.Parallel()

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
		assert.Equal(t, v, cfg.Get(k), "unexpected value")
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
			"PING":         newPingInterval.String(),
			"PING_AVERAGE": fmt.Sprintf("%0.1f", newPingAverage),
		}
	)

	// Exports v into a testDest and verifies it against x.
	testBind := func(v interface{}, x map[string]string) {
		dst := &testDest{cfg, map[string]string{}}

		variables, err := cfg.bindVars(v)
		assert.Nil(t, err, "bindVars failed")
		err = dst.setMulti(variables, false)
		require.Nil(t, err, "setMulti failed")
		assert.Equal(t, x, dst.Saves, "unexpected saves")
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

// generated script
func TestConfig_From_script(t *testing.T) {
	orig := runJS
	defer func() { runJS = orig }()
	mj := &mockJSRunner{}
	runJS = mj.Run

	cfg := NewConfig(MapEnv{
		EnvVarAlfredVersion: "4.0.4",
		EnvVarBundleID:      "net.deanishe.awgo",
	})

	require.Nil(t, cfg.From(privTestSrc), "cfg.From failed")

	x := `Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_BOOL", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"true"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_DURATION", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"5m0s"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_FLOAT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"6.6"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_INT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"10"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_NAME", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"Hello World"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_QUOTED", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"\"QUOTED\""});`
	assert.Equal(t, x, mj.script, "bad script")
}

// envvar name algorithm.
func TestEnvVarForField(t *testing.T) {
	t.Parallel()
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
		td := td // capture variable
		t.Run(fmt.Sprintf("VarName(%s)", td.in), func(t *testing.T) {
			t.Parallel()
			v := EnvVarForField(td.in)
			if v != td.out {
				t.Errorf("Expected=%v, Got=%v", td.out, v)
			}
		})
	}
}

// Populate a struct from workflow/environment variables. See EnvVarForField
// for information on how fields are mapped to environment variables if
// no variable name is specified using an `env:"name"` tag.
func ExampleConfig_To() {
	// Set some test values
	_ = os.Setenv("USERNAME", "dave")
	_ = os.Setenv("API_KEY", "hunter2")
	_ = os.Setenv("INTERVAL", "5m")
	_ = os.Setenv("FORCE", "1")

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
func TestZeroValue(t *testing.T) {
	t.Parallel()

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

	t.Run("isZeroValueStruct", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(zero)
		typ := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			field := typ.Field(i)
			value := rv.Field(i)

			assert.Truef(t, isZeroValue(value), "zero value not recognised for %q", field.Name)
		}
	})

	t.Run("isNonZeroValueStruct", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(nonzero)
		typ := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			field := typ.Field(i)
			value := rv.Field(i)

			assert.Falsef(t, isZeroValue(value), "non-zero value not recognised for %q", field.Name)
		}
	})
}

// Verify *string* zero values for other types, e.g. "0" is zero value for int.
func TestZeroString(t *testing.T) {
	t.Parallel()

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
		{"invalid", reflect.Int, false},
		{"", reflect.Int8, true},
		{"0", reflect.Int8, true},
		{"0000", reflect.Int8, true},
		{"invalid", reflect.Int8, false},
		{"", reflect.Int16, true},
		{"0", reflect.Int16, true},
		{"0000", reflect.Int16, true},
		{"invalid", reflect.Int16, false},
		{"", reflect.Int32, true},
		{"0", reflect.Int32, true},
		{"0000", reflect.Int32, true},
		{"invalid", reflect.Int32, false},
		{"", reflect.Int64, true},
		{"0", reflect.Int64, true},
		{"0000", reflect.Int64, true},
		{"1,23", reflect.Int64, false},
		{"invalid", reflect.Int64, false},
		// Uints
		{"", reflect.Uint, true},
		{"0", reflect.Uint, true},
		{"0000", reflect.Uint, true},
		{"invalid", reflect.Uint, false},
		{"", reflect.Uint8, true},
		{"0", reflect.Uint8, true},
		{"0000", reflect.Uint8, true},
		{"invalid", reflect.Uint8, false},
		{"", reflect.Uint16, true},
		{"0", reflect.Uint16, true},
		{"0000", reflect.Uint16, true},
		{"invalid", reflect.Uint16, false},
		{"", reflect.Uint32, true},
		{"0", reflect.Uint32, true},
		{"0000", reflect.Uint32, true},
		{"invalid", reflect.Uint32, false},
		{"", reflect.Uint64, true},
		{"0", reflect.Uint64, true},
		{"0000", reflect.Uint64, true},
		{"1,23", reflect.Uint64, false},
		{"invalid", reflect.Uint64, false},
		// Floats
		{"", reflect.Float32, true},
		{"0", reflect.Float32, true},
		{"0.0", reflect.Float32, true},
		{"00.00", reflect.Float32, true},
		{"1,23", reflect.Float32, false},
		{"invalid", reflect.Float32, false},
		{"", reflect.Float64, true},
		{"0", reflect.Float64, true},
		{"0.0", reflect.Float64, true},
		{"00.00", reflect.Float64, true},
		{"1,23", reflect.Float64, false},
		{"invalid", reflect.Float64, false},
		// Booleans
		{"", reflect.Bool, true},
		{"0", reflect.Bool, true},
		{"false", reflect.Bool, true},
		{"FALSE", reflect.Bool, true},
		{"f", reflect.Bool, true},
		{"F", reflect.Bool, true},
		{"False", reflect.Bool, true},
		{"invalid", reflect.Bool, false},
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
		td := td // capture variable
		t.Run(fmt.Sprintf("%q", td.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, isZeroString(td.in, td.kind), "unexpected value")
		})
	}
}

func TestIsCamelCase(t *testing.T) {
	t.Parallel()

	data := []struct {
		in string
		x  bool
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
		td := td // capture variable
		t.Run(fmt.Sprintf("isCamelCase(%q)", td.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, isCamelCase(td.in), "unexpected camelcase")
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in string
		x  string
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

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("splitCamelCase(%q)", td.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, splitCamelCase(td.in), "unexpected split")
		})
	}
}

func TestIsBindable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in reflect.Kind
		x  bool
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

	for _, td := range tests {
		td := td // capture variable
		t.Run(fmt.Sprintf("IsBindable(%v)", td.in), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.x, isBindable(td.in), "unexpected bindable")
		})
	}
}
