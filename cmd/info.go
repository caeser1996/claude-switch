package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/profile"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/ui"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed info about a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		p, ok := cfg.Profiles[name]
		if !ok {
			return fmt.Errorf("profile %q not found", name)
		}

		profilesDir, err := config.ProfilesDir()
		if err != nil {
			return err
		}
		profileDir := filepath.Join(profilesDir, name)

		ui.Header(fmt.Sprintf("Profile: %s", name))
		fmt.Println()

		// Basic info
		fmt.Printf("  %-15s %s\n", "Name:", p.Name)
		fmt.Printf("  %-15s %s\n", "Active:", boolIcon(p.IsActive))

		if p.Email != "" {
			fmt.Printf("  %-15s %s\n", "Email:", p.Email)
		}
		if p.Description != "" {
			fmt.Printf("  %-15s %s\n", "Description:", p.Description)
		}
		if !p.CreatedAt.IsZero() {
			fmt.Printf("  %-15s %s\n", "Created:", p.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
		}

		fmt.Println()

		// Credential files
		ui.Header("Credential Files:")
		for _, fname := range profile.CredentialFiles {
			fpath := filepath.Join(profileDir, fname)
			if profile.FileExists(fpath) {
				info, _ := os.Stat(fpath)
				size := "0 B"
				if info != nil {
					size = formatBytes(info.Size())
				}
				fmt.Printf("  %s  %s (%s)\n", ui.Colorize(ui.Green, "✓"), fname, size)
			} else {
				fmt.Printf("  %s  %s\n", ui.Colorize(ui.Gray, "-"), fname)
			}
		}

		for _, fname := range profile.HomeCredentialFiles {
			stored := "home_" + fname
			fpath := filepath.Join(profileDir, stored)
			if profile.FileExists(fpath) {
				info, _ := os.Stat(fpath)
				size := "0 B"
				if info != nil {
					size = formatBytes(info.Size())
				}
				fmt.Printf("  %s  %s (%s)\n", ui.Colorize(ui.Green, "✓"), stored, size)
			} else {
				fmt.Printf("  %s  %s\n", ui.Colorize(ui.Gray, "-"), stored)
			}
		}

		fmt.Println()

		// Try to extract extra info from credentials
		credPath := filepath.Join(profileDir, ".credentials.json")
		if profile.FileExists(credPath) {
			data, err := os.ReadFile(credPath)
			if err == nil {
				var creds map[string]interface{}
				if json.Unmarshal(data, &creds) == nil {
					ui.Header("Account Details:")
					for _, key := range []string{"plan", "org_name", "orgName", "model"} {
						if v, ok := creds[key]; ok {
							fmt.Printf("  %-15s %v\n", key+":", v)
						}
					}
				}
			}
		}

		return nil
	},
}

func boolIcon(b bool) string {
	if b {
		return ui.Colorize(ui.Green, "yes")
	}
	return "no"
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
