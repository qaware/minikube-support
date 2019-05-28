package plugins

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/plugins/ingress"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func combineCoreDnsAndIngress() ([]apis.StartStopPlugin, error) {
	coreDns := coredns.NewGrpcPlugin()

	k8sIngresses := ingress.NewK8sIngress(
		"",
		func(domain string, ip string) error {
			coreDnsBackend, e := coredns.GetServer(coreDns)
			if e != nil {
				return e
			}
			return coreDnsBackend.AddHost(domain, ip)
		},
		func(domain string, target string) error {
			coreDnsBackend, e := coredns.GetServer(coreDns)
			if e != nil {
				return e
			}
			return coreDnsBackend.AddCNAME(domain, target)
		},
		func(domain string) {
			coreDnsBackend, e := coredns.GetServer(coreDns)
			if e != nil {
				logrus.Errorf("can not determ coredns server backend: %s", e)
				return
			}
			coreDnsBackend.RemoveResourceRecord(dns.Name(domain), dns.Type(dns.TypeA))
			coreDnsBackend.RemoveResourceRecord(dns.Name(domain), dns.Type(dns.TypeAAAA))
			coreDnsBackend.RemoveResourceRecord(dns.Name(domain), dns.Type(dns.TypeCNAME))
		})

	return []apis.StartStopPlugin{coreDns, k8sIngresses}, nil
}

func NewCoreDnsIngressPlugin() apis.StartStopPlugin {
	return NewCombinedPlugin("coredns-ingress", combineCoreDnsAndIngress)
}
