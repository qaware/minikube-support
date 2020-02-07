package coredns

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
)

// Manager defines the interface for plugins to interact with the CoreDNS backend.
type Manager interface {
	// AddHost adds a new host to ip resource record. Depending on the type of
	// the ip it can be either an "A" or an "AAAA" record.
	// It allows to store multiple targets for the same host name. In this case
	// the regular round-robin mechanism for load balancing will be used.
	// If either the hostname or the target ip is not valid it will return an error.
	AddHost(hostName string, ip string) error

	// AddAlias adds a new host to CNAME resource record.
	// It allows to store multiple targets for the same host name. In this case
	// the regular round-robin mechanism for load balancing will be used.
	// If either the hostname or the target is not valid it will return an error.
	AddAlias(hostName string, target string) error

	// RemoveHost removes all "A", "AAAA" and "CNAME" records for the given hostname.
	RemoveHost(hostName string)
}

// AddResourceRecordFunc is the function signature for adding resource records.
// It is used to add A, AAAA and CNAME entries.
type AddResourceRecordFunc func(string, string) error

// AddResourceRecordFunc is the function signature for adding resource records.
// It is used to add A, AAAA and CNAME entries.
type RemoveResourceRecordFunc func(string)

// grpcManager is the default implementation of the Manager interface.
type grpcManager struct {
	plugin *grpcPlugin
}

// NewManager initializes a new Manager instance with the given plugin as backend.
func NewManager(plugin apis.StartStopPlugin) (Manager, error) {
	p, ok := plugin.(*grpcPlugin)
	if !ok {
		return nil, fmt.Errorf("try to get server from unknown plugin type %s", plugin)
	}

	return &grpcManager{plugin: p}, nil
}

func (m *grpcManager) AddHost(hostName string, ip string) error {
	coreDnsBackend, e := GetServer(m.plugin)
	if e != nil {
		return e
	}
	return coreDnsBackend.AddHost(hostName, ip)
}

func (m *grpcManager) AddAlias(hostName string, target string) error {
	coreDnsBackend, e := GetServer(m.plugin)
	if e != nil {
		return e
	}
	return coreDnsBackend.AddCNAME(hostName, target)
}

func (m *grpcManager) RemoveHost(hostName string) {
	coreDnsBackend, e := GetServer(m.plugin)
	if e != nil {
		logrus.Errorf("can not determ coredns server backend: %s", e)
		return
	}
	coreDnsBackend.RemoveResourceRecord(hostName, dns.Type(dns.TypeA))
	coreDnsBackend.RemoveResourceRecord(hostName, dns.Type(dns.TypeAAAA))
	coreDnsBackend.RemoveResourceRecord(hostName, dns.Type(dns.TypeCNAME))
}

// noOpManager is a fallback implementation of the Manager interface which just
// logs the calls to AddHost(), AddAlias() and RemoveHost().
type noOpManager struct{}

// NewNoOpManager initializes the noOpManager for the Manager interface.
func NewNoOpManager() Manager {
	return &noOpManager{}
}

// Addhost is a dummy function that just logs the addition of the given domain to the dns backend.
func (noOpManager) AddHost(hostName string, ip string) error {
	logrus.Infof("Would add new A or AAAA dns entry for %s to %s.", hostName, ip)
	return nil
}

// AddAlias is a dummy function that just logs the addition of the given domain to the dns backend.
func (noOpManager) AddAlias(hostName string, target string) error {
	logrus.Infof("Would add new CNAME dns entry for %s to %s.", hostName, target)
	return nil
}

// RemoveHost is a dummy function that just logs the removal of the given domain from the dns backend.
func (noOpManager) RemoveHost(hostName string) {
	logrus.Infof("Would remove A or AAAA dns entry for %s.", hostName)
}
