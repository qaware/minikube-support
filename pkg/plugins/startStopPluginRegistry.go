package plugins

import (
	"fmt"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
)

// The plugin registry.
type startStopPluginRegistry struct {
	plugins map[string]apis.StartStopPlugin
}

// Initializes a new plugin registry.
func NewStartStopPluginRegistry() apis.StartStopPluginRegistry {
	return &startStopPluginRegistry{
		plugins: map[string]apis.StartStopPlugin{},
	}
}

// Registers some plugins.
func (r *startStopPluginRegistry) AddPlugins(plugins ...apis.StartStopPlugin) {
	for _, plugin := range plugins {
		r.AddPlugin(plugin)
	}
}

// Registers a single plugin.
func (r *startStopPluginRegistry) AddPlugin(plugin apis.StartStopPlugin) {
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
func (r *startStopPluginRegistry) ListPlugins() []apis.StartStopPlugin {
	var values []apis.StartStopPlugin
	for _, v := range r.plugins {
		values = append(values, v)
	}
	return values
}

// FindPlugin tries to find and return a plugin with the given name. Otherwise it would return an error.
func (r *startStopPluginRegistry) FindPlugin(name string) (apis.StartStopPlugin, error) {
	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}
	return plugin, nil
}
