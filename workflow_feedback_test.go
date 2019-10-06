// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestItemHelpers(t *testing.T) {
	t.Parallel()

	var (
		wf   = New()
		data []byte
		err  error
	)

	it := wf.NewWarningItem("Warn Title", "Warn subtitle")
	x := `{"title":"Warn Title","subtitle":"Warn subtitle","valid":false,"icon":{"path":"/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"}}`
	if data, err = json.Marshal(it); err != nil {
		t.Fatalf("marshal Item: %v", err)
	}
	js := string(data)

	if js != x {
		t.Errorf("Bad Warning Item. Expected=%v, Got=%v", x, js)
	}

	it = wf.NewFileItem("/Volumes")
	x = `{"title":"Volumes","subtitle":"/Volumes","autocomplete":"Volumes","arg":"/Volumes","uid":"/Volumes","valid":true,"type":"file","icon":{"path":"/Volumes","type":"fileicon"}}`
	if data, err = json.Marshal(it); err != nil {
		t.Fatalf("marshal Item: %v", err)
	}
	js = string(data)

	if js != x {
		t.Errorf("Bad File Item. Expected=%v, Got=%v", x, js)
	}
}

func TestNewFileItem(t *testing.T) {
	t.Parallel()

	var (
		wf      = New()
		ipPath  = filepath.Join(wf.Dir(), "info.plist")
		ipShort = strings.Replace(ipPath, os.Getenv("HOME"), "~", -1)

		it = wf.NewFileItem(ipPath)
	)

	if it.title != "info.plist" {
		t.Fatalf("Incorrect title: %v", it.title)
	}

	if *it.subtitle != ipShort {
		t.Fatalf("Incorrect subtitle: %v", *it.subtitle)
	}

	if *it.uid != ipPath {
		t.Fatalf("Incorrect UID: %v", *it.uid)
	}

	if it.file != true {
		t.Fatalf("Incorrect file: %v", it.file)
	}

	if it.icon.Type != "fileicon" {
		t.Fatalf("Incorrect type: %v", it.icon.Type)
	}

	if it.icon.Value != ipPath {
		t.Fatalf("Incorrect Value: %v", it.icon.Value)
	}
}

// WarnEmpty adds an item
func TestWarnEmpty(t *testing.T) {
	wf := New()
	assert.Equal(t, 0, len(wf.Feedback.Items), "feedback not empty")
	wf.WarnEmpty("test", "test")
	assert.Equal(t, 1, len(wf.Feedback.Items), "feedback empty")
}
