//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-12
//

/*
Workflow settings demonstrates binding a struct to Alfred's settings.

The workflow's settings are stored in info.plist/the workflow's
configuration sheet in Alfred Preferences.

These are imported into the Server struct using Config.To().

The Script Filter displays these settings, and you can select one
to change its value.

If you enter a new value, this is saved to info.plist/the configuration
sheet via Config.Set(), and the workflow is run again by calling
its "settings" External Trigger via Alfred.RunTrigger().
*/
package main

import (
	"flag"
	"fmt"
	"log"

	aw "github.com/deanishe/awgo"
)

// Server contains the configuration loaded from the workflow's settings
// in the configuration sheet.
type Server struct {
	Hostname   string
	PortNumber int `env:"PORT"`
	Username   string
	APIKey     string
}

var (
	srv *Server
	wf  *aw.Workflow
	// Command-line arguments
	setKey, getKey string
	query          string
)

func init() {
	wf = aw.New()
	flag.StringVar(&setKey, "set", "", "save a value")
	flag.StringVar(&getKey, "get", "", "enter a new value")
}

func runSet(key, value string) {

	wf.Configure(aw.TextErrors(true))

	log.Printf("saving %#v to %s ...", value, key)

	if err := wf.Config.Set(key, value, false).Do(); err != nil {
		wf.FatalError(err)
	}

	if err := wf.Alfred.RunTrigger("settings", "").Do(); err != nil {
		wf.FatalError(err)
	}

	log.Printf("saved %#v to %s", value, key)
}

func runGet(key, value string) {

	log.Printf("getting new %s ...", key)

	if value != "" {

		var varname string

		switch key {
		case "API key":
			varname = "API_KEY"
		case "hostname":
			varname = "HOSTNAME"
		case "port number":
			varname = "PORT"
		case "username":
			varname = "USERNAME"
		}

		wf.NewItem(fmt.Sprintf("Set %s to “%s”", key, value)).
			Subtitle("↩ to save").
			Arg(value).
			Valid(true).
			Var("value", value).
			Var("varname", varname)

	}

	wf.WarnEmpty(fmt.Sprintf("Enter %s", key), "")
	wf.SendFeedback()
}

func run() {

	wf.Args() // call to handle magic actions

	// ----------------------------------------------------------------
	// Load configuration

	// Default configuration
	srv = &Server{
		Hostname:   "localhost",
		PortNumber: 6000,
		Username:   "anonymous",
	}

	// Update config from environment variables
	if err := wf.Config.To(srv); err != nil {
		panic(err)
	}

	log.Printf("loaded: %#v", srv)

	// ----------------------------------------------------------------
	// Parse command-line flags and decide what to do

	flag.Parse()
	args := flag.Args()

	if len(args) > 0 {
		query = args[0]
	}

	log.Printf("query=%s", query)

	if setKey != "" {
		runSet(setKey, query)
		return
	}

	if getKey != "" {
		runGet(getKey, query)
		return
	}

	// ----------------------------------------------------------------
	// Run Script action.

	wf.NewItem("Hostname: "+srv.Hostname).
		Subtitle("↩ to edit").
		Valid(true).
		Var("name", "hostname")

	wf.NewItem(fmt.Sprintf("Port: %d", srv.PortNumber)).
		Subtitle("↩ to edit").
		Valid(true).
		Var("name", "port number")

	wf.NewItem("Username: "+srv.Username).
		Subtitle("↩ to edit").
		Valid(true).
		Var("name", "username")

	wf.NewItem("API Key: "+srv.APIKey).
		Subtitle("↩ to edit").
		Valid(true).
		Var("name", "API key")

	if query != "" {
		wf.Filter(query)
	}

	wf.WarnEmpty("No Matching Items", "Try a different query?")
	wf.SendFeedback()

}

func main() {
	wf.Run(run)
}
