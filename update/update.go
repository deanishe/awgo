//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-03
//

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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/deanishe/awgo/util"
)

// DefaultUpdateInterval is how often to check for updates.
const DefaultUpdateInterval = time.Duration(24 * time.Hour)

// HTTPTimeout is the timeout for establishing an HTTP(S) connection.
var HTTPTimeout = (60 * time.Second)

// Versioned has a semantic version number (for comparing to releases)
// and a cache directory (for saving information about available versions
// and time of last update check).
//
// aw.Workflow implements this interface.
type Versioned interface {
	Version() string  // Returns a semantic version string
	CacheDir() string // Path to directory to store cache files
}

// Releaser is what concrete updaters should implement.
// The Updater should call the Releaser after every update interval
// to check if an update is available.
type Releaser interface {
	Releases() ([]*Release, error)
}

// Releases is a slice of Releases that implements sort.Interface
type Releases []*Release

// Len implements sort.Interface
func (r Releases) Len() int { return len(r) }

// Less implements sort.Interface
func (r Releases) Less(i, j int) bool { return r[i].Version.LT(r[j].Version) }

// Swap implements sort.Interface
func (r Releases) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// Release is the metadata of a release. Each Releaser must return one
// or more Release structs. If one has a higher version number than
// the workflow's current version, it will be considered an update.
type Release struct {
	Filename   string   // Filename of workflow file
	URL        *url.URL // URL of the .alfredworkflow (or .alfred3workflow) file
	Prerelease bool     // Whether this release is a pre-release
	Version    SemVer   // The version number of the release
}

// SortReleases sorts a slice of Releases, lowest to highest version number.
func SortReleases(releases []*Release) {
	sort.Sort(Releases(releases))
}

// Updater checks for newer version of the workflow. Available versions are
// provided by a Releaser, such as the built-in GitHub releaser, which
// reads the releases in a GitHub repo.
//
// CheckForUpdate() retrieves the list of available releases from the
// releaser and caches them. UpdateAvailable() reads the cache. Install()
// downloads the latest version and asks Alfred to install it.
//
// LastCheck and available releases are cached.
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
	CurrentVersion SemVer        // Version of the installed workflow
	LastCheck      time.Time     // When the remote release list was last checked
	Prereleases    bool          // Include pre-releases when checking for updates
	Releaser       Releaser      // Provides available versions
	updateInterval time.Duration // How often to check for an update
	cacheDir       string        // Directory to store cache files in
	pathLastCheck  string        // Cache path for check time
	pathReleases   string        // Cache path for available releases
	releases       []*Release    // Available releases
}

// New creates a new Updater for Versioned and Releaser.
//
// CurrentVersion is set to the workflow's version by calling Version().
// If you've created your own Workflow struct and subsequently called
// wf.SetVersion(), you'll also need to set CurrentVersion manually.
//
// LastCheck is loaded from the cache, and UpdateInterval is set to
// DefaultUpdateInterval.
func New(v Versioned, r Releaser) (*Updater, error) {
	semver, err := NewSemVer(v.Version())
	if err != nil {
		return nil, fmt.Errorf("invalid semantic version (%s): %v", v.Version(), err)
	}

	u := &Updater{
		CurrentVersion: semver,
		LastCheck:      time.Time{},
		Releaser:       r,
		cacheDir:       v.CacheDir(),
		updateInterval: DefaultUpdateInterval,
	}
	u.pathLastCheck = u.cachePath("LastCheckTime")
	u.pathReleases = u.cachePath("Releases.json")

	// Load LastCheck
	data, err := ioutil.ReadFile(u.pathLastCheck)
	if err == nil {
		t, err := time.Parse(time.RFC3339, string(data))
		if err != nil {
			log.Printf("Failed to load last update check time from disk: %s", err)
		} else {
			u.LastCheck = t
		}
	}
	return u, nil
}

// UpdateInterval sets the interval between checks for new versions.
func (u *Updater) UpdateInterval(interval time.Duration) { u.updateInterval = interval }

// UpdateAvailable returns true if an update is available. Retrieves
// the list of releases from the cache written by CheckForUpdate.
func (u *Updater) UpdateAvailable() bool {
	r := u.latest()
	if r == nil {
		log.Println("No releases available.")
		return false
	}
	log.Printf("Latest release: %s", r.Version.String())
	return r.Version.GT(u.CurrentVersion)
}

