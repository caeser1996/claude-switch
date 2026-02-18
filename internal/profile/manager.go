package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/caeser1996/claude-switch/internal/config"
)

// Manager handles profile CRUD operations.
type Manager struct {
	Config *config.Config
}

// NewManager creates a new profile manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{Config: cfg}
}

// Import saves the current Claude credentials as a named profile.
func (m *Manager) Import(name, description string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Validate name (alphanumeric, hyphens, underscores)
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("invalid profile name %q: use only letters, numbers, hyphens, underscores", name)
		}
	}

	profileDir, err := m.profileDir(name)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(profileDir, 0700); err != nil {
		return fmt.Errorf("cannot create profile directory: %w", err)
	}

	claudeDir, err := config.ClaudeConfigDir()
	if err != nil {
		return err
	}

	// Copy credential files from Claude config dir
	copied := 0
	for _, fname := range CredentialFiles {
		src := filepath.Join(claudeDir, fname)
		if !FileExists(src) {
			continue
		}
		dst := filepath.Join(profileDir, fname)
		if err := CopyFile(src, dst); err != nil {
			return fmt.Errorf("cannot copy %s: %w", fname, err)
		}
		copied++
	}

	// Copy home-level credential files
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	for _, fname := range HomeCredentialFiles {
		src := filepath.Join(home, fname)
		if !FileExists(src) {
			continue
		}
		// Store with a prefix to avoid conflicts
		dst := filepath.Join(profileDir, "home_"+fname)
		if err := CopyFile(src, dst); err != nil {
			return fmt.Errorf("cannot copy %s: %w", fname, err)
		}
		copied++
	}

	if copied == 0 {
		// Cleanup empty dir
		os.Remove(profileDir)
		return fmt.Errorf("no credential files found in %s — is Claude Code installed and logged in?", claudeDir)
	}

	// Try to extract email from credentials
	email := extractEmailFromCredentials(filepath.Join(profileDir, ".credentials.json"))

	// Update config
	isFirst := len(m.Config.Profiles) == 0
	m.Config.Profiles[name] = config.ProfileEntry{
		Name:        name,
		Email:       email,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		IsActive:    isFirst,
	}
	if isFirst {
		m.Config.ActiveProfile = name
	}

	return m.Config.Save()
}

// Use switches to the given profile.
func (m *Manager) Use(name string) error {
	if _, ok := m.Config.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found (use 'cs list' to see available profiles)", name)
	}

	profileDir, err := m.profileDir(name)
	if err != nil {
		return err
	}

	if !DirExists(profileDir) {
		return fmt.Errorf("profile directory for %q is missing — try re-importing", name)
	}

	// Auto-backup current credentials before switching
	if m.Config.Settings.AutoBackup {
		if err := m.createBackup(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	claudeDir, err := config.ClaudeConfigDir()
	if err != nil {
		return err
	}

	// Restore credential files to Claude config dir
	for _, fname := range CredentialFiles {
		src := filepath.Join(profileDir, fname)
		if !FileExists(src) {
			continue
		}
		dst := filepath.Join(claudeDir, fname)
		if err := CopyFile(src, dst); err != nil {
			return fmt.Errorf("cannot restore %s: %w", fname, err)
		}
	}

	// Restore home-level credential files
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	for _, fname := range HomeCredentialFiles {
		src := filepath.Join(profileDir, "home_"+fname)
		if !FileExists(src) {
			continue
		}
		dst := filepath.Join(home, fname)
		if err := CopyFile(src, dst); err != nil {
			return fmt.Errorf("cannot restore %s: %w", fname, err)
		}
	}

	// Update active profile in config
	for k, p := range m.Config.Profiles {
		p.IsActive = (k == name)
		m.Config.Profiles[k] = p
	}
	m.Config.ActiveProfile = name

	return m.Config.Save()
}

// Remove deletes a profile.
func (m *Manager) Remove(name string) error {
	if _, ok := m.Config.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}

	if m.Config.ActiveProfile == name {
		return fmt.Errorf("cannot remove active profile %q — switch to another profile first", name)
	}

	profileDir, err := m.profileDir(name)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("cannot remove profile directory: %w", err)
	}

	delete(m.Config.Profiles, name)
	return m.Config.Save()
}

// List returns all profiles sorted by name.
func (m *Manager) List() []config.ProfileEntry {
	profiles := make([]config.ProfileEntry, 0, len(m.Config.Profiles))
	for _, p := range m.Config.Profiles {
		profiles = append(profiles, p)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	return profiles
}

// Current returns the active profile entry, or an error if none is set.
func (m *Manager) Current() (*config.ProfileEntry, error) {
	if m.Config.ActiveProfile == "" {
		return nil, fmt.Errorf("no active profile — import a profile first with 'cs import <name>'")
	}
	p, ok := m.Config.Profiles[m.Config.ActiveProfile]
	if !ok {
		return nil, fmt.Errorf("active profile %q not found in config", m.Config.ActiveProfile)
	}
	return &p, nil
}

// profileDir returns the full path for a named profile.
func (m *Manager) profileDir(name string) (string, error) {
	base, err := config.ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, name), nil
}

// createBackup saves the current Claude credentials to a timestamped backup dir.
func (m *Manager) createBackup() error {
	backupsDir, err := config.BackupsDir()
	if err != nil {
		return err
	}

	ts := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(backupsDir, ts)
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("cannot create backup directory: %w", err)
	}

	claudeDir, err := config.ClaudeConfigDir()
	if err != nil {
		return err
	}

	copied := false
	for _, fname := range CredentialFiles {
		src := filepath.Join(claudeDir, fname)
		if !FileExists(src) {
			continue
		}
		dst := filepath.Join(backupDir, fname)
		if err := CopyFile(src, dst); err != nil {
			// Non-fatal for backups
			continue
		}
		copied = true
	}

	home, err := os.UserHomeDir()
	if err == nil {
		for _, fname := range HomeCredentialFiles {
			src := filepath.Join(home, fname)
			if !FileExists(src) {
				continue
			}
			dst := filepath.Join(backupDir, "home_"+fname)
			if err := CopyFile(src, dst); err != nil {
				continue
			}
			copied = true
		}
	}

	if !copied {
		// No files to backup, remove the empty dir
		os.Remove(backupDir)
	}

	// Prune old backups
	m.pruneBackups(backupsDir)

	return nil
}

// pruneBackups keeps only the most recent N backups.
func (m *Manager) pruneBackups(backupsDir string) {
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return
	}

	maxBackups := m.Config.Settings.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 10
	}

	if len(entries) <= maxBackups {
		return
	}

	// Entries are sorted by name (timestamp), oldest first
	toRemove := entries[:len(entries)-maxBackups]
	for _, e := range toRemove {
		os.RemoveAll(filepath.Join(backupsDir, e.Name()))
	}
}

// extractEmailFromCredentials tries to read an email from the credentials JSON.
func extractEmailFromCredentials(path string) string {
	// We do a best-effort extraction without pulling in a full JSON parser
	// for the credential structure. If the credentials contain an email field,
	// we'll find it.
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Simple JSON extraction: look for "email":"value"
	type credShape struct {
		Email string `json:"email"`
	}

	// Try unmarshaling as a simple object
	var cred credShape
	if err := jsonUnmarshal(data, &cred); err == nil && cred.Email != "" {
		return cred.Email
	}

	return ""
}
