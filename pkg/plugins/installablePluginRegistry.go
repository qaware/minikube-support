package plugins

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/chr-fritz/minikube-support/pkg/plugins/mkcert"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Singleton Plugin Registry
var installPlugins *installablePluginRegistry

// The plugin registry.
type installablePluginRegistry struct {
	plugins map[string]apis.InstallablePlugin
}
type runnerFunc func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string)

// Initializes the plugin registry.
func init() {
	installPlugins = newInstallablePluginRegistry()
	installPlugins.addPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(),
	)
}

// Create the install command for all registered plugins.
func CreateInstallCommands() []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		plugin.Install()
	}
	return createCommands("Installs the %s plugin.", runner)
}

// Create the update command for all registered plugins.
func CreateUpdateCommands() []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		plugin.Update()
	}
	return createCommands("Updates the %s plugin.", runner)
}

// Create the uninstall command for all registered plugins.
func CreateUninstallCommands() []*cobra.Command {
	runner := func(plugin apis.InstallablePlugin, cmd *cobra.Command, args []string) {
		purge, e := cmd.Flags().GetBool("purge")
		if e != nil {
			logrus.Panicf("can not find flag purge: %s", e)
		}
		plugin.Uninstall(purge)
	}
	return createCommands("Uninstall the %s plugin.", runner)
}

// GetInstallablePlugins returns a list with all registered installable plugins.
func GetInstallablePlugins() []apis.InstallablePlugin {
	var values []apis.InstallablePlugin
	for _, v := range installPlugins.plugins {
		values = append(values, v)
	}
	return values
}

// Initializes a new plugin registry.
func newInstallablePluginRegistry() *installablePluginRegistry {
	return &installablePluginRegistry{
		plugins: map[string]apis.InstallablePlugin{},
	}
}

// Registers some plugins.
func (r *installablePluginRegistry) addPlugins(plugins ...apis.InstallablePlugin) {
	for _, plugin := range plugins {
		r.addPlugin(plugin)
	}
}

// Registers a single plugin.
func (r *installablePluginRegistry) addPlugin(plugin apis.InstallablePlugin) {
	if plugin == nil {
		logrus.Panicf("Can not add nil plugin to registry")
		return
	}

	if _, ok := r.plugins[plugin.String()]; ok {
		logrus.Panicf("Can not add plugin '%s' twice.", plugin)
		return
	}

	r.plugins[plugin.String()] = plugin
}

// Create the commands for all registered plugins with the given description format and runner function.
func createCommands(short string, runner runnerFunc) []*cobra.Command {
	var commands []*cobra.Command
	for _, plugin := range installPlugins.plugins {
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
