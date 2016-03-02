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
	StartDir string
)

func init() {
	rand.Seed(time.Now().Unix())
	StartDir = os.Getenv("HOME")
}

// shouldI returns true/false with 50/50 probability
func shouldI() bool {
	i := rand.Intn(2)
	return i == 1
}

// readDir returns the paths to all the visible subdirectories of `dirpath`
func readDir(dirpath string) []string {
	paths := []string{}
	files, _ := ioutil.ReadDir(dirpath)
	for _, fi := range files {
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		paths = append(paths, filepath.Join(dirpath, fi.Name()))
	}
	return paths
}

// run runs the workflow
func run() {
	root := workflow.GetWorkflowDir()
	log.Printf("workflow root=%v", root)
	for i, path := range readDir(StartDir) {
		log.Printf("i=%v, f=%v", i, path)
		it := workflow.NewFileItem(path)
		it.SetSubtitle("cmd", "Open in your underpants")
	}
	workflow.SendFeedback()
}

func main() {
	workflow.Run(run)
}
