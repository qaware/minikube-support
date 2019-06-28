package plugins

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
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

// GetInstallablePluginRegistry returns the instance for the InstallablePluginRegistry.
func GetInstallablePluginRegistry() apis.InstallablePluginRegistry {
	return installPlugins
}

// Initializes a new plugin registry.
func newInstallablePluginRegistry() *installablePluginRegistry {
	return &installablePluginRegistry{
		plugins: map[string]apis.InstallablePlugin{},
	}
}

// Registers some plugins.
func (r *installablePluginRegistry) AddPlugins(plugins ...apis.InstallablePlugin) {
	for _, plugin := range plugins {
		r.AddPlugin(plugin)
	}
}

// Registers a single plugin.
func (r *installablePluginRegistry) AddPlugin(plugin apis.InstallablePlugin) {
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

// ListPlugins returns a list with all registered installable plugins.
func (r *installablePluginRegistry) ListPlugins() []apis.InstallablePlugin {
	var values []apis.InstallablePlugin
	for _, v := range r.plugins {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Phase() < values[j].Phase()
	})
	return values
}

// FindPlugin finds a single plugin by its name. If not found it returns an error.
func (r *installablePluginRegistry) FindPlugin(name string) (apis.InstallablePlugin, error) {
	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}
	return plugin, nil
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
