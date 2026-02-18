package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
)

func TestSetupIsolatedEnv(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	// Import a profile first
	if err := mgr.Import("isolated-test", "Test profile"); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Setup isolated environment
	env, err := SetupIsolatedEnv("isolated-test", cfg)
	if err != nil {
		t.Fatalf("SetupIsolatedEnv failed: %v", err)
	}
	defer env.Cleanup()

	// Verify temp dir was created
	if env.TempDir == "" {
		t.Fatal("TempDir is empty")
	}
	if !DirExists(env.TempDir) {
		t.Fatalf("TempDir %s does not exist", env.TempDir)
	}

	// Verify credentials were copied
	credPath := filepath.Join(env.TempDir, ".credentials.json")
	if !FileExists(credPath) {
		t.Error("credentials not copied to isolated env")
	}

	// Verify profile name
	if env.ProfileName != "isolated-test" {
		t.Errorf("expected ProfileName 'isolated-test', got %q", env.ProfileName)
	}
}

func TestSetupIsolatedEnvNonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()

	_, err := SetupIsolatedEnv("nonexistent", cfg)
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestIsolatedEnvEnv(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	mgr.Import("env-test", "")

	env, err := SetupIsolatedEnv("env-test", cfg)
	if err != nil {
		t.Fatalf("SetupIsolatedEnv failed: %v", err)
	}
	defer env.Cleanup()

	envVars := env.Env()

	// Check CLAUDE_CONFIG_DIR is set
	found := false
	for _, v := range envVars {
		if strings.HasPrefix(v, "CLAUDE_CONFIG_DIR=") {
			found = true
			val := strings.TrimPrefix(v, "CLAUDE_CONFIG_DIR=")
			if val != env.TempDir {
				t.Errorf("expected CLAUDE_CONFIG_DIR=%s, got %s", env.TempDir, val)
			}
		}
	}
	if !found {
		t.Error("CLAUDE_CONFIG_DIR not found in env")
	}

	// Ensure no duplicate CLAUDE_CONFIG_DIR
	count := 0
	for _, v := range envVars {
		if strings.HasPrefix(v, "CLAUDE_CONFIG_DIR=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 CLAUDE_CONFIG_DIR, found %d", count)
	}
}

func TestIsolatedEnvCleanup(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	mgr.Import("cleanup-test", "")

	env, err := SetupIsolatedEnv("cleanup-test", cfg)
	if err != nil {
		t.Fatalf("SetupIsolatedEnv failed: %v", err)
	}

	tmpDir := env.TempDir

	// Cleanup
	if err := env.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify temp dir was removed
	if DirExists(tmpDir) {
		t.Errorf("TempDir %s still exists after cleanup", tmpDir)
	}
}

func TestIsolatedEnvSharedDirs(t *testing.T) {
	tmpHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create shared dirs in the Claude config
	claudeDir := filepath.Join(tmpHome, ".claude")
	commandsDir := filepath.Join(claudeDir, "commands")
	os.MkdirAll(commandsDir, 0700)
	os.WriteFile(filepath.Join(commandsDir, "test.md"), []byte("# Test command"), 0600)

	settingsFile := filepath.Join(claudeDir, "settings.json")
	os.WriteFile(settingsFile, []byte(`{"key":"value"}`), 0600)

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	mgr.Import("shared-test", "")

	env, err := SetupIsolatedEnv("shared-test", cfg)
	if err != nil {
		t.Fatalf("SetupIsolatedEnv failed: %v", err)
	}
	defer env.Cleanup()

	// Check that symlinks were created for shared dirs
	symlinkPath := filepath.Join(env.TempDir, "commands")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("commands symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("commands is not a symlink")
	}

	// Check settings.json symlink
	settingsSymlink := filepath.Join(env.TempDir, "settings.json")
	info, err = os.Lstat(settingsSymlink)
	if err != nil {
		t.Fatalf("settings.json symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("settings.json is not a symlink")
	}
}
