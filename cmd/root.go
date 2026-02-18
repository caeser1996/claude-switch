package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var (
	verbose bool
	noColor bool

	// Version is set at build time via ldflags.
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// rootCmd is the base command for claude-switch.
var rootCmd = &cobra.Command{
	Use:   "cs",
	Short: "Claude Switch — manage multiple Claude Code accounts",
	Long: `Claude Switch (cs) is a CLI tool for managing multiple Claude Code accounts.

Save, switch, and manage profiles for different Claude accounts with a single command.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			ui.SetColorEnabled(false)
		}
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.Error("%s", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}
