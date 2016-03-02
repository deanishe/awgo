package workflow

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"

	"gogs.deanishe.net/deanishe/awgo/util"
)

// The workflow object operated on by top-level functions.
var defaultWorkflow Workflow

// Workflow provides an API for Alfred.
type Workflow struct {
	Feedback Feedback

	bundleID    string
	dataDir     string
	cacheDir    string
	workflowDir string
}

// GetBundleID returns the workflow's bundle ID. This library will not
// work without a bundle ID, which is set in info.plist.
func (wf *Workflow) GetBundleID() string {
	if wf.bundleID != "" {
		return wf.bundleID
	}
	// TODO: Find bundle ID
	return "net.deanishe.alfred-go"
}

func (wf *Workflow) GetWorkflowDir() string {
	if wf.workflowDir != "" {
		log.Printf("wf.workflowDir already set to %v", wf.workflowDir)
		return wf.workflowDir
	}
	dir, err := util.GetWorkflowRoot()
	if err != nil {
		log.Fatalf("%v", err)
	}
	wf.workflowDir = dir
	return dir
}

func (wf *Workflow) GetDataDir() string {
	if wf.dataDir != "" {
		return wf.dataDir
	}
	dir := os.ExpandEnv(
		fmt.Sprintf(
			"$HOME/Application Support/Alfred 2/Workflow Data/%s",
			wf.GetBundleID()))
	os.MkdirAll(dir, 0700)
	wf.dataDir = dir
	return dir
}

// NewItem adds and returns a new feedback Item
func (wf *Workflow) NewItem() *Item {
	return wf.Feedback.NewItem()
}

// NewFileItem adds and returns a new feedback Item pre-populated from path.
func (wf *Workflow) NewFileItem(path string) *Item {
	return wf.Feedback.NewFileItem(path)
}

// SendFeedback generates and sends the XML response to Alfred.
func (wf *Workflow) SendFeedback() {
	output, err := xml.MarshalIndent(wf.Feedback, "", "  ")
	if err != nil {
		log.Fatalf("Error generating XML : %v", err)
	}
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
}

// NewWorkflow creates a new Workflow.
func NewWorkflow() Workflow {
	return Workflow{}
}

func init() {
	defaultWorkflow = NewWorkflow()
}

func GetBundleID() string {
	return defaultWorkflow.GetBundleID()
}

func GetDataDir() string {
	return defaultWorkflow.GetDataDir()
}

func GetWorkflowDir() string {
	return defaultWorkflow.GetWorkflowDir()
}

// NewItem adds and returns a new feedback Item.
func NewItem() *Item {
	return defaultWorkflow.NewItem()
}

// NewFileItem adds and returns an Item pre-populated from path
func NewFileItem(path string) *Item {
	return defaultWorkflow.NewFileItem(path)
}

// SendFeedback generates and sends the XML response to Alfred.
// The XML is output to STDOUT. At this point, Alfred considers your
// workflow complete; sending further responses will have no effect.
func SendFeedback() {
	defaultWorkflow.SendFeedback()
}

// Run runs your workflow function, catching any errors.
func Run(fn func()) {
	startTime := time.Now()
	fn()
	elapsed := time.Now().Sub(startTime)
	log.Printf("Finished in %0.4f seconds ------------------",
		elapsed.Seconds())
}
