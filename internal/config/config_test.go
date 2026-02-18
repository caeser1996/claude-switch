package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}
	if cfg.Profiles == nil {
		t.Fatal("Profiles map is nil")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ActiveProfile != "" {
		t.Errorf("expected empty active profile, got %q", cfg.ActiveProfile)
	}
}

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if !s.AutoBackup {
		t.Error("expected AutoBackup to be true")
	}
	if s.MaxBackups != 10 {
		t.Errorf("expected MaxBackups=10, got %d", s.MaxBackups)
	}
	if !s.ColorOutput {
		t.Error("expected ColorOutput to be true")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temp dir to act as home
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := NewConfig()
	cfg.ActiveProfile = "test"
	cfg.Profiles["test"] = ProfileEntry{
		Name:  "test",
		Email: "test@example.com",
	}

	// Save
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tmpDir, AppDir, ConfigFile)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", expectedPath)
	}

	// Load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ActiveProfile != "test" {
		t.Errorf("expected active profile 'test', got %q", loaded.ActiveProfile)
	}
	if p, ok := loaded.Profiles["test"]; !ok {
		t.Error("profile 'test' not found after load")
	} else if p.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", p.Email)
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load of non-existent config should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil config")
	}
	if len(cfg.Profiles) != 0 {
		t.Error("expected empty profiles for new config")
	}
}

func TestConfigJSON(t *testing.T) {
	cfg := NewConfig()
	cfg.ActiveProfile = "work"
	cfg.Profiles["work"] = ProfileEntry{
		Name:        "work",
		Email:       "user@work.com",
		Description: "Work account",
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if loaded.ActiveProfile != "work" {
		t.Errorf("expected 'work', got %q", loaded.ActiveProfile)
	}
}

func TestEnsureDirs(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Check all directories exist
	dirs := []string{
		filepath.Join(tmpDir, AppDir),
		filepath.Join(tmpDir, AppDir, "profiles"),
		filepath.Join(tmpDir, AppDir, "backups"),
	}
	for _, d := range dirs {
		info, err := os.Stat(d)
		if os.IsNotExist(err) {
			t.Errorf("directory %s not created", d)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}
