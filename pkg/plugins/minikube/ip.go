package minikube

import (
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/qaware/minikube-support/pkg/sh"
)

// ip is a simple plugin which adds a new resource entry for "vm.minikube." to the minikube ip address.
type ip struct {
	mutex             sync.Mutex
	addIpTimer        *time.Timer
	dnsBackendManager coredns.Manager
	contextHandler    kubernetes.ContextHandler
}

const ipPluginName = "minikube-ip"

// NewIpPlugin initializes the minikube ip address plugin.
func NewIpPlugin(manager coredns.Manager, handler kubernetes.ContextHandler) apis.StartStopPlugin {
	return &ip{dnsBackendManager: manager, contextHandler: handler, mutex: sync.Mutex{}}
}

func (i *ip) String() string {
	return ipPluginName
}

func (i *ip) Start(chan *apis.MonitoringMessage) (boxName string, err error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.addIpTimer = time.AfterFunc(10*time.Second, i.addVmIp)
	return ipPluginName, nil
}

func (*ip) IsSingleRunnable() bool {
	return false
}

func (i *ip) Stop() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.addIpTimer.Stop()
	return nil
}

// addVmIp tries to get the current minikube ip and adds a new resource entry "vm.minikube" to this ip.
func (i *ip) addVmIp() {
	isMinikube, e := i.contextHandler.IsMinikube()
	if e != nil {
		logrus.Errorf("can not determ if running in minikube: %s", e)
		return
	}
	if !isMinikube {
		logrus.Info("Context is not set to minikube. Do not add A-record for vm.minikube.")
		return
	}

	ip, e := sh.RunCmd("minikube", "ip")
	if e != nil {
		logrus.Errorf("can not determ minikube ip: %s", e)
		return
	}

	ip = strings.Trim(ip, "\n\r \t")
	e = i.dnsBackendManager.AddHost("vm.minikube", ip)
	if e != nil {
		logrus.Errorf("unable to add record for vm.minikube: %s", e)
		return
	}
}
