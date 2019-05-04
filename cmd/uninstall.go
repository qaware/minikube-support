package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type UninstallOptions struct {
	purge bool
}

func NewUninstallOptions() *UninstallOptions {
	return &UninstallOptions{}
}

func NewUninstallCommand() *cobra.Command {
	options := NewUninstallOptions()

	command := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls all or one of the available plugins.",
		Run:   options.Run,
	}
	command.PersistentFlags().BoolVarP(&options.purge, "purge", "p", false, "Remove also any local installed tools regarding this plugin.")
	command.AddCommand(plugins.CreateUninstallCommands()...)
	return command
}

func (i *UninstallOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range plugins.GetInstallablePlugins() {
		logrus.Info("Uninstall plugin:", plugin)
		plugin.Uninstall(i.purge)
	}
}

func init() {
	rootCmd.AddCommand(NewUninstallCommand())
}
