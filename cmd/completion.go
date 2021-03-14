package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the `completion` command
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(minikube-support completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ minikube-support completion bash > /etc/bash_completion.d/minikube-support
  # macOS:
  $ minikube-support completion bash > /usr/local/etc/bash_completion.d/minikube-support

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ minikube-support completion zsh > "${fpath[1]}/_minikube-support"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ minikube-support completion fish | source

  # To load completions for each session, execute once:
  $ minikube-support completion fish > ~/.config/fish/completions/minikube-support.fish

PowerShell:

  PS> minikube-support completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> minikube-support completion powershell > minikube-support.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell", "ps1"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				fallthrough
			case "ps1":
				_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCompletionCmd())
}
