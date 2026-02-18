package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/caeser1996/claude-switch/internal/config"
)

// setupTestEnv creates a temporary home directory with a mock Claude config.
// Works on both Unix and Windows by setting the appropriate env vars.
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")

	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("cannot set HOME: %v", err)
	}
	if runtime.GOOS == "windows" {
		if err := os.Setenv("USERPROFILE", tmpDir); err != nil {
			t.Fatalf("cannot set USERPROFILE: %v", err)
		}
	}

	// Create mock Claude config directory with credentials
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatalf("cannot create claude dir: %v", err)
	}

	// Create mock credentials file
	credContent := `{"email": "test@example.com", "token": "mock-token-123"}`
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(credContent), 0600); err != nil {
		t.Fatalf("cannot write credentials: %v", err)
	}

	// Create claude-switch directories
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("cannot create dirs: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", origHome)
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", origUserProfile)
		}
	}
	return tmpDir, cleanup
}

func TestImportProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	if err := mgr.Import("work", "Work account"); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify profile was added to config
	if _, ok := cfg.Profiles["work"]; !ok {
		t.Error("profile 'work' not found in config")
	}

	if cfg.Profiles["work"].Description != "Work account" {
		t.Errorf("expected description 'Work account', got %q", cfg.Profiles["work"].Description)
	}

	// First import should set active profile
	if cfg.ActiveProfile != "work" {
		t.Errorf("expected active profile 'work', got %q", cfg.ActiveProfile)
	}

	// Verify credentials were copied
	profilesDir, _ := config.ProfilesDir()
	credPath := filepath.Join(profilesDir, "work", ".credentials.json")
	if !FileExists(credPath) {
		t.Error("credentials file not copied to profile dir")
	}
}

func TestImportInvalidName(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name", false},
		{"valid_name", false},
		{"ValidName123", false},
		{"", true},
		{"has spaces", true},
		{"has/slash", true},
		{"has.dot", true},
	}

	for _, tt := range tests {
		err := mgr.Import(tt.name, "")
		if tt.wantErr && err == nil {
			t.Errorf("Import(%q): expected error, got nil", tt.name)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("Import(%q): unexpected error: %v", tt.name, err)
		}
	}
}

func TestUseProfile(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	// Import two profiles (need to change credentials between imports)
	if err := mgr.Import("profile1", "First"); err != nil {
		t.Fatalf("Import profile1 failed: %v", err)
	}

	// Change the credentials to simulate a different account
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"),
		[]byte(`{"email": "other@example.com", "token": "other-token"}`), 0600); err != nil {
		t.Fatalf("cannot write credentials: %v", err)
	}

	if err := mgr.Import("profile2", "Second"); err != nil {
		t.Fatalf("Import profile2 failed: %v", err)
	}

	// Switch to profile1
	if err := mgr.Use("profile1"); err != nil {
		t.Fatalf("Use profile1 failed: %v", err)
	}

	if cfg.ActiveProfile != "profile1" {
		t.Errorf("expected active profile 'profile1', got %q", cfg.ActiveProfile)
	}

	// Verify the credentials were restored
	credPath := filepath.Join(claudeDir, ".credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("cannot read credentials: %v", err)
	}
	if string(data) == `{"email": "other@example.com", "token": "other-token"}` {
		t.Error("credentials were not updated after switching")
	}
}

func TestUseNonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	err := mgr.Use("nonexistent")
	if err == nil {
		t.Error("expected error when using nonexistent profile")
	}
}

func TestRemoveProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	if err := mgr.Import("to-remove", ""); err != nil {
		t.Fatalf("Import to-remove failed: %v", err)
	}

	// Cannot remove active profile
	err := mgr.Remove("to-remove")
	if err == nil {
		t.Error("expected error when removing active profile")
	}

	// Import another and switch
	if err := mgr.Import("keeper", ""); err != nil {
		t.Fatalf("Import keeper failed: %v", err)
	}
	if err := mgr.Use("keeper"); err != nil {
		t.Fatalf("Use keeper failed: %v", err)
	}

	// Now remove should work
	if err := mgr.Remove("to-remove"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, ok := cfg.Profiles["to-remove"]; ok {
		t.Error("profile 'to-remove' still exists after removal")
	}
}

func TestListProfiles(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	// Empty list
	profiles := mgr.List()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// Add some profiles
	if err := mgr.Import("beta", ""); err != nil {
		t.Fatalf("Import beta failed: %v", err)
	}
	if err := mgr.Import("alpha", ""); err != nil {
		t.Fatalf("Import alpha failed: %v", err)
	}

	profiles = mgr.List()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Should be sorted
	if profiles[0].Name != "alpha" {
		t.Errorf("expected first profile 'alpha', got %q", profiles[0].Name)
	}
	if profiles[1].Name != "beta" {
		t.Errorf("expected second profile 'beta', got %q", profiles[1].Name)
	}
}

func TestCurrentProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	// No active profile
	_, err := mgr.Current()
	if err == nil {
		t.Error("expected error when no profile is active")
	}

	// Import sets first as active
	if err := mgr.Import("myprofile", ""); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	p, err := mgr.Current()
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}
	if p.Name != "myprofile" {
		t.Errorf("expected 'myprofile', got %q", p.Name)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "subdir", "dest.txt")

	content := "hello, world"
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatalf("cannot write source: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("cannot read destination: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent
	if FileExists(filepath.Join(tmpDir, "nope")) {
		t.Error("expected false for non-existent file")
	}

	// File
	f := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}
	if !FileExists(f) {
		t.Error("expected true for existing file")
	}

	// Directory should return false
	if FileExists(tmpDir) {
		t.Error("expected false for directory")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()

	if !DirExists(tmpDir) {
		t.Error("expected true for existing directory")
	}

	if DirExists(filepath.Join(tmpDir, "nope")) {
		t.Error("expected false for non-existent directory")
	}

	// File should return false
	f := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}
	if DirExists(f) {
		t.Error("expected false for file")
	}
}
