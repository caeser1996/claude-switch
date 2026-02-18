package profile

import (
	"testing"

	"github.com/caeser1996/claude-switch/internal/config"
)

func TestExportImportRoundTrip(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)

	// Import a profile
	if err := mgr.Import("export-test", "Test profile for export"); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Export it
	passphrase := "test-passphrase-12345678"
	data, err := ExportProfile("export-test", cfg, passphrase)
	if err != nil {
		t.Fatalf("ExportProfile failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("exported data is empty")
	}

	// Clear profiles and re-import from file
	cfg2 := config.NewConfig()
	name, err := ImportFromFile(data, passphrase, "", cfg2)
	if err != nil {
		t.Fatalf("ImportFromFile failed: %v", err)
	}

	if name != "export-test" {
		t.Errorf("expected name 'export-test', got %q", name)
	}

	if _, ok := cfg2.Profiles["export-test"]; !ok {
		t.Error("profile not found in config after import")
	}
}

func TestExportImportWithOverrideName(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	if err := mgr.Import("original", "Original"); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	passphrase := "test-passphrase-12345678"
	data, err := ExportProfile("original", cfg, passphrase)
	if err != nil {
		t.Fatalf("ExportProfile failed: %v", err)
	}

	cfg2 := config.NewConfig()
	name, err := ImportFromFile(data, passphrase, "renamed", cfg2)
	if err != nil {
		t.Fatalf("ImportFromFile with override name failed: %v", err)
	}

	if name != "renamed" {
		t.Errorf("expected name 'renamed', got %q", name)
	}
}

func TestExportWrongPassphrase(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	if err := mgr.Import("wrong-pass", ""); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	data, err := ExportProfile("wrong-pass", cfg, "correct-passphrase")
	if err != nil {
		t.Fatalf("ExportProfile failed: %v", err)
	}

	cfg2 := config.NewConfig()
	_, err = ImportFromFile(data, "wrong-passphrase", "", cfg2)
	if err == nil {
		t.Error("expected error with wrong passphrase")
	}
}

func TestExportNonExistentProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	_, err := ExportProfile("nonexistent", cfg, "passphrase")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}
