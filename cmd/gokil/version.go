package main

import (
	"fmt"
	"runtime/debug"
)

// Set at build time via -ldflags (optional).
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func printVersion() {
	version := resolveVersion()
	fmt.Printf("gokil %s", version)
	if Commit != "" {
		fmt.Printf(" (%s)", Commit)
	}
	if Date != "" {
		fmt.Printf(" built %s", Date)
	}
	fmt.Println()
}

func resolveVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return Version
}
