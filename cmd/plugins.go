package cmd

import (
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/packagemanager/os"
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
	var errors *multierror.Error
	logPlugin := logs.NewLogsPlugin(logrus.StandardLogger())
	corednsPrefix := plugins.PluginInstallPrefix + "coredns"
	os.RegisterOsPackage()

	handler := kubernetes.NewContextHandler(&options.kubeConfig, &options.contextName)
	options.contextNameSupplier = handler.GetContextName
	helmManager, e := helm.NewHelmManager(handler)
	errors = multierror.Append(errors, e)

	coreDns := coredns.NewGrpcPlugin(corednsPrefix)
	manager, e := coredns.NewManager(coreDns)
	errors = multierror.Append(errors, e)

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
		coredns.NewInstaller(corednsPrefix, ghClient),
	)

	options.startStopPluginRegistry.AddPlugins(
		logPlugin,
		minikube.NewTunnel(handler),
		coreDnsIngressPlugin,
		minikube.NewIpPlugin(manager, handler),
	)
	if errors.Len() != 0 {
		logrus.Errorf("unable to initialize all plugins: %s", errors)
	}
}
