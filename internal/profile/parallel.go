package profile

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/caeser1996/claude-switch/internal/config"
)

// SharedDirs are directories from the Claude config that should be symlinked
// (not copied) during parallel execution, so commands/skills are shared.
var SharedDirs = []string{
	"commands",
	"todos",
}

// SharedFiles are individual files to symlink during parallel exec.
var SharedFiles = []string{
	"settings.json",
	"settings.local.json",
}

// IsolatedEnv represents a temporary isolated environment for running
// claude with a specific profile's credentials.
type IsolatedEnv struct {
	ProfileName string
	TempDir     string
	OrigEnv     map[string]string
}

// SetupIsolatedEnv creates a temporary directory with the profile's credentials
// and symlinks to shared resources (commands, skills, settings).
func SetupIsolatedEnv(profileName string, cfg *config.Config) (*IsolatedEnv, error) {
	if _, ok := cfg.Profiles[profileName]; !ok {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}

	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return nil, err
	}
	profileDir := filepath.Join(profilesDir, profileName)

	if !DirExists(profileDir) {
		return nil, fmt.Errorf("profile directory for %q is missing", profileName)
	}

	// Create temp dir
	suffix := fmt.Sprintf("cs-%s-%08x", profileName, rand.Int31())
	tmpDir, err := os.MkdirTemp("", suffix)
	if err != nil {
		return nil, fmt.Errorf("cannot create temp directory: %w", err)
	}

	// Copy credential files from profile to temp dir
	for _, fname := range CredentialFiles {
		src := filepath.Join(profileDir, fname)
		if !FileExists(src) {
			continue
		}
		dst := filepath.Join(tmpDir, fname)
		if err := CopyFile(src, dst); err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("cannot copy %s: %w", fname, err)
		}
	}

	// Symlink shared directories from the real Claude config
	claudeDir, err := config.ClaudeConfigDir()
	if err == nil {
		for _, dirName := range SharedDirs {
			src := filepath.Join(claudeDir, dirName)
			if DirExists(src) {
				dst := filepath.Join(tmpDir, dirName)
				_ = os.Symlink(src, dst) // best-effort
			}
		}

		for _, fileName := range SharedFiles {
			src := filepath.Join(claudeDir, fileName)
			if FileExists(src) {
				dst := filepath.Join(tmpDir, fileName)
				_ = os.Symlink(src, dst) // best-effort
			}
		}
	}

	env := &IsolatedEnv{
		ProfileName: profileName,
		TempDir:     tmpDir,
		OrigEnv:     make(map[string]string),
	}

	return env, nil
}

// Env returns the environment variables to set for isolated execution.
func (e *IsolatedEnv) Env() []string {
	result := os.Environ()

	// Filter out existing CLAUDE_CONFIG_DIR if present, then add ours
	filtered := make([]string, 0, len(result)+1)
	for _, env := range result {
		if !strings.HasPrefix(env, "CLAUDE_CONFIG_DIR=") {
			filtered = append(filtered, env)
		}
	}
	filtered = append(filtered, "CLAUDE_CONFIG_DIR="+e.TempDir)

	return filtered
}

// Cleanup removes the temporary directory.
func (e *IsolatedEnv) Cleanup() error {
	if e.TempDir == "" {
		return nil
	}
	return os.RemoveAll(e.TempDir)
}
