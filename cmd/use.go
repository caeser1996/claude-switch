package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/profile"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/ui"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a named profile",
	Long: `Use switches Claude Code to the credentials saved in the given profile.

The current credentials are automatically backed up before switching.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.ActiveProfile == name {
			ui.Info("Already using profile %q", name)
			return nil
		}

		mgr := profile.NewManager(cfg)

		if verbose {
			ui.Info("Switching from %q to %q...", cfg.ActiveProfile, name)
		}

		if err := mgr.Use(name); err != nil {
			return err
		}

		ui.Success("Switched to profile %q", name)

		if p, ok := cfg.Profiles[name]; ok && p.Email != "" {
			ui.Info("Email: %s", p.Email)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
