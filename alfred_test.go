//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-11
//

package aw

import (
	"os"
	"testing"
)

// Whether to run tests that actually call Alfred
var (
	testSearch   = false
	testAction   = false
	testBrowse   = false
	testTrigger  = false
	testSetConf  = false
	testRmConf   = false
	testSetTheme = false
)

func TestAlfred(t *testing.T) {

	a := NewAlfred()

	if testSearch {

		if err := a.Search("").Do(); err != nil {
			t.Error(err)
		}

		if err := a.Search("awgo alfred").Do(); err != nil {
			t.Error(err)
		}
	}

	if testAction {

		h := os.ExpandEnv("$HOME")

		if err := a.Action(h+"/Desktop", ".").Do(); err != nil {
			t.Error(err)
		}
	}

	if testBrowse {

		if err := a.Browse(".").Do(); err != nil {
			t.Error(err)
		}
	}

	if testTrigger {

		if err := a.RunTrigger("test", "AwGo, yo!").Do(); err != nil {
			t.Error(err)
		}
	}

	if testSetConf {

		if err := a.SetConfig("AWGO_TEST_UNITTEST", "AwGo, yo!", true).Do(); err != nil {
			t.Error(err)
		}

		many := map[string]string{
			"MANY_0": "VALUE_0",
			"MANY_1": "VALUE_1",
			"MANY_2": "VALUE_2",
			"MANY_3": "VALUE_3",
			"MANY_4": "VALUE_4",
			"MANY_5": "VALUE_5",
			"MANY_6": "VALUE_6",
			"MANY_7": "VALUE_7",
			"MANY_8": "VALUE_8",
			"MANY_9": "VALUE_9",
		}

		a := NewAlfred()
		for k, v := range many {
			a.SetConfig(k, v, true)
		}
		if err := a.Do(); err != nil {
			t.Error(err)
		}
	}

	if testRmConf {

		keys := []string{
			"AWGO_TEST_BOOL", "AWGO_TEST_DURATION", "AWGO_TEST_EMPTY",
			"AWGO_TEST_FLOAT", "AWGO_TEST_INT", "AWGO_TEST_NAME",
			"AWGO_TEST_QUOTED", "AWGO_TEST_UNITTEST",
			"BENCH_0", "BENCH_1", "BENCH_10", "BENCH_11",
			"BENCH_12", "BENCH_13", "BENCH_14", "BENCH_15",
			"BENCH_16", "BENCH_17", "BENCH_18", "BENCH_19",
			"BENCH_2", "BENCH_20", "BENCH_21", "BENCH_22",
			"BENCH_23", "BENCH_24", "BENCH_3", "BENCH_4",
			"BENCH_5", "BENCH_6", "BENCH_7", "BENCH_8",
			"BENCH_9", "MANY_0", "MANY_1", "MANY_2",
			"MANY_3", "MANY_4", "MANY_5", "MANY_6",
			"MANY_7", "MANY_8", "MANY_9",
		}
		a := NewAlfred()

		for _, k := range keys {
			a.RemoveConfig(k)
		}

		if err := a.Do(); err != nil {
			t.Error(err)
		}
	}

	if testSetTheme {

		if err := a.SetTheme("Alfred Notepad").Do(); err != nil {
			t.Error(err)
		}
	}
}
