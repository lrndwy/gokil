package main

import (
	"github.com/lrndwy/gokil/cliui"
	"github.com/lrndwy/gokil/version"
)

// Override at build time via -ldflags (optional).
var (
	Version = ""
	Commit  = ""
	Date    = ""
)

func printVersion() {
	info := version.Get()

	if Version != "" {
		info.Version = Version
	}
	if Commit != "" {
		info.Commit = Commit
	}
	if Date != "" {
		info.Date = Date
	}

	println(cliui.Cyan("gokil") + " " + info.String())
}
