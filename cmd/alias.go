package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
	"github.com/caeser1996/claude-switch/internal/ui"
)

var aliasShell string

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Generate shell aliases for each profile",
	Long: `Alias generates shell alias commands that create shortcuts like
'claude-work' and 'claude-personal' which automatically switch profiles
and launch claude.

Add the output to your shell profile (.bashrc, .zshrc, etc.):
  cs alias >> ~/.bashrc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		mgr := profile.NewManager(cfg)
		profiles := mgr.List()

		if len(profiles) == 0 {
			ui.Info("No profiles to create aliases for.")
			return nil
		}

		shell := aliasShell
		if shell == "" {
			shell = "bash"
		}

		switch shell {
		case "bash", "zsh":
			printBashAliases(profiles)
		case "fish":
			printFishAliases(profiles)
		case "powershell", "ps":
			printPowerShellAliases(profiles)
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, fish, or powershell)", shell)
		}

		return nil
	},
}

func printBashAliases(profiles []config.ProfileEntry) {
	fmt.Println("# Claude Switch aliases")
	fmt.Println("# Add these to your ~/.bashrc or ~/.zshrc")
	fmt.Println()

	for _, p := range profiles {
		safeName := strings.ReplaceAll(p.Name, "-", "_")
		// Switch alias: cs-<name> switches profile
		fmt.Printf("alias cs-%s='claude-switch use %s'\n", p.Name, p.Name)
		// Claude alias: claude-<name> switches and runs claude
		fmt.Printf("alias claude-%s='claude-switch use %s && claude'\n", p.Name, p.Name)
		// Exec alias: claude-<name>-exec runs without switching
		fmt.Printf("alias claude-%s-exec='claude-switch exec %s --'\n", p.Name, p.Name)
		_ = safeName
	}

	fmt.Println()
	fmt.Println("# Quick switch function")
	fmt.Println("cs() {")
	fmt.Println("  if [ $# -eq 0 ]; then")
	fmt.Println("    claude-switch list")
	fmt.Println("  else")
	fmt.Println("    claude-switch \"$@\"")
	fmt.Println("  fi")
	fmt.Println("}")
}

func printFishAliases(profiles []config.ProfileEntry) {
	fmt.Println("# Claude Switch aliases for Fish")
	fmt.Println("# Add these to your ~/.config/fish/config.fish")
	fmt.Println()

	for _, p := range profiles {
		fmt.Printf("alias cs-%s 'claude-switch use %s'\n", p.Name, p.Name)
		fmt.Printf("alias claude-%s 'claude-switch use %s; and claude'\n", p.Name, p.Name)
		fmt.Printf("alias claude-%s-exec 'claude-switch exec %s --'\n", p.Name, p.Name)
	}
}

func printPowerShellAliases(profiles []config.ProfileEntry) {
	fmt.Println("# Claude Switch aliases for PowerShell")
	fmt.Println("# Add these to your $PROFILE")
	fmt.Println()

	for _, p := range profiles {
		funcName := strings.ReplaceAll(p.Name, "-", "")
		titled := strings.ToUpper(funcName[:1]) + funcName[1:]
		fmt.Printf("function Switch-Claude%s { claude-switch use %s }\n", titled, p.Name)
		fmt.Printf("function Claude%s { claude-switch use %s; claude @args }\n", titled, p.Name)
	}
}

func init() {
	aliasCmd.Flags().StringVarP(&aliasShell, "shell", "s", "bash", "Shell type (bash, zsh, fish, powershell)")
	rootCmd.AddCommand(aliasCmd)
}
