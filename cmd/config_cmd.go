package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or edit configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("cannot serialize config: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config file exists
		if err := config.EnsureDirs(); err != nil {
			return err
		}

		cfgPath, err := config.ConfigPath()
		if err != nil {
			return err
		}

		// Create config if it doesn't exist
		if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
			cfg := config.NewConfig()
			if err := cfg.Save(); err != nil {
				return err
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Try common editors
			for _, e := range []string{"nano", "vim", "vi"} {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}
		if editor == "" {
			return fmt.Errorf("no editor found — set $EDITOR environment variable")
		}

		if verbose {
			ui.Info("Opening %s with %s", cfgPath, editor)
		}

		c := exec.Command(editor, cfgPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
