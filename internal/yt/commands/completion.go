package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion scripts for your shell",
	Long: `Generate shell completion scripts for the yt command.

The completion script will provide tab completion for commands, options, and flags.

To load completions:

Bash:
  $ source <(yt completion bash)
  
  # To load completions for each session, execute once:
  # Linux:
  $ yt completion bash > /etc/bash_completion.d/yt
  # macOS:
  $ yt completion bash > /usr/local/etc/bash_completion.d/yt

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  
  # To load completions for each session, execute once:
  $ yt completion zsh > "${fpath[1]}/_yt"
  
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ yt completion fish | source
  
  # To load completions for each session, execute once:
  $ yt completion fish > ~/.config/fish/completions/yt.fish

PowerShell:
  PS> yt completion powershell | Out-String | Invoke-Expression
  
  # To load completions for every new session, run:
  PS> yt completion powershell > yt.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell type: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
