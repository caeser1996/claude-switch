package cmd

import (
	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

// switchCmd is the TUI interactive profile selector.
var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Interactively select and switch to a profile",
	Long: `Switch presents an interactive numbered list of all profiles
and lets you pick one to switch to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		profiles := mgr.List()

		if len(profiles) == 0 {
			ui.Info("No profiles saved yet. Use 'cs import <name>' to create one.")
			return nil
		}

		options := make([]ui.SelectOption, 0, len(profiles))
		for _, p := range profiles {
			desc := ""
			if p.Email != "" {
				desc = p.Email
			}
			if p.IsActive {
				desc += " (active)"
			}
			options = append(options, ui.SelectOption{
				Label:       p.Name,
				Description: desc,
				Value:       p.Name,
			})
		}

		selected, err := ui.Select("Select a profile to switch to:", options)
		if err != nil {
			return err
		}

		if cfg.ActiveProfile == selected {
			ui.Info("Already using profile %q", selected)
			return nil
		}

		if err := mgr.Use(selected); err != nil {
			return err
		}

		ui.Success("Switched to profile %q", selected)

		// Token health warning
		status := profile.CheckTokenStatus(selected)
		if status.IsExpired {
			ui.Warn("This profile's token appears expired. Run: cs login %s", selected)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
