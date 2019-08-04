package cmd

import (
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/plugins"
	"github.com/qaware/minikube-support/pkg/plugins/certmanager"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/qaware/minikube-support/pkg/plugins/ingress"
	"github.com/qaware/minikube-support/pkg/plugins/logs"
	"github.com/qaware/minikube-support/pkg/plugins/minikube"
	"github.com/qaware/minikube-support/pkg/plugins/mkcert"
	"github.com/sirupsen/logrus"
)

// Initializes all active plugins and register them in the two (installable and start stop) plugin registries.
func initPlugins(options *RootCommandOptions) {
	logPlugin := logs.NewLogsPlugin(logrus.StandardLogger())
	var errors *multierror.Error

	handler := kubernetes.NewContextHandler(&options.kubeConfig, &options.contextName)
	helmManager := helm.NewHelmManager(handler)

	coreDns := coredns.NewGrpcPlugin()
	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := ingress.NewK8sIngress(handler, manager)

	ghClient := github.NewClient()
	options.AddPreRunInitFunction(func(o *RootCommandOptions) error {
		ghClient.SetApiToken(o.githubAccessToken)
		return nil
	})

	coreDnsIngressPlugin, _ := plugins.NewCombinedPlugin("coredns-ingress", []apis.StartStopPlugin{coreDns, k8sIngresses}, true)

	certManager, e := certmanager.NewCertManager(helmManager, handler, ghClient)
	errors = multierror.Append(errors, e)

	options.installablePluginRegistry.AddPlugins(
		mkcert.CreateMkcertInstallerPlugin(),
		ingress.NewControllerInstaller(helmManager),
		certManager,
		coredns.NewInstaller("/opt/mks/coredns/", ghClient),
	)

	options.startStopPluginRegistry.AddPlugins(
		logPlugin,
		minikube.NewTunnel(),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager),
	)
}
