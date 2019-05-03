// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

// Package keychain implements a simple interface to the macOS Keychain.
// Based on /usr/bin/security.
package keychain

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Specific errors returned by the API.
var (
	// Returned by Keychain.Get() and Keychain.Delete() if the specified
	// account doesn't exist.
	ErrNotFound = errors.New("password not found")
	// Used internally. Swallowed by Keychain.Set() if account already exists.
	errDuplicate = errors.New("duplicate password")
)

// Keychain manages macOS Keychain passwords for a specific service.
type Keychain struct {
	service string
}

// New Keychain for specified service.
func New(service string) *Keychain {
	return &Keychain{service: service}
}

// Get password from user's Keychain. Returns ErrNotFound if specified account doesn't exist.
func (kc *Keychain) Get(account string) (string, error) {
	s, err := kc.run("find-generic-password", account, "-g")
	if err != nil {
		return "", err
	}
	if s = parseSecret(s); s == "" {
		return "", ErrNotFound
	}
	return s, nil
}

// Set password in user's Keychain. If the account already exists, it is replaced.
func (kc *Keychain) Set(account, password string) error {
	_, err := kc.run("add-generic-password", account, "-w", password)
	if err == errDuplicate {
		if err := kc.Delete(account); err != nil {
			return err
		}
		_, err = kc.run("add-generic-password", account, "-w", password)
	}
	return err
}

// Delete a password from user's Keychain. Returns ErrNotFound if account doesn't exist.
func (kc *Keychain) Delete(account string) error {
	_, err := kc.run("delete-generic-password", account)
	return err
}

// run executes a Keychain command.
func (kc *Keychain) run(command, account string, args ...string) (string, error) {
	args = append([]string{command, "-s", kc.service, "-a", account}, args...)
	cmd := exec.Command("/usr/bin/security", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(stderr)
	err = cmd.Wait()
	if err != nil {
		switch cmd.ProcessState.ExitCode() {
		case 44:
			return "", ErrNotFound
		case 45:
			return "", errDuplicate
		}
		return "", fmt.Errorf("%v: %s", err, string(data))
	}
	s := strings.TrimSpace(string(data))
	return s, nil
}

// Extract password from /usr/bin/security output.
// If the secret is ASCII, output looks like:
//
//     password: "secret"
//
// If the secret is non-ASCII, output looks like:
//
//     password: 0x74C3AB73745F73C3A96372C3A974  "t\303\253st_s\303\251cr\303\251t"
//
// where the first field is 0x + hex-encoded secret.
func parseSecret(s string) string {
	i := strings.Index(s, "password: ")
	if i < 0 {
		return ""
	}
	s = s[10:]
	if strings.HasPrefix(s, `"`) {
		return s[1 : len(s)-1]
	}
	if strings.HasPrefix(s, "0x") {
		i = strings.Index(s, " ")
		if i < 0 {
			log.Println("error: parse output")
			return ""
		}
		s = s[2:i]
		data, err := hex.DecodeString(s)
		if err != nil {
			log.Printf("error: decode secret: %v", err)
			return ""
		}
		return string(data)
	}

	return ""
}
