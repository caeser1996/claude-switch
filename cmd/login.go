package cmd

import (
	"fmt"

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
	Long: `Login runs 'claude login' interactively, then automatically imports
the resulting credentials as a named profile.

This is a convenience shortcut for:
  claude login
  cs import <name>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		ui.Info("Starting Claude login...")
		ui.Info("Complete the login in your browser, then return here.")
		fmt.Println()

		// Run claude login interactively
		err := claude.Run(claude.RunOptions{
			Args: []string{"login"},
		})
		if err != nil {
			return fmt.Errorf("claude login failed: %w", err)
		}

		fmt.Println()
		ui.Info("Login complete. Importing as profile %q...", name)

		// Now import
		if err := config.EnsureDirs(); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		if err := mgr.Import(name, loginDescription); err != nil {
			return err
		}

		ui.Success("Profile %q created from login", name)
		if p, ok := cfg.Profiles[name]; ok && p.Email != "" {
			ui.Info("Email: %s", p.Email)
		}

		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginDescription, "description", "d", "", "Description for this profile")
	rootCmd.AddCommand(loginCmd)
}
