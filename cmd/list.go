package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/profile"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/ui"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all saved profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		profiles := mgr.List()

		if len(profiles) == 0 {
			ui.Info("No profiles saved yet. Use 'cs import <name>' to save your current session.")
			return nil
		}

		table := ui.NewTable("", "NAME", "EMAIL", "DESCRIPTION")
		for _, p := range profiles {
			marker := " "
			name := p.Name
			if p.IsActive {
				marker = ui.Colorize(ui.Green, "*")
				name = ui.Colorize(ui.Green, p.Name)
			}

			email := p.Email
			if email == "" {
				email = ui.Colorize(ui.Gray, "-")
			}

			desc := p.Description
			if desc == "" {
				desc = ui.Colorize(ui.Gray, "-")
			}

			table.AddRow(marker, name, email, desc)
		}

		table.Render()
		fmt.Printf("\n%d profile(s)\n", len(profiles))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
