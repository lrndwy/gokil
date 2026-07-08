package cliui

import (
	"fmt"
	"io"
	"os"
)

const (
	codeReset  = "\033[0m"
	codeBold   = "\033[1m"
	codeDim    = "\033[2m"
	codeGreen  = "\033[32m"
	codeCyan   = "\033[36m"
	codeYellow = "\033[33m"
	codeRed    = "\033[31m"
)

type Style struct {
	out     io.Writer
	enabled bool
}

func NewStyle(out io.Writer) *Style {
	w, ok := out.(fdWriter)
	enabled := ok && ColorsEnabled(w)
	return &Style{out: out, enabled: enabled}
}

func (s *Style) color(code, text string) string {
	if !s.enabled {
		return text
	}
	return code + text + codeReset
}

func (s *Style) Bold(text string) string   { return s.color(codeBold, text) }
func (s *Style) Dim(text string) string    { return s.color(codeDim, text) }
func (s *Style) Green(text string) string  { return s.color(codeGreen, text) }
func (s *Style) Cyan(text string) string   { return s.color(codeCyan, text) }
func (s *Style) Yellow(text string) string { return s.color(codeYellow, text) }
func (s *Style) Red(text string) string    { return s.color(codeRed, text) }

var defaultStyle = NewStyle(os.Stdout)

func Bold(text string) string   { return defaultStyle.Bold(text) }
func Dim(text string) string    { return defaultStyle.Dim(text) }
func Green(text string) string  { return defaultStyle.Green(text) }
func Cyan(text string) string   { return defaultStyle.Cyan(text) }
func Yellow(text string) string { return defaultStyle.Yellow(text) }
func Red(text string) string    { return defaultStyle.Red(text) }

func Successf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Green("✔"), fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Cyan("ℹ"), fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", Yellow("⚠"), fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("✖"), fmt.Sprintf(format, args...))
}
