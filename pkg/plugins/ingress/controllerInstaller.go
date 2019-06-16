package ingress

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/packagemanager/helm"
)

type controllerInstaller struct {
	manager     helm.Manager
	releaseName string
	namespace   string
	values      map[string]interface{}
}

func NewControllerInstaller() apis.InstallablePlugin {
	return &controllerInstaller{manager: helm.NewHelmManager(), releaseName: "nginx-ingress"}
}

func (*controllerInstaller) String() string {
	return "ingress-controller"
}

func (i *controllerInstaller) Install() {
	i.Update()
}

func (i *controllerInstaller) Update() {
	i.values["controller.publishService.enabled"] = "true"

	i.manager.Install("stable/nginx-ingress", i.releaseName, i.namespace, i.values)
}

func (i *controllerInstaller) Uninstall(purge bool) {
	i.manager.Uninstall(i.releaseName, purge)
}
