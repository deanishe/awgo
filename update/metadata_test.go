// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	aw "github.com/deanishe/awgo"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		jsName   string
		ok       bool
		filename string
		url      string
		version  SemVer
	}{
		{
			"metadata-valid.json",
			true,
			"Secure-SHell-0.8.0.alfredworkflow",
			"https://github.com/deanishe/alfred-ssh/releases/download/v0.8.0/Secure-SHell-0.8.0.alfredworkflow",
			mustVersion("v0.8"),
		},
		{
			"metadata-alfred4.json",
			true,
			"Secure-SHell-0.8.0.alfred4workflow",
			"https://github.com/deanishe/alfred-ssh/releases/download/v0.8.0/Secure-SHell-0.8.0.alfred4workflow",
			mustVersion("v0.8"),
		},
		{
			"metadata-old.json",
			true,
			"Secure-SHell-0.1.0.alfredworkflow",
			"https://github.com/deanishe/alfred-ssh/releases/download/v0.1.0/Secure-SHell-0.1.0.alfredworkflow",
			mustVersion("v0.1"),
		},

		// invalid metadata files
		{
			"invalid.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-empty.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-empty-url.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-invalid-url.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-invalid-url-scheme.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-invalid-version.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-bad-filename.json",
			false,
			"",
			"",
			SemVer{},
		},
	}

	for _, td := range tests {
		td := td
		t.Run(td.jsName, func(t *testing.T) {
			t.Parallel()

			data := mustRead("testdata/" + td.jsName)
			dl, err := parseMetadata(data)
			if !td.ok {
				if err == nil {
					t.Fatal("bad release accepted")
				}
				return
			}

			if err != nil {
				t.Fatalf("parse metadata: %v", err)
			}

			assert.Equal(t, td.filename, dl.Filename, "Bad filename")
			assert.Equal(t, td.url, dl.URL, "Bad URL")
			assert.True(t, dl.Version.Eq(td.version), "Bad version")
			assert.False(t, dl.Prerelease, "Prerelease is true")

			src := &metadataSource{
				url:   "https://raw.githubusercontent.com/deanishe/alfred-ssh/master/metadata.json",
				fetch: func(URL string) ([]byte, error) { return data, nil },
			}

			dls, err := src.Downloads()
			if err != nil {
				t.Fatal("parse empty JSON")
			}
			assert.Equal(t, 1, len(dls), "Bad download count")
		})
	}
}

func TestMetadataSource_Downloads(t *testing.T) {
	// fetch fails
	fetch := func(URL string) ([]byte, error) { return nil, errors.New("i ded") }
	src := &metadataSource{url: "blah", fetch: fetch}
	if _, err := src.Downloads(); err == nil {
		t.Fatal("bad fetch didn't fail")
	}

	// fetch returns invalid data
	fetch = func(URL string) ([]byte, error) { return []byte("totes not JSON"), nil }
	src = &metadataSource{url: "blah", fetch: fetch}
	if _, err := src.Downloads(); err == nil {
		t.Fatal("bad fetch didn't fail")
	}
}

// Configure Workflow to update from a remote `metadata.json` file.
func ExampleMetadata() {
	// Set source repo using Gitea Option
	wf := aw.New(Metadata("https://raw.githubusercontent.com/deanishe/alfred-ssh/master/metadata.json"))
	// Is a check for a newer version due?
	fmt.Println(wf.UpdateCheckDue())
	// Output:
	// true
}
