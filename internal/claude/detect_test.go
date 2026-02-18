package claude

import (
	"testing"
)

func TestDetect(t *testing.T) {
	result := Detect()
	// We can't guarantee claude is installed in test environments,
	// so we just verify the struct is populated correctly.
	if result.Found {
		if result.Path == "" {
			t.Error("Found is true but Path is empty")
		}
	} else {
		if result.Path != "" {
			t.Error("Found is false but Path is not empty")
		}
	}
}