// CheckDue returns true if the time since the last check is greater than
// Updater.UpdateInterval.
func (u *Updater) CheckDue() bool {
	if u.LastCheck.IsZero() {
		log.Println("Never checked for updates")
		return true
	}
	elapsed := time.Now().Sub(u.LastCheck)
	log.Printf("%s since last check for update", util.HumanDuration(elapsed))
	return elapsed.Nanoseconds() > u.updateInterval.Nanoseconds()
}

// CheckForUpdate fetches the list of releases from remote (via Releaser)
// and caches it locally.
func (u *Updater) CheckForUpdate() error {
	u.clearCache()
	rels, err := u.Releaser.Releases()
	if err != nil {
		return err
	}
	u.releases = rels
	data, err := json.Marshal(u.releases)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(u.pathReleases, data, 0600); err != nil {
		return err
	}
	u.LastCheck = time.Now()
	u.cacheLastCheck()
	return nil
}

// Install downloads and installs the latest available version.
// After the workflow file is downloaded, Install calls Alfred to
// install the update.
func (u *Updater) Install() error {
	r := u.latest()
	if r == nil {
		return errors.New("no releases available")
	}
	log.Printf("downloading release %s ...", r.Version.String())
	p := u.cachePath(r.Filename)
	if err := download(r.URL, p); err != nil {
		return err
	}
	cmd := exec.Command("open", "-a", "Alfred 3", p)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// cachePath returns a filepath within AwGo's update cache directory.
func (u *Updater) cachePath(filename string) string {
	dp := util.MustExist(filepath.Join(u.cacheDir, "_aw/update"))
	return filepath.Join(dp, filename)
}

// clearCache removes the update cache.
func (u *Updater) clearCache() { util.ClearDirectory(u.cachePath("")) }

// cacheLastCheck saves time to cachepath.
func (u *Updater) cacheLastCheck() {
	data, err := u.LastCheck.MarshalText()
	if err != nil {
		log.Printf("Error marshalling time: %s", err)
		return
	}
	if err := ioutil.WriteFile(u.pathLastCheck, data, 0600); err != nil {
		log.Printf("Failed to cache update time: %s", err)
	}
}

// latest returns the release with the highest version number. Data is
// loaded from the local cache of releases. Call CheckUpdate() to update
// the cache.
func (u *Updater) latest() *Release {
	if u.releases == nil {
		if !util.PathExists(u.pathReleases) {
			log.Println("No cached releases.")
			return nil
		}
		// Load from cache
		data, err := ioutil.ReadFile(u.pathReleases)
		if err != nil {
			log.Printf("Error reading cached releases: %s", err)
			return nil
		}
		if err := json.Unmarshal(data, &u.releases); err != nil {
			log.Printf("Error unmarshalling cached releases: %s", err)
			return nil
		}
	}
	// log.Printf("%d releases available.", len(u.releases))
	if len(u.releases) == 0 {
		return nil
	}

	SortReleases(u.releases)

	if u.Prereleases {
		return u.releases[len(u.releases)-1]
	}

	// Find newest non-pre-release version
	i := len(u.releases) - 1
	for i > 0 {
		if !u.releases[i].Prerelease {
			break
		}
		i--
	}

	r := u.releases[i]
	if r.Prerelease {
		return nil
	}
	return r
}

// makeHTTPClient returns an http.Client with a sensible configuration.
func makeHTTPClient() http.Client {
	return http.Client{
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
func getURL(u *url.URL) ([]byte, error) {
	res, err := openURL(u)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	return data, err
}

// download saves a URL to a filepath.
func download(u *url.URL, path string) error {
	res, err := openURL(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	n, err := io.Copy(out, res.Body)
	if err != nil {
		return err
	}
	log.Printf("Wrote `%s` (%d bytes)", path, n)
	return nil
}

// openURL returns an http.Response. It will return an error if the
// HTTP status code > 299.
func openURL(u *url.URL) (*http.Response, error) {
	log.Printf("Fetching %s ...", u.String())
	client := makeHTTPClient()
	res, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	log.Printf("[%d] %s", res.StatusCode, u.String())
	if res.StatusCode > 299 {
		res.Body.Close()
		return nil, errors.New(res.Status)
	}
	return res, nil
}
