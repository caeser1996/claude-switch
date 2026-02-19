package profile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/caeser1996/claude-switch/internal/config"
)

// keychainService is the macOS Keychain service name used by Claude Code.
const keychainService = "Claude Code-credentials"

// ReadCurrentCredentials reads the active Claude Code credentials.
// On macOS it reads from the system Keychain; on Linux from the credentials file.
func ReadCurrentCredentials() ([]byte, error) {
	if runtime.GOOS == "darwin" {
		return readKeychainCredentials()
	}
	return readFileCredentials()
}

// WriteCurrentCredentials writes credentials to Claude Code's active storage.
// On macOS it writes to the system Keychain; on Linux to the credentials file.
func WriteCurrentCredentials(data []byte) error {
	if runtime.GOOS == "darwin" {
		return writeKeychainCredentials(data)
	}
	return writeFileCredentials(data)
}

// readKeychainCredentials reads from macOS Keychain.
func readKeychainCredentials() ([]byte, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", keychainService, "-w")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("no credentials found in macOS Keychain (service: %s): %w", keychainService, err)
	}
	return []byte(strings.TrimSpace(string(out))), nil
}

// writeKeychainCredentials writes to macOS Keychain.
func writeKeychainCredentials(data []byte) error {
	user := os.Getenv("USER")
	if user == "" {
		user = "claude-switch"
	}

	// Delete existing entry first (add-generic-password -U can be unreliable).
	_ = exec.Command("security", "delete-generic-password",
		"-s", keychainService).Run()

	cmd := exec.Command("security", "add-generic-password",
		"-s", keychainService,
		"-a", user,
		"-w", string(data))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot write credentials to macOS Keychain: %w", err)
	}
	return nil
}

// readFileCredentials reads from ~/.claude/.credentials.json.
func readFileCredentials() ([]byte, error) {
	claudeDir, err := config.ClaudeConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(claudeDir, ".credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	return data, nil
}

// writeFileCredentials writes to ~/.claude/.credentials.json.
func writeFileCredentials(data []byte) error {
	claudeDir, err := config.ClaudeConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		return fmt.Errorf("cannot create %s: %w", claudeDir, err)
	}
	path := filepath.Join(claudeDir, ".credentials.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	return nil
}

// SaveCredentialsToProfile reads the current credentials and saves them
// to the given profile directory as .credentials.json.
func SaveCredentialsToProfile(profileDir string) error {
	data, err := ReadCurrentCredentials()
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("credentials are empty")
	}
	dst := filepath.Join(profileDir, ".credentials.json")
	return os.WriteFile(dst, data, 0600)
}

// RestoreCredentialsFromProfile reads .credentials.json from a profile
// directory and writes it to the active credential store.
func RestoreCredentialsFromProfile(profileDir string) error {
	src := filepath.Join(profileDir, ".credentials.json")
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("cannot read profile credentials: %w", err)
	}
	return WriteCurrentCredentials(data)
}
