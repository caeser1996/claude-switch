package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show auth status for all profiles",
	Long: `Status checks each profile's credentials and shows whether tokens
are valid, expired, or missing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		profiles := mgr.List()

		if len(profiles) == 0 {
			ui.Info("No profiles saved yet.")
			return nil
		}

		table := ui.NewTable("", "PROFILE", "EMAIL", "STATUS", "EXPIRES")

		for _, p := range profiles {
			status := profile.CheckTokenStatus(p.Name)

			marker := " "
			name := p.Name
			if p.IsActive {
				marker = ui.Colorize(ui.Green, "*")
				name = ui.Colorize(ui.Green, p.Name)
			}

			email := status.Email
			if email == "" {
				email = ui.Colorize(ui.Gray, "-")
			}

			var statusStr, expiresStr string
			if !status.HasCreds {
				statusStr = ui.Colorize(ui.Red, "no credentials")
				expiresStr = ui.Colorize(ui.Gray, "-")
			} else if status.IsExpired {
				statusStr = ui.Colorize(ui.Red, "expired")
				expiresStr = ui.Colorize(ui.Red, "expired")
			} else if status.ExpiresAt != nil {
				statusStr = ui.Colorize(ui.Green, "valid")
				expiresStr = status.ExpiresIn
			} else {
				statusStr = ui.Colorize(ui.Yellow, "unknown")
				expiresStr = ui.Colorize(ui.Gray, "-")
			}

			table.AddRow(marker, name, email, statusStr, expiresStr)
		}

		table.Render()
		fmt.Printf("\n%d profile(s)\n", len(profiles))

		// Token refresh hint
		for _, p := range profiles {
			status := profile.CheckTokenStatus(p.Name)
			if status.IsExpired {
				fmt.Println()
				ui.Warn("Profile %q has expired credentials. Run: cs login %s", p.Name, p.Name)
				break
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
