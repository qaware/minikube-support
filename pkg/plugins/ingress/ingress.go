package ingress

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"text/tabwriter"
)

type k8sIngress struct {
	ctxHandler     kubernetes.ContextHandler
	messageChannel chan *apis.MonitoringMessage
	recordManager  coredns.Manager
	watch          watch.Interface

	currentIngresses map[string]ingressEntry
}

const pluginName = "kubernetes-ingress"

// NewK8sIngress will initialize a new ingress plugin.
// It allows to configure the functions to add and remove the hosts in the dns backend.
func NewK8sIngress(contextHandler kubernetes.ContextHandler, recordManager coredns.Manager) apis.StartStopPlugin {
	if recordManager == nil {
		recordManager = coredns.NewNoOpManager()
	}

	return &k8sIngress{
		ctxHandler:       contextHandler,
		recordManager:    recordManager,
		currentIngresses: make(map[string]ingressEntry),
	}
}

func (*k8sIngress) IsSingleRunnable() bool {
	return true
}

func (*k8sIngress) String() string {
	return pluginName
}

// Start starts the ingress plugin. It will automatically add all current ingresses.
func (k8s *k8sIngress) Start(messageChannel chan *apis.MonitoringMessage) (string, error) {
	k8s.messageChannel = messageChannel

	clientSet, e := k8s.ctxHandler.GetClientSet()
	if e != nil {
		return "", fmt.Errorf("can not get clientSet: %s", e)
	}

	ingresses := clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)
	ingressList, e := ingresses.List(metav1.ListOptions{})
	if e != nil {
		return "", fmt.Errorf("can not list ingresses: %s", e)
	}

	for _, ingress := range ingressList.Items {
		e := k8s.handleAddedIngress(ingress)
		if e != nil {
			logrus.Warnf("Can not add entries for ingress: %s", e)
		}
	}
	k8s.printInfo()
	go k8s.watchIngresses(ingressList.ResourceVersion)
	return pluginName, nil
}

// Stop stopps the plugin.
// It will shutdown the ingress watcher.
func (k8s *k8sIngress) Stop() error {
	k8s.watch.Stop()
	return nil
}

// watchIngresses watches for all changes of ingresses in the connected kubernetes cluster.
// It will add, update or remove the ingresses directly after it get notified about changes.
func (k8s *k8sIngress) watchIngresses(resourceVersion string) {
	clientSet, e := k8s.ctxHandler.GetClientSet()
	if e != nil {
		logrus.Errorf("can not get clientSet: %s", e)
		return
	}

	ingresses := clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)

	w, e := ingresses.Watch(metav1.ListOptions{ResourceVersion: resourceVersion})
	if e != nil {
		logrus.Errorf("Can not start watch for ingresses: %s", e)
	}
	k8s.watch = w
	for event := range w.ResultChan() {
		ingress := event.Object.(*v1beta1.Ingress)
		var e error
		switch event.Type {
		case watch.Added:
			e = k8s.handleAddedIngress(*ingress)
		case watch.Modified:
			e = k8s.handleUpdatedIngress(*ingress)
		case watch.Deleted:
			k8s.handleDeletedIngress(*ingress)
		default:
			logrus.Infof("Received unhandled event %s for ingress %s/%s", event.Type, ingress.GetNamespace(), ingress.GetName())
		}

		if e != nil {
			logrus.Warnf("Can not handle %s event: %s", event.Type, e)
		}
		k8s.printInfo()
	}
}

// handleAddedIngress adds the given ingress and adds all the host names to the dns
// backend if they point to a target.
func (k8s *k8sIngress) handleAddedIngress(ingress v1beta1.Ingress) error {
	ingressEntry := convertToIngressEntry(ingress)

	if !ingressEntry.hasTargets() {
		k8s.currentIngresses[ingressEntry.String()] = ingressEntry
		return fmt.Errorf("ingress %s has no target ip addresses", ingressEntry)
	}

	var errors *multierror.Error
	for _, host := range ingressEntry.hostNames {
		errors = multierror.Append(errors, k8s.addTargets(ingressEntry, host))
	}
	k8s.currentIngresses[ingressEntry.String()] = ingressEntry
	return errors.ErrorOrNil()
}

// handleUpdatedIngress updates the given ingress.
// It tries to change at least as possible entries.
func (k8s *k8sIngress) handleUpdatedIngress(ingress v1beta1.Ingress) error {
	ingressEntry := convertToIngressEntry(ingress)
	oldEntry, ok := k8s.currentIngresses[ingressEntry.String()]

	if !ok {
		logrus.Warnf("Can not find old entry for ingress %s. Add the new one.", ingressEntry)
		return k8s.handleAddedIngress(ingress)
	}

	if !ingressEntry.hasTargets() {
		logrus.Debugf("Ingress %s updated. It is not anymore associated to a loadbalancer. Removing all dns entries.", ingressEntry)
		for _, host := range oldEntry.hostNames {
			k8s.recordManager.RemoveHost(host)
		}
		return nil
	}

	var errors *multierror.Error
	// remove old host entries
	for _, host := range ingressEntry.getRemovedHostNames(oldEntry) {
		k8s.recordManager.RemoveHost(host)
	}

	// if targets has changed remove the updated hosts and add the new ones
	if !ingressEntry.hasSameTargets(oldEntry) {
		for _, host := range ingressEntry.getUpdatedHostNames(oldEntry) {
			k8s.recordManager.RemoveHost(host)
			errors = multierror.Append(errors, k8s.addTargets(ingressEntry, host))
		}
	}

	// add the new ones
	for _, host := range ingressEntry.getAddedHostNames(oldEntry) {
		errors = multierror.Append(errors, k8s.addTargets(ingressEntry, host))
	}
	k8s.currentIngresses[ingressEntry.String()] = ingressEntry
	return errors.ErrorOrNil()
}

// addTargets adds all targets of the ingress entry as target of the given host to the dns backend.
func (k8s *k8sIngress) addTargets(entry ingressEntry, host string) *multierror.Error {
	var errors *multierror.Error
	for _, ip := range entry.targetIps {
		errors = multierror.Append(errors, k8s.recordManager.AddHost(host, ip))
	}
	for _, target := range entry.targetHosts {
		errors = multierror.Append(errors, k8s.recordManager.AddAlias(host, target))
	}
	return errors
}

// handleDeletedIngress is always called when the ingress was deleted or updated and can
// not be reached anymore.
func (k8s *k8sIngress) handleDeletedIngress(ingress v1beta1.Ingress) {
	ingressEntry := convertToIngressEntry(ingress)

	for _, host := range ingressEntry.hostNames {
		k8s.recordManager.RemoveHost(host)
	}
	delete(k8s.currentIngresses, ingressEntry.String())
	logrus.Infof("DNS records for ingress %s successfully removed", ingressEntry)
}

func (k8s *k8sIngress) printInfo() {
	buffer := new(bytes.Buffer)
	writer := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(writer, "Name\t Namespace\t Hostname\t Targets\n")
	for _, ingress := range k8s.currentIngresses {
		fmt.Fprintf(writer, "%s\t %s\t %s\t %s\n", ingress.name, ingress.namespace, strings.Join(ingress.hostNames, ","), strings.Join(ingress.targetIps, ","))
	}
	writer.Flush()
	k8s.messageChannel <- &apis.MonitoringMessage{pluginName, buffer.String()}
}
