package plugins

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/miekg/dns"
)

func combineCoreDnsAndIngress() ([]apis.StartStopPlugin, error) {
	coreDns := coredns.NewGrpcPlugin()
	coreDnsBackend, e := coredns.GetServer(coreDns)
	if e != nil {
		return nil, fmt.Errorf("can not determ coredns server backend: %s", e)
	}

	k8sIngresses := ingress.NewK8sIngress(
		"",
		coreDnsBackend.AddA,
		coreDnsBackend.AddAAAA,
		func(domain string) {
			coreDnsBackend.RemoveResourceRecord(dns.Name(domain), dns.Type(dns.TypeA))
		},
		func(domain string) {
			coreDnsBackend.RemoveResourceRecord(dns.Name(domain), dns.Type(dns.TypeAAAA))
		})

	return []apis.StartStopPlugin{coreDns, k8sIngresses}, nil
}

func NewCoreDnsIngressPlugin() apis.StartStopPlugin {
	return NewCombinedPlugin("coredns-ingress", combineCoreDnsAndIngress)
}
