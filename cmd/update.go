package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type UpdateOptions struct {
	plugins []apis.InstallablePlugin
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		plugins: plugins.GetInstallablePluginRegistry().ListPlugins(),
	}
}

func NewUpdateCommand() *cobra.Command {
	options := NewUpdateOptions()

	command := &cobra.Command{
		Use:   "update",
		Short: "Updates all or one of the available plugins.",
		Run:   options.Run,
	}
	command.AddCommand(plugins.CreateUpdateCommands()...)
	return command

}

func (i *UpdateOptions) Run(cmd *cobra.Command, args []string) {
	for _, plugin := range i.plugins {
		logrus.Info("Update plugin:", plugin)
		plugin.Update()
	}
}

func init() {
	rootCmd.AddCommand(NewUpdateCommand())
}
