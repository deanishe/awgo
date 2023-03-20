// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/deanishe/awgo/util"

	"howett.net/plist"
)

// extract major version number from version string
var rxVersion = regexp.MustCompile(`^\d+`)

var (
	// Alfred's standard preferences folder, which is where the preferences bundle
	// is stored when the user isn't syncing their settings between machines
	defaultSyncDirV3 = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3")
	defaultSyncDirV4 = os.ExpandEnv("${HOME}/Library/Application Support/Alfred")
)

// Option configures Info created by New.
type Option func(info *Info)

// LibDir tells New to search a specific directory for Alfred config files.
// Default is ~/Library.
func LibDir(dir string) Option {
	return func(info *Info) {
		info.dir = dir
	}
}

// InfoPlist tells New to parse a specific info.plist file. Default is ./info.plist.
func InfoPlist(path string) Option {
	return func(info *Info) {
		info.ipPath = path
	}
}

// Info contains information about a workflow and Alfred.
//
// The information is extracted from environment variables,
// the workflow's info.plist, Alfred's own configuration files
// and finally, some defaults.
type Info struct {
	// Workflow info read from environment variables/info.plist
	Name     string // Workflow name
	Version  string // Workflow version
	BundleID string // Workflow bundle ID

	// Workflow directories
	CacheDir string // Workflow cache directory
	DataDir  string // Workflow data directory
	// Where workflow should be installed. This is
	// Alfred's workflow directory (AlfredWorkflowDir)
	// plus workflow's bundle ID.
	InstallDir string

	// Alfred info
	AlfredMajorVersion int    // Alfred's major version number
	AlfredSyncDir      string // Path of Alfred's syncfolder
	AlfredPrefsBundle  string // Path to the Alfred.alfredpreferences bundle
	AlfredWorkflowDir  string // Directory workflows are stored in
	AlfredCacheDir     string // Root directory for all workflow cache data
	AlfredDataDir      string // Root directory for all persistent workflow data

	// Directory searched for preferences files.
	// Default is ~/Library.
	dir string
	// Path to workflow's info.plist.
	// Default is ./info.plist
	ipPath string
}

// NewInfo creates a new Info. Workflow info is read from Alfred environment
// variables (if set), and from info.plist in the working directory and
// Alfred's configuration files. These paths may be changed using the
// the LibDir and InfoPlist Options. Settings from info.plist take priority
// over those from environment variables.
//
// It returns an error if info.plist or the configuration files cannot be found.
func NewInfo(option ...Option) (*Info, error) {
	info := &Info{
		dir:    os.ExpandEnv("${HOME}/Library"),
		ipPath: "info.plist",
	}
	for _, opt := range option {
		opt(info)
	}
	info.readEnv()
	if err := info.readPlist(); err != nil {
		return nil, err
	}
	if err := info.findAlfredVersion(); err != nil {
		return nil, err
	}
	if err := info.findFolders(); err != nil {
		return nil, err
	}
	return info, nil
}

// Env returns an Alfred-like environment.
func (info *Info) Env() map[string]string {
	env := map[string]string{
		"alfred_workflow_name":     info.Name,
		"alfred_workflow_version":  info.Version,
		"alfred_workflow_bundleid": info.BundleID,
		"alfred_workflow_uid":      info.BundleID,
		"alfred_workflow_cache":    info.CacheDir,
		"alfred_workflow_data":     info.DataDir,
		"alfred_preferences":       info.AlfredPrefsBundle,
		"alfred_version":           fmt.Sprintf("%d", info.AlfredMajorVersion),
		"alfred_debug":             "1",
	}
	return env
}

func (info *Info) findFolders() error {
	syncDir, err := findSyncFolder(info.AlfredMajorVersion, info.dir)
	if err != nil {
		return err
	}
	info.AlfredSyncDir = syncDir
	info.AlfredPrefsBundle = filepath.Join(syncDir, "Alfred.alfredpreferences")
	info.AlfredWorkflowDir = filepath.Join(syncDir, "Alfred.alfredpreferences/workflows")
	info.InstallDir = filepath.Join(info.AlfredWorkflowDir, info.BundleID)
	if info.AlfredCacheDir == "" {
		switch info.AlfredMajorVersion {
		case 3:
			info.AlfredCacheDir = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred-3/Workflow Data")
		default:
			info.AlfredCacheDir = os.ExpandEnv("${HOME}/Library/Caches/com.runningwithcrayons.Alfred/Workflow Data")
		}
	}
	if info.AlfredDataDir == "" {
		switch info.AlfredMajorVersion {
		case 3:
			info.AlfredDataDir = os.ExpandEnv("${HOME}/Library/Application Support/Alfred 3/Workflow Data")
		default:
			info.AlfredDataDir = os.ExpandEnv("${HOME}/Library/Application Support/Alfred/Workflow Data")
		}
	}
	if info.CacheDir == "" {
		info.CacheDir = filepath.Join(info.AlfredCacheDir, info.BundleID)
	}
	if info.DataDir == "" {
		info.DataDir = filepath.Join(info.AlfredDataDir, info.BundleID)
	}

	return nil
}

