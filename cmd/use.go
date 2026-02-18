package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch to a named profile",
	Long: `Use switches Claude Code to the credentials saved in the given profile.

The current credentials are automatically backed up before switching.

If no name is given and a .claude-profile file exists in the current
directory (or any parent), that profile is used automatically.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			// Try .claude-profile detection
			name = profile.DetectProjectProfile()
			if name == "" {
				return cmd.Help()
			}
			ui.Info("Detected .claude-profile: %s", name)
		}

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

		// Token health check after switch
		status := profile.CheckTokenStatus(name)
		if status.IsExpired {
			ui.Warn("Token for %q appears expired. Run: cs login %s", name, name)
		} else if profile.NeedsRefresh(status, 24*time.Hour) {
			ui.Warn("Token for %q expires soon (%s). Consider: cs login %s", name, status.ExpiresIn, name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
