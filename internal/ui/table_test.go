package ui

import (
	"bytes"
	"os"
	"testing"
)

func TestNewTable(t *testing.T) {
	table := NewTable("Col1", "Col2")
	if len(table.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(table.Headers))
	}
	if len(table.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(table.Rows))
	}
}

func TestTableAddRow(t *testing.T) {
	table := NewTable("Name", "Value")
	table.AddRow("foo", "bar")
	table.AddRow("baz", "qux")

	if len(table.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.Rows))
	}
}

func TestTableRender(t *testing.T) {
	SetColorEnabled(false)
	defer SetColorEnabled(true)

	table := NewTable("Name", "Status")
	table.AddRow("profile1", "active")
	table.AddRow("profile2", "inactive")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	table.Render()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Render produced no output")
	}

	// Should contain headers
	if !contains(output, "Name") || !contains(output, "Status") {
		t.Error("output should contain header names")
	}

	// Should contain data
	if !contains(output, "profile1") || !contains(output, "active") {
		t.Error("output should contain row data")
	}
}

func TestTableRenderEmpty(t *testing.T) {
	table := &Table{}

	// Should not panic
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	table.Render()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	// Empty headers = no output, just verify no panic
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
