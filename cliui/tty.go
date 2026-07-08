package cliui

import (
	"os"
	"strings"
)

type fdWriter interface {
	Write([]byte) (int, error)
	Fd() uintptr
}

func IsTTY(w fdWriter) bool {
	if w == nil {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func ColorsEnabled(w fdWriter) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if strings.EqualFold(os.Getenv("CI"), "true") {
		return false
	}
	return IsTTY(w)
}
