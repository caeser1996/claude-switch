package cmd

import (
	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks on your installation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewConfig()
		}

		results := doctor.RunAll(cfg)
		doctor.PrintResults(results)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
