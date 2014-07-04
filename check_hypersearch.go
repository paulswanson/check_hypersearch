package main

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	NAGIOS_OK       int = 0
	NAGIOS_WARNING  int = 1
	NAGIOS_CRITICAL int = 2
	NAGIOS_UNKNOWN  int = 3
)

var exitCode int

func main() {

	app := cli.NewApp()
	app.Name = "check_hypersearch"
	app.Version = "1.0.0"
	app.Author = "Paul Swanson"
	app.Usage = "Search for text on a web page"
	app.Flags = []cli.Flag{
		cli.StringFlag{"require,r", "all", "Require 'all' or 'some'"},
		cli.BoolFlag{"quiet, q", "Be quiet"},
		cli.BoolFlag{"verbose", "Be verbose; for debugging etc."},
	}

	cli.AppHelpTemplate = `NAME:
{{.Name}} - {{.Usage}}

USAGE:
{{.Name}} [global options] [arguments...]

For example, {{.Name}} --require some http://www.google.com/ "<title>Google</title>" "Privacy & Terms"

VERSION:
{{.Version}}

GLOBAL OPTIONS:
{{range .Flags}}{{.}}
{{end}}
`

	app.Action = func(c *cli.Context) {

		var requireAll, quiet, verbose bool
		var found int

		if c.String("require") == "some" {
			requireAll = false
		} else {
			requireAll = true
		}

		if c.Bool("quiet") {
			quiet = true
		}

		if c.Bool("verbose") {
			verbose = true
		}

		args := c.Args()
		argCount := len(args)

		if argCount < 2 {
			cli.ShowAppHelp(c)
			exitCode = NAGIOS_UNKNOWN
			return
		}

		if verbose {
			println("Accessing", args[0])
		}
		resp, err := http.Get(args[0])
		if err != nil {
			println("Couldn't access that link!")
			exitCode = NAGIOS_UNKNOWN
			return
		}
		defer resp.Body.Close()

		if verbose {
			println("Reading page...")
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("Couldn't read that page!")
			exitCode = NAGIOS_UNKNOWN
		}

		queryCount := argCount - 1

		for _, s := range args[1:] {
			if bytes.Contains(body, []byte(s)) {
				found++
				if verbose {
					println("Found:", s)
				}
			} else {
				if verbose {
					println("Not found:", s)
				}
			}
		}

		var statusMessage string

		switch {
		case queryCount == found:
			statusMessage = "OK."
			exitCode = NAGIOS_OK
		case found == 0 || requireAll:
			statusMessage = "FAIL."
			exitCode = NAGIOS_CRITICAL
		default:
			statusMessage = "Some OK."
			exitCode = NAGIOS_WARNING
		}

		if !quiet {
			fmt.Printf("Found %v of %v %v\n", found, queryCount, statusMessage)
		}
		if verbose {
			println("Nagios exit code:", exitCode)
		}

	}

	app.Run(os.Args)

	os.Exit(exitCode)
}
