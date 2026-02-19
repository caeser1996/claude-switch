package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/claude"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var loginDescription string

var loginCmd = &cobra.Command{
	Use:   "login <name>",
	Short: "Log into Claude and save as a named profile",
	Long: `Login opens Claude Code to trigger its authentication flow, then
automatically imports the resulting credentials as a named profile.

After authenticating in your browser, exit Claude Code (/exit or Ctrl+C)
and the credentials will be saved.

This is a convenience shortcut for:
  claude   (complete login)
  cs import <name>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Create a temporary config dir so Claude starts fresh (no existing creds)
		// and triggers its OAuth flow.
		tmpDir, err := os.MkdirTemp("", "cs-login-*")
		if err != nil {
			return fmt.Errorf("cannot create temp directory: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		// Symlink shared settings from the real Claude config so the user
		// gets their usual settings during the login session.
		claudeDir, err := config.ClaudeConfigDir()
		if err == nil {
			for _, f := range []string{"settings.json", "settings.local.json"} {
				src := filepath.Join(claudeDir, f)
				if profile.FileExists(src) {
					_ = os.Symlink(src, filepath.Join(tmpDir, f))
				}
			}
		}

		ui.Info("Starting Claude login...")
		ui.Info("Complete the login in your browser.")
		ui.Info("After authenticating, exit Claude Code (/exit or Ctrl+C) to continue.")
		fmt.Println()

		// Run claude with the empty config dir — this triggers the OAuth flow.
		err = claude.Run(claude.RunOptions{
			Args: []string{},
			Env:  buildLoginEnv(tmpDir),
		})
		if err != nil {
			return fmt.Errorf("claude login failed: %w", err)
		}

		fmt.Println()

		// Check if credentials appeared in the temp dir.
		credsPath := filepath.Join(tmpDir, ".credentials.json")
		if !profile.FileExists(credsPath) {
			return fmt.Errorf("no credentials found after login — authentication may not have completed")
		}

		ui.Info("Login complete. Importing as profile %q...", name)

		if err := config.EnsureDirs(); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		if err := mgr.ImportFromDir(name, loginDescription, tmpDir); err != nil {
			return err
		}

		ui.Success("Profile %q created from login", name)
		if p, ok := cfg.Profiles[name]; ok && p.Email != "" {
			ui.Info("Email: %s", p.Email)
		}

		return nil
	},
}

// buildLoginEnv returns the current environment with CLAUDE_CONFIG_DIR
// set to the given temporary directory.
func buildLoginEnv(tmpDir string) []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, "CLAUDE_CONFIG_DIR=") {
			continue
		}
		filtered = append(filtered, e)
	}
	filtered = append(filtered, "CLAUDE_CONFIG_DIR="+tmpDir)
	return filtered
}

func init() {
	loginCmd.Flags().StringVarP(&loginDescription, "description", "d", "", "Description for this profile")
	rootCmd.AddCommand(loginCmd)
}
