package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/qaware/minikube-support/pkg/apis"
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
		Short: "Installs the available cluster plugins.",
		Long: "The install command installs at least all cluster plugins. If you add the -l or --installLocal flag it " +
			"will also install the local plugins.",
		Run: options.Run,
	}
	flags := command.Flags()
	flags.BoolVarP(&options.includeLocalPlugins, "installLocal", "l", false, "Also install the local plugins.")

	command.AddCommand(createInstallCommands(options.registry)...)
	return command

}

func (i *InstallOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.registry.ListPlugins() {
		if !i.includeLocalPlugins && apis.IsLocalPlugin(plugin) {
			continue
		}

		logrus.Info("Install plugin:", plugin)
		plugin.Install()
	}
}
