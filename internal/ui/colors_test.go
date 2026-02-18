package ui

import (
	"testing"
)

func TestColorize(t *testing.T) {
	// With colors enabled
	SetColorEnabled(true)
	result := Colorize(Red, "error")
	if result != Red+"error"+Reset {
		t.Errorf("expected colored output, got %q", result)
	}

	// With colors disabled
	SetColorEnabled(false)
	result = Colorize(Red, "error")
	if result != "error" {
		t.Errorf("expected plain output, got %q", result)
	}

	// Restore
	SetColorEnabled(true)
}

func TestSetColorEnabled(t *testing.T) {
	SetColorEnabled(false)
	result := Colorize(Green, "ok")
	if result != "ok" {
		t.Error("expected colors disabled")
	}

	SetColorEnabled(true)
	result = Colorize(Green, "ok")
	if result == "ok" {
		t.Error("expected colors enabled")
	}
}
