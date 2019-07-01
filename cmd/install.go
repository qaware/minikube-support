package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type InstallOptions struct {
	registry apis.InstallablePluginRegistry
}

func NewInstallOptions(registry apis.InstallablePluginRegistry) *InstallOptions {
	return &InstallOptions{
		registry: registry,
	}
}

func NewInstallCommand(registry apis.InstallablePluginRegistry) *cobra.Command {
	options := NewInstallOptions(registry)

	command := &cobra.Command{
		Use:   "install",
		Short: "Installs all or one of the available plugins.",
		Run:   options.Run,
	}
	command.AddCommand(createInstallCommands(options.registry)...)
	return command

}

func (i *InstallOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.registry.ListPlugins() {
		logrus.Info("Install plugin:", plugin)
		plugin.Install()
	}
}
