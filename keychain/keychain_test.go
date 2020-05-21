// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package keychain

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeychain(t *testing.T) {
	t.Parallel()

	var (
		service   = "net.deanishe.awgo"
		name      = "test_password"
		password  = "test_secret"
		password2 = "tëst_sécrét"
		kc        = New(service)
	)

	t.Run("sanity check", func(t *testing.T) {
		// ensure test account doesn't exist
		cmd := exec.Command("/usr/bin/security", "delete-generic-password", "-s", service, "-a", name)
		if err := cmd.Run(); err != nil {
			require.Equal(t, 44, cmd.ProcessState.ExitCode(), "unexpected exit code")
		}
	})

	t.Run("missing items", func(t *testing.T) {
		assert.Equal(t, ErrNotFound, kc.Delete(name), "delete missing item did not fail")
		_, err := kc.Get(name)
		assert.Equal(t, ErrNotFound, err, "retrieve missing item did not fail")
	})

	t.Run("set password", func(t *testing.T) {
		assert.Nil(t, kc.Set(name, password), "set password failed")
		v, err := kc.Get(name)
		assert.Nil(t, err, "get password failed")
		assert.Equal(t, password, v, "unexpected password")
	})

	t.Run("change password", func(t *testing.T) {
		assert.Nil(t, kc.Set(name, password2), "change password failed")
		v, err := kc.Get(name)
		assert.Nil(t, err, "get changed password failed")
		assert.Equal(t, password2, v, "unexpected password")
	})

	t.Run("delete password", func(t *testing.T) {
		assert.Nil(t, kc.Delete(name), "delete failed")
	})
}

func TestParse(t *testing.T) {
	t.Parallel()

	data := []struct {
		in string
		pw string
	}{
		// ascii passwords
		{in: `password: "hunter2"`, pw: "hunter2"},
		{in: `password: "hunter two"`, pw: "hunter two"},
		// unicode passwords
		{in: `password: 0x74C3AB73745F73C3A96372C3A974  "t\303\253st_s\303\251cr\303\251t"`, pw: "tëst_sécrét"},
		{in: `password: 0x68C3BC6E74657232  "h\303\274nter2"`, pw: "hünter2"},

		// invalid
		{in: ``, pw: ""},
		{in: `password: `, pw: ""},
		{in: `password: 0x"`, pw: ""},
		{in: `password: 0xinvalid`, pw: ""},
	}

	for _, td := range data {
		td := td
		t.Run(td.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, td.pw, parseKeychainPassword(td.in), "unexpected password")
		})
	}
}
