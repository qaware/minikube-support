package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/kubernetes"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/chr-fritz/minikube-support/pkg/plugins/minikube"
	"github.com/chr-fritz/minikube-support/pkg/plugins/mkcert"
)

// Initializes all active plugins and register them in the two (installable and start stop) plugin registries.
func initPlugins(options *RootCommandOptions) {
	handler := kubernetes.NewContextHandler(&options.kubeConfig, &options.contextName)

	coreDns := coredns.NewGrpcPlugin()
	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := ingress.NewK8sIngress(handler, manager)
	coreDnsIngressPlugin, _ := plugins.NewCombinedPlugin("coredns-ingress", []apis.StartStopPlugin{coreDns, k8sIngresses}, true)

	options.installablePluginRegistry.AddPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(),
	)

	options.startStopPluginRegistry.AddPlugins(
		minikube.NewTunnel(),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager),
	)
}
