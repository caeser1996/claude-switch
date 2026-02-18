package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	AppName    = "claude-switch"
	AppDir     = ".claude-switch"
	ConfigFile = "config.json"
)

// ProfileEntry holds metadata about a saved profile.
type ProfileEntry struct {
	Name        string    `json:"name"`
	Email       string    `json:"email,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// Settings holds application-level settings.
type Settings struct {
	AutoBackup  bool `json:"auto_backup"`
	MaxBackups  int  `json:"max_backups"`
	ColorOutput bool `json:"color_output"`
}

// Config is the top-level configuration.
type Config struct {
	ActiveProfile string                  `json:"active_profile"`
	Profiles      map[string]ProfileEntry `json:"profiles"`
	Settings      Settings                `json:"settings"`
}

// DefaultSettings returns sensible defaults.
func DefaultSettings() Settings {
	return Settings{
		AutoBackup:  true,
		MaxBackups:  10,
		ColorOutput: true,
	}
}

// NewConfig creates a fresh configuration.
func NewConfig() *Config {
	return &Config{
		Profiles: make(map[string]ProfileEntry),
		Settings: DefaultSettings(),
	}
}

// AppDataDir returns the path to ~/.claude-switch/
func AppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, AppDir), nil
}

// ProfilesDir returns the path to ~/.claude-switch/profiles/
func ProfilesDir() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "profiles"), nil
}

// BackupsDir returns the path to ~/.claude-switch/backups/
func BackupsDir() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "backups"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ConfigFile), nil
}

// ClaudeConfigDir returns the path to Claude's config directory.
// On all platforms this is ~/.claude/
func ClaudeConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

// ClaudeCredentialsPath returns the path to Claude's credentials file.
func ClaudeCredentialsPath() (string, error) {
	dir, err := ClaudeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".credentials.json"), nil
}

// ClaudeAuthJSONPath returns the path to the legacy ~/.claude.json auth file.
func ClaudeAuthJSONPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude.json"), nil
}

// EnsureDirs creates the app directory structure if it doesn't exist.
func EnsureDirs() error {
	dirs := []func() (string, error){AppDataDir, ProfilesDir, BackupsDir}
	perm := os.FileMode(0700)
	for _, fn := range dirs {
		dir, err := fn()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, perm); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}
	return nil
}

// Load reads the config from disk, or returns a new config if the file doesn't exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	cfg := NewConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk with secure permissions.
func (c *Config) Save() error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}
	return nil
}

// IsWindows returns true when running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}
