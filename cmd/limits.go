package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/claude"
	"github.com/caeser1996/claude-switch/internal/config"
)

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Show usage limits for the active profile",
	Long: `Limits displays usage and rate limit information for the currently
active Claude Code profile, extracted from local credential and config files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		profileName := cfg.ActiveProfile
		if profileName == "" {
			return fmt.Errorf("no active profile — import a profile first with 'cs import <name>'")
		}

		info, err := claude.GetUsageInfo()
		if err != nil {
			return err
		}

		fmt.Println(claude.FormatUsageInfo(info, profileName))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(limitsCmd)
}
