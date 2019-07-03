package cmd

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/kubernetes"
	"github.com/chr-fritz/minikube-support/pkg/packagemanager/helm"
	"github.com/chr-fritz/minikube-support/pkg/plugins"
	"github.com/chr-fritz/minikube-support/pkg/plugins/certmanager"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/chr-fritz/minikube-support/pkg/plugins/minikube"
	"github.com/chr-fritz/minikube-support/pkg/plugins/mkcert"
	"github.com/hashicorp/go-multierror"
)

// Initializes all active plugins and register them in the two (installable and start stop) plugin registries.
func initPlugins(options *RootCommandOptions) {
	var errors *multierror.Error

	handler := kubernetes.NewContextHandler(&options.kubeConfig, &options.contextName)
	helmManager := helm.NewHelmManager()

	coreDns := coredns.NewGrpcPlugin()
	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := ingress.NewK8sIngress(handler, manager)
	coreDnsIngressPlugin, _ := plugins.NewCombinedPlugin("coredns-ingress", []apis.StartStopPlugin{coreDns, k8sIngresses}, true)

	certManager, e := certmanager.NewCertManager(helmManager, handler)
	errors = multierror.Append(errors, e)

	options.installablePluginRegistry.AddPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(helmManager),
		certManager,
	)

	options.startStopPluginRegistry.AddPlugins(
		minikube.NewTunnel(),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager),
	)
}
