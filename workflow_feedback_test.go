// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package aw

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	data, err = json.Marshal(it)
	assert.Nil(t, err, "marshal Item failed")
	js := string(data)
	assert.Equal(t, x, js, "unexpected Warning item")

	it = wf.NewFileItem("/Volumes")
	x = `{"title":"Volumes","subtitle":"/Volumes","autocomplete":"Volumes","arg":"/Volumes","uid":"/Volumes","valid":true,"type":"file","icon":{"path":"/Volumes","type":"fileicon"}}`
	data, err = json.Marshal(it)
	assert.Nil(t, err, "marshal Item failed")
	js = string(data)
	assert.Equal(t, x, js, "unexpected File item")
}

// TestNewFileItem verifies Item creation by Workflow.NewFileItem().
func TestNewFileItem(t *testing.T) {
	t.Parallel()

	var (
		wf      = New()
		ipPath  = filepath.Join(wf.Dir(), "info.plist")
		ipShort = strings.ReplaceAll(ipPath, os.Getenv("HOME"), "~")

		it = wf.NewFileItem(ipPath)
	)

	assert.Equal(t, "info.plist", it.title, "unexpected title")
	assert.Equal(t, ipShort, *it.subtitle, "unexpected subtitle")
	assert.Equal(t, ipPath, *it.uid, "unexpected UID")
	assert.True(t, it.file, "unexpected file")
	assert.Equal(t, IconType("fileicon"), it.icon.Type, "unexpected value type")
	assert.Equal(t, ipPath, it.icon.Value, "unexpected icon value")
}

// TestWarnEmpty verifies Item creation by Workflow.WarnEmpty().
func TestWarnEmpty(t *testing.T) {
	wf := New()
	assert.Equal(t, 0, len(wf.Feedback.Items), "feedback not empty")
	wf.WarnEmpty("test", "test")
	assert.Equal(t, 1, len(wf.Feedback.Items), "feedback empty")
}
