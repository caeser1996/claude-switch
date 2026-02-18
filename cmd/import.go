package cmd

import (
	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var importDescription string

var importCmd = &cobra.Command{
	Use:   "import <name>",
	Short: "Save current Claude session as a named profile",
	Long: `Import saves the currently active Claude Code credentials as a named profile.

Before importing, make sure you're logged into the Claude account you want to save.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if err := config.EnsureDirs(); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Check if profile already exists
		if _, exists := cfg.Profiles[name]; exists {
			ui.Warn("Profile %q already exists — it will be overwritten", name)
		}

		mgr := profile.NewManager(cfg)
		if err := mgr.Import(name, importDescription); err != nil {
			return err
		}

		ui.Success("Profile %q imported successfully", name)
		if p, ok := cfg.Profiles[name]; ok && p.Email != "" {
			ui.Info("Email: %s", p.Email)
		}

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importDescription, "description", "d", "", "Description for this profile")
	rootCmd.AddCommand(importCmd)
}
