package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/crypto"
)

// ExportBundle contains all profile data for export.
type ExportBundle struct {
	Version     int               `json:"version"`
	ProfileName string            `json:"profile_name"`
	Email       string            `json:"email,omitempty"`
	Description string            `json:"description,omitempty"`
	Files       map[string][]byte `json:"files"`
}

// ExportProfile packages a profile into an encrypted bundle.
func ExportProfile(name string, cfg *config.Config, passphrase string) ([]byte, error) {
	entry, ok := cfg.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return nil, err
	}
	profileDir := filepath.Join(profilesDir, name)

	if !DirExists(profileDir) {
		return nil, fmt.Errorf("profile directory for %q is missing", name)
	}

	bundle := ExportBundle{
		Version:     1,
		ProfileName: name,
		Email:       entry.Email,
		Description: entry.Description,
		Files:       make(map[string][]byte),
	}

	// Collect all credential files
	allFiles := append([]string{}, CredentialFiles...)
	for _, f := range HomeCredentialFiles {
		allFiles = append(allFiles, "home_"+f)
	}

	for _, fname := range allFiles {
		fpath := filepath.Join(profileDir, fname)
		if !FileExists(fpath) {
			continue
		}
		data, err := os.ReadFile(fpath)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", fname, err)
		}
		bundle.Files[fname] = data
	}

	if len(bundle.Files) == 0 {
		return nil, fmt.Errorf("profile %q has no credential files to export", name)
	}

	// Serialize bundle
	plaintext, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("cannot serialize bundle: %w", err)
	}

	// Encrypt
	return crypto.Encrypt(plaintext, passphrase)
}

// ImportFromFile decrypts and imports a profile bundle.
func ImportFromFile(data []byte, passphrase string, overrideName string, cfg *config.Config) (string, error) {
	// Decrypt
	plaintext, err := crypto.Decrypt(data, passphrase)
	if err != nil {
		return "", err
	}

	var bundle ExportBundle
	if err := json.Unmarshal(plaintext, &bundle); err != nil {
		return "", fmt.Errorf("invalid profile bundle: %w", err)
	}

	name := bundle.ProfileName
	if overrideName != "" {
		name = overrideName
	}

	if name == "" {
		return "", fmt.Errorf("profile name is empty in bundle")
	}

	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return "", err
	}
	profileDir := filepath.Join(profilesDir, name)

	if err := os.MkdirAll(profileDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create profile directory: %w", err)
	}

	// Write credential files
	for fname, content := range bundle.Files {
		fpath := filepath.Join(profileDir, fname)
		if err := os.WriteFile(fpath, content, 0600); err != nil {
			return "", fmt.Errorf("cannot write %s: %w", fname, err)
		}
	}

	// Update config
	isFirst := len(cfg.Profiles) == 0
	entry := config.ProfileEntry{
		Name:        name,
		Email:       bundle.Email,
		Description: bundle.Description,
		IsActive:    isFirst,
	}
	cfg.Profiles[name] = entry
	if isFirst {
		cfg.ActiveProfile = name
	}

	if err := cfg.Save(); err != nil {
		return "", err
	}

	return name, nil
}
