// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.deanishe.net/env"
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
)

var (
	privTestName     = "Hello World"
	privTestQuoted   = `"QUOTED"`
	privTestEmpty    = ""
	privTestBool     = true
	privTestDuration = time.Minute * 5
	privTestInt      = 10
	privTestFloat    = 6.6

	privTestEnv = env.MapEnv{
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

// Returns a test implementation of Env
func bindTestEnv() env.MapEnv {
	return env.MapEnv{
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

	cfg := NewConfig(env.MapEnv{
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

// generated script
func TestConfig_From_script(t *testing.T) {
	orig := runJS
	defer func() { runJS = orig }()
	mj := &mockJSRunner{}
	runJS = mj.Run

	cfg := NewConfig(env.MapEnv{
		EnvVarAlfredVersion: "4.0.4",
		EnvVarBundleID:      "net.deanishe.awgo",
	})

	require.Nil(t, cfg.From(privTestSrc, env.IgnoreZeroValues), "cfg.From failed")

	x := `Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_BOOL", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"true"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_DURATION", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"5m0s"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_FLOAT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"6.6"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_INT", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"10"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_NAME", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"Hello World"});
Application("com.runningwithcrayons.Alfred").setConfiguration("TEST_QUOTED", {"exportable":false,"inWorkflow":"net.deanishe.awgo","toValue":"\"QUOTED\""});`
	assert.Equal(t, x, mj.script, "bad script")
}

func TestConfig_From_invalid_source(t *testing.T) {
	invalid := []interface{}{
		"string",
		[]string{},
		map[string]string{},
		int(10),
	}

	cfg := NewConfig()

	for _, v := range invalid {
		assert.EqualError(t, cfg.From(v), "not a struct", "dump accepted invalid target")
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
