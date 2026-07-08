package cliui

import (
	"bytes"
	"strings"
	"testing"
)

type fakeTTY struct {
	bytes.Buffer
}

func (f *fakeTTY) Fd() uintptr { return 1 }

func TestSpinnerNonTTYPrintsSuccess(t *testing.T) {
	var buf bytes.Buffer
	sp := NewSpinner(&buf)
	sp.Start("working")
	sp.Success("done")

	out := buf.String()
	if strings.Contains(out, "working") {
		t.Fatalf("non-tty start should not print spinner text, got %q", out)
	}
	if !strings.Contains(out, "✔") || !strings.Contains(out, "done") {
		t.Fatalf("expected success output, got %q", out)
	}
}

func TestColorsDisabledWithoutTTY(t *testing.T) {
	var buf fakeTTY
	if ColorsEnabled(&buf) {
		t.Fatal("colors should be disabled for non-tty writer")
	}
}

func TestBoldWithoutColors(t *testing.T) {
	var buf fakeTTY
	st := NewStyle(&buf)
	if st.Bold("x") != "x" {
		t.Fatal("expected plain text without colors")
	}
}
