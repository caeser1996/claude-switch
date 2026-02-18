package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Claude Switch.

To load completions:

Bash:
  $ source <(cs completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ cs completion bash > /etc/bash_completion.d/cs
  # macOS:
  $ cs completion bash > $(brew --prefix)/etc/bash_completion.d/cs

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ cs completion zsh > "${fpath[1]}/_cs"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ cs completion fish | source
  # To load completions for each session, execute once:
  $ cs completion fish > ~/.config/fish/completions/cs.fish

PowerShell:
  PS> cs completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> cs completion powershell > cs.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
