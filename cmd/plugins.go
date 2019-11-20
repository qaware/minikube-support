package cmd

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/qaware/minikube-support/pkg/plugins/certmanager"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/qaware/minikube-support/pkg/plugins/ingress"
	"github.com/qaware/minikube-support/pkg/plugins/k8sdns"
	"github.com/qaware/minikube-support/pkg/plugins/logs"
	"github.com/qaware/minikube-support/pkg/plugins/minikube"
	"github.com/qaware/minikube-support/pkg/plugins/mkcert"
	"github.com/sirupsen/logrus"
)

// Initializes all active plugins and register them in the two (installable and start stop) plugin registries.
func initPlugins(options *RootCommandOptions) {
	logPlugin := logs.NewLogsPlugin(logrus.StandardLogger())

	handler := kubernetes.NewContextHandler(&options.kubeConfig, &options.contextName)
	options.contextNameSupplier = handler.GetContextName
	helmManager := helm.NewHelmManager(handler)

	coreDns := coredns.NewGrpcPlugin()
	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := k8sdns.NewK8sDns(handler, manager, k8sdns.AccessTypeIngress)
	k8sServices := k8sdns.NewK8sDns(handler, manager, k8sdns.AccessTypeService)

	ghClient := github.NewClient()
	options.AddPreRunInitFunction(func(o *RootCommandOptions) error {
		ghClient.SetApiToken(o.githubAccessToken)
		return nil
	})

	coreDnsIngressPlugin, _ := plugins.NewCombinedPlugin("coredns-ingress", []apis.StartStopPlugin{coreDns, k8sIngresses, k8sServices}, true)
	certManager := certmanager.NewCertManager(helmManager, handler, ghClient)

	options.installablePluginRegistry.AddPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(helmManager),
		certManager,
		coredns.NewInstaller("/opt/mks/coredns", ghClient),
	)

	options.startStopPluginRegistry.AddPlugins(
		logPlugin,
		minikube.NewTunnel(handler),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager, handler),
	)
}
