package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/ui"
	"github.com/caeser1996/claude-switch/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update to the latest release",
	Long:  `Update checks GitHub for the latest release and replaces the current binary.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Info("Checking for updates...")

		result, err := update.Check(Version)
		if err != nil {
			return err
		}

		if !result.UpdateNeeded {
			ui.Success("Already up to date (v%s)", result.CurrentVersion)
			return nil
		}

		fmt.Printf("  Current version: %s\n", result.CurrentVersion)
		fmt.Printf("  Latest version:  %s\n", result.LatestVersion)
		fmt.Println()

		if result.DownloadURL == "" {
			ui.Warn("No binary found for your platform")
			ui.Info("Download manually: %s", result.ReleaseURL)
			return nil
		}

		if !ui.Confirm("Update to v" + result.LatestVersion + "?") {
			ui.Info("Update cancelled")
			return nil
		}

		ui.Info("Downloading %s...", result.AssetName)
		if err := update.Apply(result.DownloadURL); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		ui.Success("Updated to v%s", result.LatestVersion)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
