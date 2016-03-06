package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gogs.deanishe.net/deanishe/awgo"
)

var (
	StartDir string // Directory to read
)

func init() {
	rand.Seed(time.Now().Unix())
	StartDir = os.Getenv("HOME")
}

// tossCoin returns true/false with 50/50 probability
func tossCoin() bool {
	i := rand.Intn(2)
	return i == 1
}

// readDir returns the paths to all the visible subdirectories of `dirpath`
func readDir(dirpath string) []string {
	paths := []string{}
	files, _ := ioutil.ReadDir(dirpath)
	for _, fi := range files {
		// Ignore files and hidden files
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		paths = append(paths, filepath.Join(dirpath, fi.Name()))
	}
	return paths
}

// run runs the workflow
func run() {
	log.Printf("bundleId=%v", workflow.GetBundleId())
	root := workflow.GetWorkflowDir()
	log.Printf("workflowdir=%v", root)
	log.Printf("datadir=%v", workflow.GetDataDir())
	log.Printf("cachedir=%v", workflow.GetCacheDir())
	for i, path := range readDir(StartDir) {
		log.Printf("i=%02d, f=%v", i, path)
		it := workflow.NewFileItem(path)
		it.SetSubtitle("cmd", "Open in your underpants")
	}
	workflow.SendFeedback()
}

func main() {
	// Call workflow via `Run` wrapper to catch errors.
	workflow.Run(run)
}
