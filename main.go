// +build go1.7

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/jawher/mow.cli"
	"github.com/k-takata/go-iscygpty"
	"github.com/mattn/go-isatty"
	"github.com/raoulh/go-progress"
)

const (
	CharStar     = "\u2737"
	CharAbort    = "\u2718"
	CharCheck    = "\u2714"
	CharWarning  = "\u26A0"
	CharArrow    = "\u2012\u25b6"
	CharVertLine = "\u2502"
)

var (
	blue       = color.New(color.FgBlue).SprintFunc()
	errorRed   = color.New(color.FgRed).SprintFunc()
	errorBgRed = color.New(color.BgRed, color.FgBlack).SprintFunc()
	green      = color.New(color.FgGreen).SprintFunc()
	cyan       = color.New(color.FgCyan).SprintFunc()
	bgCyan     = color.New(color.FgWhite).SprintFunc()
)

var (
	mcUrl    *string
	debugOpt *bool

	optContext   *string
	optLogin     *string
	optPass      *string
	optDesc      *string
	optPrintDesc *bool
	optFilename  *string
	optParameter *string
	optValue     *string

	isTerminal  bool
	progressBar *progress.ProgressBar
)

func exit(err error, exit int) {
	fmt.Fprintln(os.Stderr, errorRed(CharAbort), err)
	cli.Exit(exit)
}

func checkLog() {
	if !*debugOpt {
		//completely disable debug output
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
}

func addDefaultArgs(cmd *cli.Cmd) {
	mcUrl = cmd.StringOpt("m moolticute_url", MOOLTICUTE_DAEMON_URL, "Use a different url for connecting to moolticute")
	debugOpt = cmd.BoolOpt("debug", false, "Add debug log to stdout")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// fix for cygwin terminal
	if iscygpty.IsCygwinPty(os.Stdout.Fd()) || isatty.IsTerminal(os.Stdout.Fd()) {
		color.NoColor = false
		isTerminal = true
	}

	app := cli.App("mc-cli", "Command line tool to interact with a mooltipass device through a moolticute daemon")

	app.Command("login", "Manage credentials stored in the device", func(cmd *cli.Cmd) {
		cmd.Command("get", "Get a password for given context", func(cmd *cli.Cmd) {
			optContext = cmd.StringArg("CONTEXT", "", "Context to work on")
			optLogin = cmd.StringArg("LOGIN", "", "Login to use")
			addDefaultArgs(cmd)
			optPrintDesc = cmd.BoolOpt("d description", false, "Output service description instead of password")
			cmd.Spec = "CONTEXT LOGIN [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				if err := processLoginCmd("get", *optContext, *optLogin, "", "", *optPrintDesc); err != nil {
					exit(err, 1)
				}
			}
		})
		cmd.Command("set", "Add or update a context", func(cmd *cli.Cmd) {
			optContext = cmd.StringArg("CONTEXT", "", "Context to work on")
			optLogin = cmd.StringArg("LOGIN", "", "Login to set")
			optPass = cmd.StringArg("PASS", "", "Password to set (would be asked if not set in command line)")
			optDesc = cmd.StringArg("DESC", "", "Description to set")
			addDefaultArgs(cmd)
			cmd.Spec = "CONTEXT LOGIN [PASS] [DESC] [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				if err := processLoginCmd("set", *optContext, *optLogin, *optPass, *optDesc, false); err != nil {
					exit(err, 1)
				}
			}
		})
	})

	app.Command("data", "Import & export small files stored in the device", func(cmd *cli.Cmd) {
		cmd.Command("get", "Retrieve data for given context", func(cmd *cli.Cmd) {
			optContext = cmd.StringArg("CONTEXT", "", "Context to work on")
			addDefaultArgs(cmd)
			cmd.Spec = "CONTEXT [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				if err := processDataCmd("get", *optContext, "", progressFunc); err != nil {
					exit(err, 1)
				}
			}
		})
		cmd.Command("set", "Add or update data for given context", func(cmd *cli.Cmd) {
			optContext = cmd.StringArg("CONTEXT", "", "Context to work on")
			optFilename = cmd.StringArg("FILENAME", "", "File to save in the device")
			addDefaultArgs(cmd)
			cmd.Spec = "CONTEXT FILENAME [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				if err := processDataCmd("set", *optContext, *optFilename, progressFunc); err != nil {
					exit(err, 1)
				}
			}
		})
	})

	app.Command("parameters", "Get/Set device parameters", func(cmd *cli.Cmd) {
		cmd.Command("get", "Retrieve device parameter", func(cmd *cli.Cmd) {
			optParameter = cmd.StringArg("PARAMETER", "", "Selected device parameter")
			addDefaultArgs(cmd)
			cmd.Spec = "PARAMETER [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				err := processParameterCmd("get", *optParameter, "")
				if err != nil {
					exit(err, 1)
				}
			}
		})
		cmd.Command("set", "Set device parameter", func(cmd *cli.Cmd) {
			optParameter = cmd.StringArg("PARAMETER", "", "Selected device parameter")
			optValue = cmd.StringArg("VALUE", "", "Value to set the parameter to")
			addDefaultArgs(cmd)
			cmd.Spec = "PARAMETER VALUE [OPTIONS]"

			cmd.Action = func() {
				checkLog()

				err := processParameterCmd("set", *optParameter, *optValue)
				if err != nil {
					exit(err, 1)
				}
			}
		})
	})

	if err := app.Run(os.Args); err != nil {
		exit(err, 1)
	}
}

func progressFunc(total, current int) {
	if !isTerminal {
		return
	}

	if progressBar == nil {
		progressBar = progress.New(total)
		progressBar.Format = progress.ProgressFormats[2]
	}

	progressBar.Set(current)
}