func (info *Info) findAlfredVersion() error {
	if info.AlfredMajorVersion != 0 {
		return nil
	}
	if util.PathExists(filepath.Join(info.dir, "Application Support/Alfred/prefs.json")) {
		info.AlfredMajorVersion = 4
		return nil
	}
	if util.PathExists(filepath.Join(info.dir, "Preferences/com.runningwithcrayons.Alfred-Preferences-3.plist")) {
		info.AlfredMajorVersion = 3
		return nil
	}
	return errors.New("Alfred version not found")
}

func (info *Info) readEnv() {
	info.Name = os.Getenv("alfred_workflow_name")
	info.BundleID = os.Getenv("alfred_workflow_bundleid")
	info.Version = os.Getenv("alfred_workflow_version")
	if s := os.Getenv("alfred_workflow_data"); s != "" {
		info.DataDir = s
		info.AlfredDataDir = filepath.Dir(s)
	}
	if s := os.Getenv("alfred_workflow_cache"); s != "" {
		info.CacheDir = s
		info.AlfredCacheDir = filepath.Dir(s)
	}

	if s := rxVersion.FindString(os.Getenv("alfred_version")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			info.AlfredMajorVersion = n
		}
	}
}

// readPlist reads workflow information from info.plist.
func (info *Info) readPlist() error {
	file, err := os.Open(info.ipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	p := struct {
		Name     string `plist:"name"`
		Version  string `plist:"version"`
		BundleID string `plist:"bundleid"`
	}{}
	if _, err = plist.Unmarshal(data, &p); err != nil {
		return err
	}
	if p.Name != "" {
		info.Name = p.Name
	}
	if p.Version != "" {
		info.Version = p.Version
	}
	if p.BundleID != "" {
		info.BundleID = p.BundleID
	}
	return nil
}

// expand ~ in a filepath.
func expand(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		return filepath.Clean(filepath.Join(os.ExpandEnv("$HOME"), path[1:]))
	}
	return path
}

// get path to Alfred's sync folder (parent of Alfred.alfredpreferences) from
// environment or Alfred's config files
func findSyncFolder(v int, dir string) (string, error) {
	if s := os.Getenv("alfred_preferences"); s != "" {
		return filepath.Dir(s), nil
	}

	var (
		// Alfred 4+ has a dedicated prefs.json file, but earlier versions store
		// the setting in Alfred Preference's version-specific prefs file
		prefsJSON  = filepath.Join(dir, "Application Support/Alfred/prefs.json")
		prefsPlist = filepath.Join(dir, "Preferences/com.runningwithcrayons.Alfred-Preferences-3.plist")
		err        error
	)

	// Look for Alfred 4+ prefs.json
	if util.PathExists(prefsJSON) && v != 3 {
		var (
			prefs = struct {
				Current  string            `json:"current"`
				Versions map[string]string `json:"syncfolders"`
			}{}
			data []byte
		)
		if data, err = ioutil.ReadFile(prefsJSON); err != nil {
			return "", err
		}
		if err = json.Unmarshal(data, &prefs); err != nil {
			return "", err
		}

		return filepath.Dir(prefs.Current), nil
	}

	// Look for Alfred 3 preferences plist
	if util.PathExists(prefsPlist) {
		var (
			prefs = struct {
				SyncDir string `plist:"syncfolder"`
			}{}

			data []byte
		)
		if data, err = ioutil.ReadFile(prefsPlist); err != nil {
			return "", err
		}
		if _, err = plist.Unmarshal(data, &prefs); err != nil {
			return "", err
		}

		p := expand(prefs.SyncDir)
		if util.PathExists(p) {
			return p, nil
		}
		return defaultSyncDirV3, nil
	}

	return "", errors.New("Alfred preferences not found")
}
