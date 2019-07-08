package cmd

import (
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/qaware/minikube-support/pkg/plugins/certmanager"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/qaware/minikube-support/pkg/plugins/ingress"
	"github.com/qaware/minikube-support/pkg/plugins/minikube"
	"github.com/qaware/minikube-support/pkg/plugins/mkcert"
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
		coredns.NewInstaller("/opt/mks/coredns/"),
	)

	options.startStopPluginRegistry.AddPlugins(
		minikube.NewTunnel(),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager),
	)
}
