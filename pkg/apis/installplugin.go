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
