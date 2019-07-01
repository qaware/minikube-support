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

	Phase() Phase
}

type Phase int

const (
	LOCAL_TOOLS_INSTALL   Phase = 0
	LOCAL_TOOLS_CONFIG    Phase = 5
	CLUSTER_INIT          Phase = 10
	CLUSTER_CONFIG        Phase = 15
	CLUSTER_TOOLS_INSTALL Phase = 20
	CLUSTER_TOOLS_CONFIG  Phase = 25
)

// InstallablePluginRegistry is the registry which collects all InstallablePlugins and provides easy access to them.
type InstallablePluginRegistry interface {
	// AddPlugin adds a single plugin to the registry.
	AddPlugin(plugin InstallablePlugin)
	// AddPlugins adds the given list of plugins to the registry.
	AddPlugins(plugins ...InstallablePlugin)
	// ListPlugins returns a list of all currently registered plugins in the registry.
	ListPlugins() InstallablePluginList
	// FindPlugin tries to find and return a plugin with the given name. Otherwise it would return an error.
	FindPlugin(name string) (InstallablePlugin, error)
}

// InstallablePluginList is a simple slice which can be sorted by the install Phase.
type InstallablePluginList []InstallablePlugin

func (l InstallablePluginList) Len() int {
	return len(l)
}

func (l InstallablePluginList) Less(i, j int) bool {
	return l[i].Phase() < l[j].Phase()
}

func (l InstallablePluginList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
