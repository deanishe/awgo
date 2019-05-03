// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package keychain

import (
	"os/exec"
	"testing"
)

func TestKeychain(t *testing.T) {
	var (
		service   = "net.deanishe.awgo"
		name      = "test_password"
		password  = "test_secret"
		password2 = "tëst_sécrét"
		kc        = New(service)
	)

	// ensure test account doesn't exist
	cmd := exec.Command("/usr/bin/security", "delete-generic-password", "-s", service, "-a", name)
	if err := cmd.Run(); err != nil {
		if cmd.ProcessState.ExitCode() != 44 {
			t.Fatal(err)
		}
	}

	// Missing items
	if err := kc.Delete(name); err != ErrNotFound {
		t.Errorf("Delete missing item. Expected=ErrNotFound, Got=%v", err)
	}
	if _, err := kc.Get(name); err != ErrNotFound {
		t.Errorf("Get missing item. Expected=ErrNotFound, Got=%v", err)
	}

	// Set password
	if err := kc.Set(name, password); err != nil {
		t.Errorf("Set password. Expected=nil, Got=%v", err)
	}
	if v, err := kc.Get(name); err != nil {
		t.Errorf("Get password. Expected=nil, Got=%v", err)
	} else if v != password {
		t.Errorf("Get password. Expected=%q, Got=%q", password, v)
	}

	// Change password
	if err := kc.Set(name, password2); err != nil {
		t.Errorf("Change password. Expected=nil, Got=%v", err)
	}
	if v, err := kc.Get(name); err != nil {
		t.Errorf("Get changed password. Expected=nil, Got=%v", err)
	} else if v != password2 {
		t.Errorf("Get changed password. Expected=%q, Got=%q", password2, v)
	}

	if err := kc.Delete(name); err != nil && err != ErrNotFound {
		t.Fatal(err)
	}
}

func TestParse(t *testing.T) {
	data := []struct {
		in string
		pw string
	}{
		{in: `password: "hunter2"`, pw: "hunter2"},
		{in: `password: "hunter two"`, pw: "hunter two"},
		{in: `password: 0x74C3AB73745F73C3A96372C3A974  "t\303\253st_s\303\251cr\303\251t"`, pw: "tëst_sécrét"},
		{in: `password: 0x68C3BC6E74657232  "h\303\274nter2"`, pw: "hünter2"},
	}

	for _, td := range data {
		v := parseSecret(td.in)
		if v != td.pw {
			t.Errorf("Bad Secret. Expected=%v, Got=%v", td.pw, v)
		}
	}
}
