// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/deanishe/awgo/util"
)

var (
	// UpdateInterval is how often to check for updates.
	UpdateInterval = 24 * time.Hour
	// HTTPTimeout is the timeout for establishing an HTTP(S) connection.
	HTTPTimeout = 60 * time.Second

	// HTTP client used to talk to APIs
	client *http.Client
)

// Mockable functions
var (
	// Run command
	runCommand = func(name string, arg ...string) error {
		return exec.Command(name, arg...).Run()
	}
	// save a URL to a filepath.
	download = func(URL, path string) error {
		res, err := openURL(URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		util.MustExist(filepath.Dir(path))
		out, err := os.Create(path)
		if err != nil {
			return err
		}
		defer out.Close()
		n, err := io.Copy(out, res.Body)
		if err != nil {
			return err
		}
		log.Printf("wrote %q (%d bytes)", util.PrettyPath(path), n)
		return nil
	}
)

// Source provides workflow files that can be downloaded.
// This is what concrete updaters (e.g. GitHub, Gitea) should implement.
// Source is called by the Updater after every updater interval.
type Source interface {
	// Downloads returns all available workflow files.
	Downloads() ([]Download, error)
}

// byVersion sorts downloads by version.
type byVersion []Download

// Len implements sort.Interface.
func (s byVersion) Len() int      { return len(s) }
func (s byVersion) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byVersion) Less(i, j int) bool {
	// Compare workflow versions first, compatible Alfred version second.
	if s[i].Version.Ne(s[j].Version) {
		return s[i].Version.Lt(s[j].Version)
	}
	return s[i].AlfredVersion().Lt(s[j].AlfredVersion())
}

// Download is an Alfred workflow available for download & installation.
// It is the primary update data structure, returned by all Sources.
type Download struct {
	URL string // Where the workflow file can be downloaded from
	// Filename for downloaded file.
	// Must have extension .alfredworkflow or .alfredXworkflow where X is a number,
	// otherwise the Download will be ignored.
	Filename   string
	Version    SemVer // Semantic version no.
	Prerelease bool   // Whether this version is a pre-release
}

// AlfredVersion returns minimum compatible version of Alfred based on file extension.
// For example, Workflow.alfred4workflow has version 4, while
// Workflow.alfred3workflow has version 3.
// The standard .alfredworkflow extension returns a zero version.
func (dl Download) AlfredVersion() SemVer {
	m := rxWorkflowFile.FindStringSubmatch(dl.Filename)
	if len(m) == 2 {
		if v, err := NewSemVer(m[1]); err == nil {
			return v
		}
	}
	return SemVer{}
}

// Updater checks for newer version of the workflow. Available versions are
// provided by a Source, such as the built-in GitHub source, which
// reads the releases in a GitHub repo. It is a concrete implementation
// of aw.Updater.
//
// CheckForUpdate() retrieves the list of available downloads from the
// source and caches them. UpdateAvailable() reads the cache and returns
// true if there is a download with a higher version than the running workflow.
// Install() downloads the latest version and asks Alfred to install it.
//
// Because downloading releases is slow and workflows need to run fast,
// you should not run CheckForUpdate() in a Script Filter.
//
// If an Updater is set on a Workflow struct, a magic action will be set for
// updates, so you can just add an Item that autocompletes to the update
// magic argument ("workflow:update" by default), and AwGo will check for an
// update and install it if available.
//
// See ../examples/update for a full example implementation of updates.
type Updater struct {
	Source         Source // Provides downloads
	CurrentVersion SemVer // Version of the installed workflow
	Prereleases    bool   // Include pre-releases when checking for updates

	// AlfredVersion is the version of the running Alfred application.
	// Read from $alfred_version environment variable.
	AlfredVersion SemVer

	// When the remote release list was last checked (and possibly cached)
	LastCheck      time.Time
	updateInterval time.Duration // How often to check for an update
	downloads      []Download    // Available workflow files

	// Cache paths
	cacheDir      string // Directory to store cache files in
	pathLastCheck string // Cache path for check time
	pathDownloads string // Cache path for available downloads
}

// NewUpdater creates a new Updater for Source. `currentVersion` is the workflow's
// version number and `cacheDir` is a directory where the Updater can cache
// a list of available releases.
func NewUpdater(src Source, currentVersion, cacheDir string) (*Updater, error) {
	v, err := NewSemVer(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q: %w", currentVersion, err)
	}
	if cacheDir == "" {
		return nil, errors.New("empty cacheDir")
	}

	u := &Updater{
		CurrentVersion: v,
		LastCheck:      time.Time{},
		Source:         src,
		cacheDir:       cacheDir,
		updateInterval: UpdateInterval,
		pathLastCheck:  filepath.Join(cacheDir, "LastCheckTime.txt"),
		pathDownloads:  filepath.Join(cacheDir, "Downloads.json"),
	}

	if s := os.Getenv("alfred_version"); s != "" {
		if v, err := NewSemVer(s); err == nil {
			u.AlfredVersion = v
		}
	}

	// Load LastCheck
	if data, err := ioutil.ReadFile(u.pathLastCheck); err == nil {
		t, err := time.Parse(time.RFC3339, string(data))
		if err != nil {
			log.Printf("error: load last update check: %v", err)
		} else {
			u.LastCheck = t
		}
	}
	return u, nil
}

