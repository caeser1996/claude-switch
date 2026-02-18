package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		p, err := mgr.Current()
		if err != nil {
			return err
		}

		fmt.Println(ui.Colorize(ui.Green, p.Name))

		if verbose {
			if p.Email != "" {
				fmt.Printf("  Email:       %s\n", p.Email)
			}
			if p.Description != "" {
				fmt.Printf("  Description: %s\n", p.Description)
			}
			fmt.Printf("  Created:     %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
