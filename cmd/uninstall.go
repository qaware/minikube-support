package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
)

type UninstallOptions struct {
	purge    bool
	registry apis.InstallablePluginRegistry
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
		Short: "Uninstalls all or one of the available plugins.",
		Run:   options.Run,
	}
	command.PersistentFlags().BoolVarP(&options.purge, "purge", "p", false, "Remove also any local installed tools regarding this plugin.")
	command.AddCommand(createUninstallCommands(options.registry)...)
	return command
}

func (i *UninstallOptions) Run(cmd *cobra.Command, args []string) {
	plugins := i.registry.ListPlugins()
	sort.Sort(sort.Reverse(plugins))

	for _, plugin := range plugins {
		logrus.Info("Uninstall plugin:", plugin)
		plugin.Uninstall(i.purge)
	}
}