// UpdateAvailable returns true if an update is available. Retrieves
// the list of releases from the cache written by CheckForUpdate.
func (u *Updater) UpdateAvailable() bool {
	dl := u.latest()
	if dl == nil {
		log.Println("no downloads available")
		return false
	}
	log.Printf("latest version: %v", dl.Version)
	return dl.Version.Gt(u.CurrentVersion)
}

// CheckDue returns true if the time since the last check is greater than
// Updater.UpdateInterval.
func (u *Updater) CheckDue() bool {
	if u.LastCheck.IsZero() {
		// log.Println("never checked for updates")
		return true
	}
	elapsed := time.Since(u.LastCheck)
	log.Printf("%s since last check for update", elapsed)
	return elapsed > u.updateInterval
}

// CheckForUpdate fetches the list of releases from remote (via Releaser)
// and caches it locally.
func (u *Updater) CheckForUpdate() error {
	// If update fails, don't try again for at least an hour
	u.LastCheck = time.Now().Add(-u.updateInterval).Add(time.Hour)
	defer u.cacheLastCheck()

	var (
		dls  []Download
		data []byte
		err  error
	)

	if dls, err = u.Source.Downloads(); err != nil {
		return err
	}
	u.downloads = dls
	if data, err = json.Marshal(dls); err != nil {
		return err
	}

	u.clearCache()
	if err := ioutil.WriteFile(u.pathDownloads, data, 0600); err != nil {
		return err
	}
	u.LastCheck = time.Now()
	return nil
}

// Install downloads and installs the latest available version.
// After the workflow file is downloaded, Install calls Alfred to
// install the update.
func (u *Updater) Install() error {
	dl := u.latest()
	if dl == nil {
		return errors.New("no downloads available")
	}
	log.Printf("downloading version %s ...", dl.Version)
	p := filepath.Join(u.cacheDir, dl.Filename)
	if err := download(dl.URL, p); err != nil {
		return err
	}

	return runCommand("open", p)
}

// clearCache removes the update cache.
func (u *Updater) clearCache() {
	if err := util.ClearDirectory(u.cacheDir); err != nil {
		log.Printf("error: clear cache: %v", err)
	}
	util.MustExist(u.cacheDir)
}

// cacheLastCheck saves time to cache.
func (u *Updater) cacheLastCheck() {
	data, err := u.LastCheck.MarshalText()
	if err != nil {
		log.Printf("error: marshal time: %s", err)
		return
	}
	if err := ioutil.WriteFile(u.pathLastCheck, data, 0600); err != nil {
		log.Printf("error: cache update time: %s", err)
	}
}

// Returns latest version that is compatible with the Updater's
// Alfred version & pre-release preference.
func (u *Updater) latest() *Download {
	if u.downloads == nil {
		u.downloads = []Download{}
		if !util.PathExists(u.pathDownloads) {
			log.Println("no cached releases")
			return nil
		}
		// Load from cache
		data, err := ioutil.ReadFile(u.pathDownloads)
		if err != nil {
			log.Printf("error: read cached releases: %s", err)
			return nil
		}
		if err := json.Unmarshal(data, &u.downloads); err != nil {
			log.Printf("error: unmarshal cached releases: %s", err)
			return nil
		}
		sort.Sort(sort.Reverse(byVersion(u.downloads)))
	}
	if len(u.downloads) == 0 {
		return nil
	}
	for _, dl := range u.downloads {
		dl := dl
		if dl.Prerelease && !u.Prereleases {
			continue
		}
		if !u.AlfredVersion.IsZero() && dl.AlfredVersion().Gt(u.AlfredVersion) {
			log.Printf("incompatible: %q: current=%v, required=%v", dl.Filename, u.AlfredVersion, dl.AlfredVersion())
			continue
		}
		return &dl
	}
	return nil
}

// // Mockable function to run commands
// type commandRunner func(name string, arg ...string) error
//
// // Run command via exec.Command
// func runCommand(name string, arg ...string) error {
// 	return exec.Command(name, arg...).Run()
// }

// makeHTTPClient returns an http.Client with a sensible configuration.
func makeHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   HTTPTimeout,
				KeepAlive: HTTPTimeout,
			}).Dial,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}
}

// getURL returns the contents of a URL.
func getURL(url string) ([]byte, error) {
	res, err := openURL(url)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// openURL returns an http.Response. It will return an error if the
// HTTP status code > 299.
func openURL(url string) (*http.Response, error) {
	log.Printf("fetching %s ...", url)
	if client == nil {
		client = makeHTTPClient()
	}
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	log.Printf("[%d] %s", r.StatusCode, url)
	if r.StatusCode > 299 {
		r.Body.Close()
		return nil, errors.New(r.Status)
	}
	return r, nil
}
