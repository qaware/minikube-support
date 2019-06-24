package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
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

	coreDns := coredns.NewGrpcPlugin()
	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := ingress.NewK8sIngress("", manager)
	coreDnsIngressPlugin, _ := plugins.NewCombinedPlugin("coredns-ingress", []apis.StartStopPlugin{coreDns, k8sIngresses}, true)

	plugins.GetStartStopPluginRegistry().AddPlugins(
		minikube.NewTunnel(),
		coreDnsIngressPlugin,
	)
}
