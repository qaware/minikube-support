package ingress

import (
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/sirupsen/logrus"
)

type controllerInstaller struct {
	manager     helm.Manager
	releaseName string
	namespace   string
	values      map[string]interface{}
}

func NewControllerInstaller(manager helm.Manager) apis.InstallablePlugin {
	return &controllerInstaller{
		manager:     manager,
		releaseName: "nginx-ingress",
		values:      map[string]interface{}{},
		namespace:   "mks",
	}
}

func (*controllerInstaller) String() string {
	return "ingress-controller"
}

func (i *controllerInstaller) Install() {
	if e := i.manager.AddRepository("ingress-nginx", "https://kubernetes.github.io/ingress-nginx"); e != nil {
		logrus.Errorf("Unable to add nginx-ingress repository: %s", e)
		return
	}
	i.Update()
}

func (i *controllerInstaller) Update() {
	if e := i.manager.UpdateRepository(); e != nil {
		logrus.Errorf("Unable to update helm repositories %s", e)
		return
	}
	i.values["controller.publishService.enabled"] = "true"

	i.manager.Install("ingress-nginx/ingress-nginx", i.releaseName, i.namespace, i.values, false)
}

func (i *controllerInstaller) Uninstall(purge bool) {
	i.manager.Uninstall(i.releaseName, i.namespace, purge)
}

func (*controllerInstaller) Phase() apis.Phase {
	return apis.CLUSTER_TOOLS_INSTALL
}
