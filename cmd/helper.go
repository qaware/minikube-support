package cmd

import (
	"fmt"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runnerFunc func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string)

// Create the install command for all registered plugins.
func createInstallCommands(registry apis.InstallablePluginRegistry) []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		plugin.Install()
	}
	return createCommands("Installs the %s plugin.", runner, registry)
}

// Create the update command for all registered plugins.
func CreateUpdateCommands(registry apis.InstallablePluginRegistry) []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		plugin.Update()
	}
	return createCommands("Updates the %s plugin.", runner, registry)
}

// Create the uninstall command for all registered plugins.
func createUninstallCommands(registry apis.InstallablePluginRegistry) []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		purge, e := cmd.Flags().GetBool("purge")
		if e != nil {
			logrus.Panicf("can not find flag purge: %s", e)
		}
		plugin.Uninstall(purge)
	}
	return createCommands("Uninstall the %s plugin.", runner, registry)
}

// Create the commands for all registered plugins with the given description format and runner function.
func createCommands(short string, runner runnerFunc, registry apis.InstallablePluginRegistry) []*cobra.Command {
	var commands []*cobra.Command
	for _, plugin := range registry.ListPlugins() {
		cmd := createCommand(plugin, short, runner)
		commands = append(commands, cmd)
	}
	return commands
}

// Create the command instance for a single plugin with the given description format and runner function.
func createCommand(plugin apis.InstallablePlugin, short string, runner runnerFunc) *cobra.Command {
	return &cobra.Command{
		Use:   plugin.String(),
		Short: fmt.Sprintf(short, plugin),
		Run: func(cmd *cobra.Command, args []string) {
			runner(plugin, cmd, args)
		},
	}
}
