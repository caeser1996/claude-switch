package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectProfileFound(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("work\n"), 0644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

	result := detectProjectProfileFrom(tmpDir)
	if result != "work" {
		t.Errorf("expected 'work', got %q", result)
	}
}

func TestDetectProjectProfileInParent(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("cannot create subdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("parent-profile\n"), 0644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

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

	if err := os.WriteFile(filepath.Join(tmpDir, ProjectProfileFile), []byte("  \n"), 0644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

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
