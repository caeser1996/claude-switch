package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage credential backups",
	Long:  `Backup provides subcommands to list and restore automatic backups.`,
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show available backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		backupsDir, err := config.BackupsDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(backupsDir)
		if err != nil {
			if os.IsNotExist(err) {
				ui.Info("No backups found.")
				return nil
			}
			return err
		}

		// Filter to only directories
		var dirs []os.DirEntry
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, e)
			}
		}

		if len(dirs) == 0 {
			ui.Info("No backups found.")
			return nil
		}

		// Sort by name (timestamp) descending (most recent first)
		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].Name() > dirs[j].Name()
		})

		table := ui.NewTable("#", "TIMESTAMP", "FILES")
		for i, d := range dirs {
			// Count files in backup
			files, _ := os.ReadDir(filepath.Join(backupsDir, d.Name()))
			fileNames := make([]string, 0, len(files))
			for _, f := range files {
				if !f.IsDir() {
					fileNames = append(fileNames, f.Name())
				}
			}

			table.AddRow(
				fmt.Sprintf("%d", i+1),
				d.Name(),
				strings.Join(fileNames, ", "),
			)
		}

		table.Render()
		fmt.Printf("\n%d backup(s)\n", len(dirs))

		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <timestamp>",
	Short: "Restore credentials from a backup",
	Long: `Restore replaces the current Claude credentials with those from
the specified backup. The timestamp should match one from 'cs backup list'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp := args[0]

		backupsDir, err := config.BackupsDir()
		if err != nil {
			return err
		}

		backupDir := filepath.Join(backupsDir, timestamp)
		if !profile.DirExists(backupDir) {
			return fmt.Errorf("backup %q not found — use 'cs backup list' to see available backups", timestamp)
		}

		claudeDir, err := config.ClaudeConfigDir()
		if err != nil {
			return err
		}

		// Restore credential files
		restored := 0
		for _, fname := range profile.CredentialFiles {
			src := filepath.Join(backupDir, fname)
			if !profile.FileExists(src) {
				continue
			}
			dst := filepath.Join(claudeDir, fname)
			if err := profile.CopyFile(src, dst); err != nil {
				return fmt.Errorf("cannot restore %s: %w", fname, err)
			}
			restored++
			if verbose {
				ui.Info("Restored %s", fname)
			}
		}

		// Restore home-level files
		home, err := os.UserHomeDir()
		if err == nil {
			for _, fname := range profile.HomeCredentialFiles {
				src := filepath.Join(backupDir, "home_"+fname)
				if !profile.FileExists(src) {
					continue
				}
				dst := filepath.Join(home, fname)
				if err := profile.CopyFile(src, dst); err != nil {
					return fmt.Errorf("cannot restore %s: %w", fname, err)
				}
				restored++
				if verbose {
					ui.Info("Restored %s", fname)
				}
			}
		}

		if restored == 0 {
			ui.Warn("Backup %q contained no credential files", timestamp)
			return nil
		}

		ui.Success("Restored %d file(s) from backup %s", restored, timestamp)
		ui.Warn("Active profile in config may be out of sync — consider running 'cs import' to re-save")

		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	rootCmd.AddCommand(backupCmd)
}
