package plugins

import (
	"fmt"
	"sort"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
)

// The plugin registry.
type installablePluginRegistry struct {
	plugins map[string]apis.InstallablePlugin
}

// Initializes a new plugin registry.
func NewInstallablePluginRegistry() *installablePluginRegistry {
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
func (r *installablePluginRegistry) ListPlugins() apis.InstallablePluginList {
	var values apis.InstallablePluginList
	for _, v := range r.plugins {
		values = append(values, v)
	}
	sort.Sort(values)
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
