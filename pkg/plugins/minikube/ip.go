package minikube

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/plugins/coredns"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

// ip is a simple plugin which adds a new resource entry for "vm.minikube." to the minikube ip address.
type ip struct {
	addIpTimer        *time.Timer
	dnsBackendManager coredns.Manager
}

const ipPluginName = "minikube-ip"

// NewIpPlugin initializes the minikube ip address plugin.
func NewIpPlugin(manager coredns.Manager) apis.StartStopPlugin {
	return &ip{dnsBackendManager: manager}
}

func (i *ip) String() string {
	return ipPluginName
}

func (i *ip) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	i.addIpTimer = time.AfterFunc(10*time.Second, i.addVmIp)
	return ipPluginName, nil
}

func (*ip) IsSingleRunnable() bool {
	return false
}

func (i *ip) Stop() error {
	i.addIpTimer.Stop()
	return nil
}

// addVmIp tries to get the current minikube ip and adds a new resource entry "vm.minikube" to this ip.
func (i *ip) addVmIp() {
	cmd := sh.ExecCommand("minikube", "ip")
	bytes, e := cmd.Output()
	if e != nil {
		logrus.Errorf("can not determ minikube ip: %s", e)
		return
	}
	ip := string(bytes)
	ip = strings.Trim(ip, "\n\r \t")
	e = i.dnsBackendManager.AddHost("vm.minikube", ip)
	if e != nil {
		logrus.Errorf("unable to add record for vm.minikube: %s", e)
		return
	}
}
