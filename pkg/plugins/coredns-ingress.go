package plugins

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
)

func combineCoreDnsAndIngress() ([]apis.StartStopPlugin, error) {
	coreDns := coredns.NewGrpcPlugin()

	manager, _ := coredns.NewManager(coreDns)
	k8sIngresses := ingress.NewK8sIngress("", manager)

	return []apis.StartStopPlugin{coreDns, k8sIngresses}, nil
}

func NewCoreDnsIngressPlugin() *CombinedStartStopPlugin {
	return NewCombinedPlugin("coredns-ingress", combineCoreDnsAndIngress, true)
}
