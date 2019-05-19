package ingress

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type AddResourceRecordFunc func(string, string) error

type k8sIngress struct {
	clientConfig   string
	clientSet      *kubernetes.Clientset
	messageChannel chan *apis.MonitoringMessage
	addA           AddResourceRecordFunc
	addAAAA        AddResourceRecordFunc
	removeA        func(string)
	removeAAAA     func(string)
	watch          watch.Interface
}

type entry struct {
	hostName string
	ip       string
	ipType   int
}

const pluginName = "kubernetes-ingress"

func NewK8sIngress(clientConfig string, addA AddResourceRecordFunc, addAAAA AddResourceRecordFunc, removeA func(string), removeAAAA func(string)) apis.StartStopPlugin {
	if addA == nil {
		addA = noopAddA
	}
	if addAAAA == nil {
		addAAAA = noopAddAAAA
	}
	if removeA == nil {
		removeA = noopRemoveA
	}
	if removeAAAA == nil {
		removeAAAA = noopRemoveAAAA
	}

	return &k8sIngress{
		clientConfig: clientConfig,
		addA:         addA,
		addAAAA:      addAAAA,
		removeA:      removeA,
		removeAAAA:   removeAAAA,
	}
}

func (*k8sIngress) String() string {
	return pluginName
}

func (k8s *k8sIngress) Start(messageChannel chan *apis.MonitoringMessage) (string, error) {
	k8s.messageChannel = messageChannel
	e := k8s.openRestConfig()
	if e != nil {
		return "", fmt.Errorf("config error: %s", e)
	}

	ingresses := k8s.clientSet.
		ExtensionsV1beta1().
		Ingresses("default")
	ingressList, e := ingresses.List(metav1.ListOptions{})
	if e != nil {
		return "", fmt.Errorf("can not list ingresses: %s", e)
	}

	for _, ingress := range ingressList.Items {
		k8s.handleAddedIngress(ingress)
	}

	go k8s.watchIngresses(ingressList.ResourceVersion)
	return pluginName, nil
}

func (k8s *k8sIngress) Stop() error {
	k8s.watch.Stop()
	return nil
}

func (k8s *k8sIngress) watchIngresses(resourceVersion string) {
	ingresses := k8s.clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)

	w, e := ingresses.Watch(metav1.ListOptions{ResourceVersion: resourceVersion})
	if e != nil {
		logrus.Errorf("Can not start watch for ingresses: %s", e)
	}
	k8s.watch = w
	for event := range w.ResultChan() {
		ingress := event.Object.(*v1beta1.Ingress)
		switch event.Type {
		case watch.Added:
			k8s.handleAddedIngress(*ingress)
		case watch.Modified:
			k8s.handleUpdatedIngress(*ingress)
		case watch.Deleted:
			k8s.handleDeletedIngress(*ingress)
		default:
			logrus.Infof("Received unhandled event %s for ingress %s/%s", event.Type, ingress.GetNamespace(), ingress.GetName())
		}
	}
}

func (k8s *k8sIngress) handleAddedIngress(ingress v1beta1.Ingress) error {
	logrus.Infof("Ingress added: %s", ingress)
	return nil
}
func (k8s *k8sIngress) handleUpdatedIngress(ingress v1beta1.Ingress) error {
	logrus.Infof("Ingress updated: %s", ingress)
	return nil
}
func (k8s *k8sIngress) handleDeletedIngress(ingress v1beta1.Ingress) error {
	logrus.Infof("Ingress removed: %s", ingress)
	return nil
}

func (k8s *k8sIngress) openRestConfig() error {
	var e error
	var config *rest.Config
	if k8s.clientConfig == "" {
		config, e = rest.InClusterConfig()

		// if not run in cluster try to use default from user home
		if e == rest.ErrNotInCluster {
			homeDir := homedir.HomeDir()
			configPath := filepath.Join(homeDir, ".kube", "config")
			config, e = clientcmd.BuildConfigFromFlags("", configPath)
		}

		// Neither in cluster config nor user home config exists.
		if e != nil {
			return fmt.Errorf("can not determ config: %s", e)
		}
	} else {
		// Use config from given file name.
		config, e = clientcmd.BuildConfigFromFlags("", k8s.clientConfig)
		if e != nil {
			return fmt.Errorf("can not read config from file %s: %s", k8s.clientConfig, e)
		}
	}

	clientSet, e := kubernetes.NewForConfig(config)
	if e != nil {
		return fmt.Errorf("unable to create clientSet: %s", e)
	}
	k8s.clientSet = clientSet
	return nil
}

func noopAddA(domain string, ip string) error {
	logrus.Infof("Would add new A dns entry for %s to %s.", domain, ip)
	return nil
}
func noopRemoveA(domain string) {
	logrus.Infof("Would remove A dns entry for %s.", domain)

}
func noopAddAAAA(domain string, ip string) error {
	logrus.Infof("Would add new AAAA dns entry for %s to %s.", domain, ip)
	return nil
}
func noopRemoveAAAA(domain string) {
	logrus.Infof("Would remove AAAA dns entry for %s.", domain)
}
