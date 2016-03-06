package workflow

// TODO: Replace calls to log.Fatalf() with Workflow.SendError()

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/mkrautz/plist"

	"gogs.deanishe.net/deanishe/awgo/util"
)

const (
	Version = 0.1
)

// The workflow object operated on by top-level functions.
var defaultWorkflow Workflow

// Info contains some of the information extracted from info.plist.
type Info struct {
	BundleId    string `plist:"bundleid"`
	Author      string `plist:"createdby"`
	Description string `plist:"description"`
	Name        string `plist:"name"`
	Readme      string `plist:"readme"`
	Website     string `plist:"webaddress"`
}

// Workflow provides an API for Alfred.
type Workflow struct {
	Feedback   Feedback
	Info       Info
	infoLoaded bool
	Env        map[string]string

	bundleId    string
	dataDir     string
	cacheDir    string
	workflowDir string
}

// readInfoPlist loads the data in `info.plist`
func (wf *Workflow) readInfoPlist() error {
	if wf.infoLoaded {
		return nil
	}
	p := path.Join(wf.GetWorkflowDir(), "info.plist")
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("Couldn't open `info.plist` (%s) :  %v", p, err)
	}
	err = plist.Unmarshal(buf, &wf.Info)
	if err != nil {
		return fmt.Errorf("Error parsing `info.plist` (%s) : %v", p, err)
	}
	log.Printf("info=%v", wf.Info)
	wf.bundleId = wf.Info.BundleId
	wf.infoLoaded = true
	return nil
}

// loadEnv reads Alfred's variables from the environment.
func (wf *Workflow) loadEnv() {
	wf.Env = make(map[string]string)
	// Variables currently exported by Alfred. These actual names
	// are prefixed with `alfred_`.
	keys := []string{
		"version",
		"version_build",
		"theme",
		"theme_background",
		"theme_subtext",
		"preferences",
		"preferences_localhash",
		"workflow_cache",
		"workflow_data",
		"workflow_name",
		"workflow_uid",
		"workflow_bundleid",
	}
	var (
		val    string
		envkey string
	)
	for _, key := range keys {
		envkey = fmt.Sprintf("alfred_%s", key)
		val = os.Getenv(envkey)
		wf.Env[key] = val
		// Some special keys
		if key == "workflow_cache" {
			wf.cacheDir = val
		} else if key == "workflow_data" {
			wf.dataDir = val
		} else if key == "workflow_bundleid" {
			wf.bundleId = val
		}
	}
}

// initializeLogging ensures future log messages are written to workflow's log file.
func (wf *Workflow) initializeLogging() {
	file, err := os.OpenFile(wf.GetLogFile(),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Couldn't open log file %s : %v", wf.GetLogFile(), err)
	}
	multi := io.MultiWriter(file, os.Stderr)
	log.SetOutput(multi)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// log.New(multi, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// GetBundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in info.plist.
func (wf *Workflow) GetBundleId() string {
	if wf.bundleId == "" { // Really old version of Alfred with no envvars?
		err := wf.readInfoPlist()
		if err != nil {
			log.Fatalf("%v", err)
		}
		if wf.bundleId == "" {
			log.Fatalf("No bundle ID set in info.plist. You *must* set a bundle ID to use awgo.")
		}
	}
	return wf.bundleId
}

// GetWorkflowDir returns the path to the workflow's root directory.
func (wf *Workflow) GetWorkflowDir() string {
	if wf.workflowDir == "" {
		dir, err := util.GetWorkflowRoot()
		if err != nil {
			log.Fatalf("%v", err)
		}
		wf.workflowDir = dir
	}
	return wf.workflowDir
}

// GetCacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func (wf *Workflow) GetCacheDir() string {
	if wf.cacheDir == "" { // Really old version of Alfred with no envvars?
		wf.cacheDir = os.ExpandEnv(fmt.Sprintf(
			"$HOME/Library/Caches/com.runningwithcrayons.Alfred-2/Workflow Data/%s",
			wf.GetBundleId()))
	}
	return util.EnsureExists(wf.cacheDir)
}

// GetDataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func (wf *Workflow) GetDataDir() string {
	if wf.dataDir == "" { // Really old version of Alfred with no envvars?
		wf.dataDir = os.ExpandEnv(fmt.Sprintf(
			"$HOME/Library/Application Support/Alfred 2/Workflow Data/%s",
			wf.GetBundleId()))
	}
	return util.EnsureExists(wf.dataDir)
}

// GetLogFile returns the path to the workflow's log file.
func (wf *Workflow) GetLogFile() string {
	return path.Join(wf.GetCacheDir(), fmt.Sprintf("%s.log", wf.GetBundleId()))
}

// NewItem adds and returns a new feedback Item
func (wf *Workflow) NewItem() *Item {
	return wf.Feedback.NewItem()
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// Run runs your workflow function, catching any errors.
func (wf *Workflow) Run(fn func()) {
	startTime := time.Now()
	log.Println("Workflow started -------------------------")
	log.Printf("awgo version %v", Version)

	// Catch any `panic` and display an error in Alfred.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered : %x", r)
			SendError(r.(string))
			os.Exit(1)
		}
	}()

	// Call the workflow's main function.
	fn()

	elapsed := time.Now().Sub(startTime)
	log.Printf("Workflow finished in %v ----------", elapsed)
}

// SendError sends an error message to Alfred.
func (wf *Workflow) SendError(errMsg string) {
	var f Feedback
	it := f.NewItem()
	it.Title = errMsg
	it.SetIcon("", "fileicon")
	err := f.Send()
	if err != nil {
		log.Fatalf("Error generating XML : %v", err)
	}
}

// SendFeedback generates and sends the XML response to Alfred.
func (wf *Workflow) SendFeedback() {
	err := wf.Feedback.Send()
	if err != nil {
		log.Fatalf("Error generating XML : %v", err)
	}
}

// NewWorkflow creates and initialises a new Workflow.
func NewWorkflow() Workflow {
	var w Workflow
	w.loadEnv()
	w.initializeLogging()
	return w
}

func init() {
	defaultWorkflow = NewWorkflow()
}

// GetBundleId returns the bundle ID of the workflow.
// It is retrieved from Alfred's environmental variables or `info.plist`.
func GetBundleId() string {
	return defaultWorkflow.GetBundleId()
}

// GetCacheDir returns the path to the workflow's cache directory.
// The directory will be created if it does not already exist.
func GetCacheDir() string {
	return defaultWorkflow.GetCacheDir()
}

// GetDataDir returns the path to the workflow's data directory.
// The directory will be created if it does not already exist.
func GetDataDir() string {
	return defaultWorkflow.GetDataDir()
}

// GetWorkflowDir returns the path to the workflow's root directory.
func GetWorkflowDir() string {
	return defaultWorkflow.GetWorkflowDir()
}

// NewItem adds and returns a new feedback Item.
func NewItem() *Item {
	return defaultWorkflow.NewItem()
}

// NewFileItem adds and returns an Item pre-populated from path.
func NewFileItem(path string) *Item {
	return defaultWorkflow.NewFileItem(path)
}

// SendError sends an error message to Alfred as XML feedback.
// TODO: Accept an error, not a string
func SendError(err string) {
	defaultWorkflow.SendError(err)
}

// SendFeedback generates and sends the XML response to Alfred.
// The XML is output to STDOUT. At this point, Alfred considers your
// workflow complete; sending further responses will have no effect.
func SendFeedback() {
	defaultWorkflow.SendFeedback()
}

// Run runs your workflow function, catching any errors.
func Run(fn func()) {
	defaultWorkflow.Run(fn)
}
