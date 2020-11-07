package cmd

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type UpdateOptions struct {
	registry            apis.InstallablePluginRegistry
	includeLocalPlugins bool
}

func NewUpdateOptions(registry apis.InstallablePluginRegistry) *UpdateOptions {
	return &UpdateOptions{
		registry: registry,
	}
}

func NewUpdateCommand(registry apis.InstallablePluginRegistry) *cobra.Command {
	options := NewUpdateOptions(registry)

	command := &cobra.Command{
		Use:   "update",
		Short: "Updates the cluster plugins.",
		Long: "The update command updates at least all cluster plugins. If you add the -l or --updateLocal flag " +
			"it will also update the local plugins.",
		Run: options.Run,
	}
	command.AddCommand(CreateUpdateCommands(options.registry)...)
	command.Flags().BoolVarP(&options.includeLocalPlugins, "uninstallLocal", "l", false, "Also remove the local plugins.")
	return command
}

func (i *UpdateOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.registry.ListPlugins() {
		if !i.includeLocalPlugins && apis.IsLocalPlugin(plugin) {
			continue
		}
		logrus.Info("Update plugin:", plugin)
		plugin.Update()
	}
}
