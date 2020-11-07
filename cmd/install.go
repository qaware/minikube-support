package cmd

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type InstallOptions struct {
	registry            apis.InstallablePluginRegistry
	includeLocalPlugins bool
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
	flags := command.Flags()
	flags.BoolVarP(&options.includeLocalPlugins, "installLocal", "l", false, "Install cluster and local plugins.")

	command.AddCommand(createInstallCommands(options.registry)...)
	return command

}

func (i *InstallOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.registry.ListPlugins() {
		logrus.Info("Install plugin:", plugin)
		if !i.includeLocalPlugins && apis.IsLocalPlugin(plugin) {
			continue
		}

		plugin.Install()
	}
}
