package apis

import (
	"fmt"
)

// InstallablePlugin is a plugin that can install/update/uninstall tools local or within minikube or both.
type InstallablePlugin interface {
	// Must return the name of the plugin. This name will also be used for single commands.
	fmt.Stringer

	// Installs the tools.
	// Should print information about the process.
	Install()

	// Updates the tools.
	// Should print information about the process.
	Update()

	// Uninstall the tools.
	// Should print information about the process.
	Uninstall(purge bool)
}

// InstallablePluginRegistry is the registry which collects all InstallablePlugins and provides easy access to them.
type InstallablePluginRegistry interface {
	// AddPlugin adds a single plugin to the registry.
	AddPlugin(plugin InstallablePlugin)
	// AddPlugins adds the given list of plugins to the registry.
	AddPlugins(plugins ...InstallablePlugin)
	// ListPlugins returns a list of all currently registered plugins in the registry.
	ListPlugins() []InstallablePlugin
	// FindPlugin tries to find and return a plugin with the given name. Otherwise it would return an error.
	FindPlugin(name string) (InstallablePlugin, error)
}
