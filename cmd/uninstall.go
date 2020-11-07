package cmd

import (
	"sort"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type UninstallOptions struct {
	purge               bool
	includeLocalPlugins bool
	registry            apis.InstallablePluginRegistry
}

func NewUninstallOptions(registry apis.InstallablePluginRegistry) *UninstallOptions {
	return &UninstallOptions{
		registry: registry,
	}
}

func NewUninstallCommand(registry apis.InstallablePluginRegistry) *cobra.Command {
	options := NewUninstallOptions(registry)

	command := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the cluster plugins.",
		Long: "The install command uninstalls at least all cluster plugins. If you add the -l or --uninstallLocal flag " +
			"it will also uninstall the local plugins.",
		Run: options.Run,
	}
	command.PersistentFlags().BoolVarP(&options.purge, "purge", "p", false, "Fully uninstall the plugin, including config files/config maps and history.")
	command.Flags().BoolVarP(&options.includeLocalPlugins, "uninstallLocal", "l", false, "Also remove the local plugins.")
	command.AddCommand(createUninstallCommands(options.registry)...)
	return command
}

func (i *UninstallOptions) Run(cmd *cobra.Command, args []string) {
	plugins := i.registry.ListPlugins()
	sort.Sort(sort.Reverse(plugins))

	for _, plugin := range plugins {
		if !i.includeLocalPlugins && apis.IsLocalPlugin(plugin) {
			continue
		}

		logrus.Info("Uninstall plugin:", plugin)
		plugin.Uninstall(i.purge)
	}
}
