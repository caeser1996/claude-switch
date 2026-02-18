package doctor

import (
	"os"
	"testing"

	"github.com/caeser1996/claude-switch/internal/config"
)

func TestRunAll(t *testing.T) {
	// Use a temp home to isolate from real system
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := config.NewConfig()
	results := RunAll(cfg)

	if len(results) == 0 {
		t.Fatal("RunAll returned no results")
	}

	// Every result should have a name and status
	for _, r := range results {
		if r.Name == "" {
			t.Error("check result has empty name")
		}
		if r.Status != "ok" && r.Status != "warn" && r.Status != "fail" {
			t.Errorf("check %q has invalid status %q", r.Name, r.Status)
		}
		if r.Message == "" {
			t.Errorf("check %q has empty message", r.Name)
		}
	}
}

func TestCheckPlatform(t *testing.T) {
	r := checkPlatform()
	if r.Status != "ok" {
		t.Errorf("platform check should always pass, got %q", r.Status)
	}
	if r.Message == "" {
		t.Error("platform message should not be empty")
	}
}

func TestCheckClaudeHome(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Without .claude dir
	r := checkClaudeHome()
	if r.Status != "fail" {
		t.Errorf("expected fail when .claude dir missing, got %q", r.Status)
	}

	// Create .claude dir
	if err := os.MkdirAll(tmpDir+"/.claude", 0700); err != nil {
		t.Fatalf("cannot create .claude dir: %v", err)
	}
	r = checkClaudeHome()
	if r.Status != "ok" {
		t.Errorf("expected ok when .claude dir exists, got %q", r.Status)
	}
}

func TestCheckCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Without credentials
	r := checkCredentials()
	if r.Status != "warn" {
		t.Errorf("expected warn when credentials missing, got %q", r.Status)
	}

	// Create credentials
	claudeDir := tmpDir + "/.claude"
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatalf("cannot create .claude dir: %v", err)
	}
	if err := os.WriteFile(claudeDir+"/.credentials.json", []byte("{}"), 0600); err != nil {
		t.Fatalf("cannot write credentials: %v", err)
	}
	r = checkCredentials()
	if r.Status != "ok" {
		t.Errorf("expected ok when credentials exist, got %q", r.Status)
	}
}

func TestCheckConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Without config
	r := checkConfigFile()
	if r.Status != "warn" {
		t.Errorf("expected warn when config missing, got %q", r.Status)
	}

	// Create valid config
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}
	cfg := config.NewConfig()
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	r = checkConfigFile()
	if r.Status != "ok" {
		t.Errorf("expected ok when valid config exists, got %q", r.Status)
	}

	// Create invalid config
	cfgPath, _ := config.ConfigPath()
	if err := os.WriteFile(cfgPath, []byte("not json!"), 0600); err != nil {
		t.Fatalf("cannot write invalid config: %v", err)
	}
	r = checkConfigFile()
	if r.Status != "fail" {
		t.Errorf("expected fail when config is invalid JSON, got %q", r.Status)
	}
}

func TestCheckProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// No profiles
	cfg := config.NewConfig()
	r := checkProfiles(cfg)
	if r.Status != "warn" {
		t.Errorf("expected warn when no profiles, got %q", r.Status)
	}

	// Add a profile with its directory
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}
	cfg.Profiles["test"] = config.ProfileEntry{Name: "test"}
	profilesDir, _ := config.ProfilesDir()
	if err := os.MkdirAll(profilesDir+"/test", 0700); err != nil {
		t.Fatalf("cannot create profile dir: %v", err)
	}

	r = checkProfiles(cfg)
	if r.Status != "ok" {
		t.Errorf("expected ok when profile dirs exist, got %q", r.Status)
	}

	// Add a profile without directory
	cfg.Profiles["missing"] = config.ProfileEntry{Name: "missing"}
	r = checkProfiles(cfg)
	if r.Status != "warn" {
		t.Errorf("expected warn when profile dir missing, got %q", r.Status)
	}
}

func TestCheckPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// App dir doesn't exist yet
	r := checkPermissions()
	if r.Status != "ok" {
		t.Errorf("expected ok when app dir not yet created, got %q", r.Status)
	}

	// Create with proper permissions
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}
	r = checkPermissions()
	if r.Status != "ok" {
		t.Errorf("expected ok with proper permissions, got %q", r.Status)
	}
}
