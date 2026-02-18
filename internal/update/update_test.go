package update

import (
	"runtime"
	"testing"
)

func TestExpectedAssetName(t *testing.T) {
	name := expectedAssetName()

	if name == "" {
		t.Error("asset name should not be empty")
	}

	// Should contain OS and arch
	os := runtime.GOOS
	arch := runtime.GOARCH

	expected := "claude-switch_" + os + "_" + arch
	if runtime.GOOS == "windows" {
		expected += ".exe"
	}

	if name != expected {
		t.Errorf("expected %q, got %q", expected, name)
	}
}

func TestCheckVersionComparison(t *testing.T) {
	// We can't actually call GitHub API in unit tests,
	// but we can verify the function handles errors gracefully.
	// Testing with a clearly invalid version to ensure it doesn't panic.
	result, err := Check("dev")
	if err != nil {
		// Network error is expected in CI - not a test failure
		t.Skipf("Skipping (network unavailable): %v", err)
	}

	if result == nil {
		t.Fatal("Check returned nil result")
	}

	if result.CurrentVersion != "dev" {
		t.Errorf("expected current version 'dev', got %q", result.CurrentVersion)
	}
}
