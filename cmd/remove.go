package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var forceRemove bool

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Delete a saved profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("profile %q not found", name)
		}

		if !forceRemove {
			ui.Warn("This will permanently delete profile %q. Use --force to confirm.", name)
			return nil
		}

		mgr := profile.NewManager(cfg)
		if err := mgr.Remove(name); err != nil {
			return err
		}

		ui.Success("Profile %q removed", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Skip confirmation")
	rootCmd.AddCommand(removeCmd)
}
