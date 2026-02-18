package cmd

import (
	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/claude"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var execCmd = &cobra.Command{
	Use:   "exec <profile> [--] <command> [args...]",
	Short: "Run a command with a specific profile's credentials",
	Long: `Exec runs the claude binary using the specified profile's credentials
in an isolated environment, without switching your active profile.

The current profile remains unchanged. Shared resources like commands,
skills, and settings are symlinked into the isolated environment.

Use -- to separate the profile name from the command arguments.`,
	Args:               cobra.MinimumNArgs(2),
	DisableFlagParsing: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		cmdArgs := args[1:]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if verbose {
			ui.Info("Setting up isolated environment for profile %q...", profileName)
		}

		env, err := profile.SetupIsolatedEnv(profileName, cfg)
		if err != nil {
			return err
		}
		defer func() {
			if verbose {
				ui.Info("Cleaning up isolated environment...")
			}
			_ = env.Cleanup()
		}()

		if verbose {
			ui.Info("Running: claude %v", cmdArgs)
			ui.Info("CLAUDE_CONFIG_DIR=%s", env.TempDir)
		}

		return claude.Run(claude.RunOptions{
			Args: cmdArgs,
			Env:  env.Env(),
		})
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
