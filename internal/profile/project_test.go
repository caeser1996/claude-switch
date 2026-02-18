package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectProfileFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .claude-profile
	os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("work\n"), 0644)

	result := detectProjectProfileFrom(tmpDir)
	if result != "work" {
		t.Errorf("expected 'work', got %q", result)
	}
}

func TestDetectProjectProfileInParent(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	os.MkdirAll(subDir, 0755)

	// Profile file in root, search from deep subdir
	os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("parent-profile\n"), 0644)

	result := detectProjectProfileFrom(subDir)
	if result != "parent-profile" {
		t.Errorf("expected 'parent-profile', got %q", result)
	}
}

func TestDetectProjectProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	result := detectProjectProfileFrom(tmpDir)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestDetectProjectProfileEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty file
	os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("  \n"), 0644)

	result := detectProjectProfileFrom(tmpDir)
	if result != "" {
		t.Errorf("expected empty for whitespace-only file, got %q", result)
	}
}

func TestWriteProjectProfile(t *testing.T) {
	tmpDir := t.TempDir()

	err := WriteProjectProfile(tmpDir, "my-profile")
	if err != nil {
		t.Fatalf("WriteProjectProfile failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, ProjectProfileFile))
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}

	if string(data) != "my-profile\n" {
		t.Errorf("expected 'my-profile\\n', got %q", string(data))
	}
}
