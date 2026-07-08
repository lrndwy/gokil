package cliui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// npm/ora-style braille dot frames.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	out      io.Writer
	style    *Style
	text     string
	frame    int
	active   bool
	mu       sync.Mutex
	stopCh   chan struct{}
	doneCh   chan struct{}
	animated bool
}

func NewSpinner(out io.Writer) *Spinner {
	if out == nil {
		out = os.Stdout
	}
	w, ok := out.(fdWriter)
	animated := ok && IsTTY(w) && os.Getenv("CI") == ""
	return &Spinner{
		out:      out,
		style:    NewStyle(out),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		animated: animated,
	}
}

func (s *Spinner) Start(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.text = text
	if !s.animated {
		return
	}

	if s.active {
		s.renderLocked()
		return
	}

	s.active = true
	s.frame = 0
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go s.run()
}

func (s *Spinner) Update(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text = text
	if s.active && s.animated {
		s.renderLocked()
	}
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	if s.animated {
		close(s.stopCh)
		s.mu.Unlock()
		<-s.doneCh
		return
	}
	s.mu.Unlock()
}

func (s *Spinner) Success(text string) {
	s.Stop()
	fmt.Fprintf(s.out, "%s %s\n", s.style.Green("✔"), text)
}

func (s *Spinner) Fail(text string) {
	s.Stop()
	fmt.Fprintf(s.out, "%s %s\n", s.style.Red("✖"), text)
}

func (s *Spinner) run() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	defer close(s.doneCh)

	s.mu.Lock()
	s.renderLocked()
	s.mu.Unlock()

	for {
		select {
		case <-s.stopCh:
			s.clearLine()
			return
		case <-ticker.C:
			s.mu.Lock()
			s.frame = (s.frame + 1) % len(spinnerFrames)
			s.renderLocked()
			s.mu.Unlock()
		}
	}
}

func (s *Spinner) renderLocked() {
	fmt.Fprintf(s.out, "\r%s %s", s.style.Cyan(spinnerFrames[s.frame]), s.text)
}

func (s *Spinner) clearLine() {
	fmt.Fprint(s.out, "\r\033[K")
}

// WithSpinner runs fn while showing an animated spinner (npm-style).
func WithSpinner(text string, fn func() error) error {
	sp := NewSpinner(os.Stdout)
	sp.Start(text)
	err := fn()
	if err != nil {
		sp.Fail(text)
		return err
	}
	sp.Success(text)
	return nil
}
