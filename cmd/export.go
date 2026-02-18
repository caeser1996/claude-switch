package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export a profile as an encrypted file",
	Long: `Export packages a profile's credentials into an encrypted file
that can be shared with team members or transferred between machines.

The file is encrypted with AES-256-GCM using a passphrase you provide.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("profile %q not found", name)
		}

		passphrase, err := ui.ReadPassword("Enter passphrase for encryption: ")
		if err != nil {
			return err
		}
		if len(passphrase) < 8 {
			return fmt.Errorf("passphrase must be at least 8 characters")
		}

		confirm, err := ui.ReadPassword("Confirm passphrase: ")
		if err != nil {
			return err
		}
		if passphrase != confirm {
			return fmt.Errorf("passphrases do not match")
		}

		data, err := profile.ExportProfile(name, cfg, passphrase)
		if err != nil {
			return err
		}

		outFile := exportOutput
		if outFile == "" {
			outFile = name + ".csprofile"
		}

		if err := os.WriteFile(outFile, data, 0600); err != nil {
			return fmt.Errorf("cannot write file: %w", err)
		}

		ui.Success("Profile %q exported to %s", name, outFile)
		ui.Info("Share this file and the passphrase separately")
		return nil
	},
}

var importFileCmd = &cobra.Command{
	Use:   "import-file <file>",
	Short: "Import a profile from an encrypted file",
	Long: `Import-file decrypts and imports a profile from an exported .csprofile file.

Use the passphrase that was set during export.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}

		passphrase, err := ui.ReadPassword("Enter passphrase: ")
		if err != nil {
			return err
		}

		if err := config.EnsureDirs(); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		name, err := profile.ImportFromFile(data, passphrase, "", cfg)
		if err != nil {
			return err
		}

		ui.Success("Profile %q imported from %s", name, filePath)
		if p, ok := cfg.Profiles[name]; ok && p.Email != "" {
			ui.Info("Email: %s", p.Email)
		}

		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: <name>.csprofile)")
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importFileCmd)
}
