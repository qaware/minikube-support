package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/chr-fritz/minikube-support/pkg/plugins/minikube"
	"github.com/chr-fritz/minikube-support/pkg/plugins/mkcert"
)

// Initializes all active plugins and register them in the two (installable and start stop) plugin registries.
func init() {
	plugins.GetInstallablePluginRegistry().AddPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(),
	)

	plugins.GetStartStopPluginRegistry().AddPlugins(
		minikube.NewTunnel(),
		plugins.NewCoreDnsIngressPlugin(),
	)
}
