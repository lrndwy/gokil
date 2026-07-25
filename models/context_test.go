package models

import (
	"testing"
)

func TestGIDNonZero(t *testing.T) {
	id := gid()
	if id == 0 {
		t.Fatal("expected non-zero goroutine id")
	}
}

func TestGIDStableWithinGoroutine(t *testing.T) {
	a := gid()
	b := gid()
	if a != b {
		t.Fatalf("gid changed within same goroutine: %d vs %d", a, b)
	}
}
