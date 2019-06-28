package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type InstallOptions struct {
	plugins []apis.InstallablePlugin
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{
		plugins: plugins.GetInstallablePluginRegistry().ListPlugins(),
	}
}

func NewInstallCommand() *cobra.Command {
	options := NewInstallOptions()

	command := &cobra.Command{
		Use:   "install",
		Short: "Installs all or one of the available plugins.",
		Run:   options.Run,
	}
	command.AddCommand(plugins.CreateInstallCommands()...)
	return command

}

func (i *InstallOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.plugins {
		logrus.Info("Install plugin:", plugin)
		plugin.Install()
	}
}
