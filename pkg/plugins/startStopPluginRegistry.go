package plugins

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/minikube"
	"github.com/sirupsen/logrus"
)

// Singleton Plugin Registry
var startStopPlugins *startStopPluginRegistry

// The plugin registry.
type startStopPluginRegistry struct {
	plugins map[string]apis.StartStopPlugin
}

// Initializes the plugin registry.
func init() {
	startStopPlugins = newStartStopPluginRegistry()
	startStopPlugins.addPlugins(minikube.NewTunnel())
}

// GetInstallablePlugins returns a list with all registered installable plugins.
func GetStartStopPlugins() []apis.StartStopPlugin {
	var values []apis.StartStopPlugin
	for _, v := range startStopPlugins.plugins {
		values = append(values, v)
	}
	return values
}

// Initializes a new plugin registry.
func newStartStopPluginRegistry() *startStopPluginRegistry {
	return &startStopPluginRegistry{
		plugins: map[string]apis.StartStopPlugin{},
	}
}

// Registers some plugins.
func (r *startStopPluginRegistry) addPlugins(plugins ...apis.StartStopPlugin) {
	for _, plugin := range plugins {
		r.addPlugin(plugin)
	}
}

// Registers a single plugin.
func (r *startStopPluginRegistry) addPlugin(plugin apis.StartStopPlugin) {
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
