// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package update

import (
	"fmt"
	"io/ioutil"
	"testing"

	aw "github.com/deanishe/awgo"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		jsname   string
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
		// returns error
		{
			"metadata-empty.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-alfred4.json",
			true,
			"Secure-SHell-0.8.0.alfred4workflow",
			"https://github.com/deanishe/alfred-ssh/releases/download/v0.8.0/Secure-SHell-0.8.0.alfred4workflow",
			mustVersion("v0.8"),
		},
		// returns error
		{
			"metadata-bad-filename.json",
			false,
			"",
			"",
			SemVer{},
		},
		{
			"metadata-old.json",
			true,
			"Secure-SHell-0.1.0.alfredworkflow",
			"https://github.com/deanishe/alfred-ssh/releases/download/v0.1.0/Secure-SHell-0.1.0.alfredworkflow",
			mustVersion("v0.1"),
		},
	}

	for i, td := range tests {
		var (
			data []byte
			err  error
			dl   Download
			src  *metadataSource
		)
		if data, err = ioutil.ReadFile("testdata/" + td.jsname); err != nil {
			t.Errorf("[%d] parse metadata: %v", i, err)
			continue
		}
		dl, err = parseMetadata(data)
		if !td.ok {
			if err == nil {
				t.Errorf("[%d] bad release parsed", i)
			}
			continue
		}
		if err != nil {
			t.Errorf("[%d] parse metadata: %v", i, err)
			continue
		}

		if dl.Filename != td.filename {
			t.Errorf("[%d] Bad Filename. Expected=%q, Got=%q", i, td.filename, dl.Filename)
		}
		if dl.URL != td.url {
			t.Errorf("[%d] Bad URL. Expected=%q, Got=%q", i, td.url, dl.URL)
		}
		if dl.Version.Ne(td.version) {
			t.Errorf("[%d] Bad Version. Expected=%v, Got=%v", i, td.version, dl.Version)
		}
		if dl.Prerelease != false {
			t.Errorf("[%d] Bad Prerelease. Expected=false, Got=%v", i, dl.Prerelease)
		}

		src = &metadataSource{
			url:   "https://raw.githubusercontent.com/deanishe/alfred-ssh/master/metadata.json",
			fetch: func(URL string) ([]byte, error) { return data, nil },
		}

		dls, err := src.Downloads()
		if err != nil {
			t.Fatal("parse empty JSON")
		}

		if len(dls) != 1 {
			t.Errorf("Bad Count. Expected=1, Got=%d", len(dls))
		}
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
