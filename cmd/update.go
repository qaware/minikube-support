package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type UpdateOptions struct {
	registry apis.InstallablePluginRegistry
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
		Short: "Updates all or one of the available plugins.",
		Run:   options.Run,
	}
	command.AddCommand(CreateUpdateCommands(options.registry)...)
	return command

}

func (i *UpdateOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.registry.ListPlugins() {
		logrus.Info("Update plugin:", plugin)
		plugin.Update()
	}
}
