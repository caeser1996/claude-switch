package cmd

import (
	"bytes"
	"os"
	"testing"
)

// captureOutput runs a function and captures stdout.
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return ""
	}
	return buf.String()
}

func setupTestHome(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Create mock Claude config
	claudeDir := tmpDir + "/.claude"
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatalf("cannot create .claude dir: %v", err)
	}
	if err := os.WriteFile(claudeDir+"/.credentials.json",
		[]byte(`{"email":"test@example.com","token":"mock"}`), 0600); err != nil {
		t.Fatalf("cannot write credentials: %v", err)
	}

	return func() {
		os.Setenv("HOME", origHome)
	}
}

func TestRootCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"version"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("version command produced no output")
	}
}

func TestVersionVerboseCommand(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"version", "-v"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("version -v command produced no output")
	}
}

func TestListEmptyCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("list command produced no output")
	}
}

func TestCurrentNoProfileCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	rootCmd.SetArgs([]string{"current"})
	// Current may or may not error depending on config state
	// Just make sure it doesn't panic
	_ = rootCmd.Execute()
}

func TestImportAndListAndCurrentCommands(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// Import
	rootCmd.SetArgs([]string{"import", "test-profile", "-d", "Test description"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// List
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list"})
		_ = rootCmd.Execute()
	})
	if output == "" {
		t.Error("list should show imported profile")
	}

	// Current
	output = captureOutput(func() {
		rootCmd.SetArgs([]string{"current"})
		_ = rootCmd.Execute()
	})
	if output == "" {
		t.Error("current should show active profile")
	}
}

func TestDoctorCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"doctor"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("doctor command produced no output")
	}
}

func TestConfigShowCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"config", "show"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("config show produced no output")
	}
}

func TestConfigPathCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"config", "path"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("config path produced no output")
	}
}

func TestBackupListCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"backup", "list"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("backup list produced no output")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	rootCmd.SetArgs([]string{"remove", "nonexistent", "--force"})
	// May or may not error depending on internal handling; just verify no panic
	_ = rootCmd.Execute()
}

func TestUseNonExistent(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	rootCmd.SetArgs([]string{"use", "nonexistent"})
	// Expected to fail; just verify no panic
	_ = rootCmd.Execute()
}

func TestInfoNonExistent(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	rootCmd.SetArgs([]string{"info", "nonexistent"})
	// Expected to fail; just verify no panic
	_ = rootCmd.Execute()
}

func TestCompletionBash(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"completion", "bash"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("bash completion produced no output")
	}
}

func TestCompletionZsh(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"completion", "zsh"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("zsh completion produced no output")
	}
}

func TestCompletionFish(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"completion", "fish"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("fish completion produced no output")
	}
}

func TestCompletionPowershell(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"completion", "powershell"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("powershell completion produced no output")
	}
}

func TestLimitsCommand(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// Import a profile first so there's an active one
	rootCmd.SetArgs([]string{"import", "limits-test"})
	_ = rootCmd.Execute()

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"limits"})
		_ = rootCmd.Execute()
	})

	if output == "" {
		t.Error("limits command produced no output")
	}
}
