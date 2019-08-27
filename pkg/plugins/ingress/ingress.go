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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"text/tabwriter"
)

type k8sIngress struct {
	ctxHandler     kubernetes.ContextHandler
	messageChannel chan *apis.MonitoringMessage
	recordManager  coredns.Manager
	watch          *kubernetes.Watcher

	currentEntries map[string]*entry
}

const pluginName = "kubernetes-ingress"

// NewK8sIngress will initialize a new ingress plugin.
// It allows to configure the functions to add and remove the hosts in the dns backend.
func NewK8sIngress(contextHandler kubernetes.ContextHandler, recordManager coredns.Manager) apis.StartStopPlugin {
	if recordManager == nil {
		recordManager = coredns.NewNoOpManager()
	}

	return &k8sIngress{
		ctxHandler:     contextHandler,
		recordManager:  recordManager,
		currentEntries: make(map[string]*entry),
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
		e := k8s.AddedEvent(ingress.DeepCopyObject())
		if e != nil {
			logrus.Warnf("Can not add entries for ingress: %s", e)
		}
	}
	_ = k8s.PostEvent()

	k8s.watch, e = kubernetes.NewWatcher(k8s, nil, ingressList.ResourceVersion)
	if e != nil {
		return "", fmt.Errorf("can not start watcher: %s", e)
	}

	k8s.watch.Start()
	return pluginName, nil
}

// Stop stopps the plugin.
// It will shutdown the ingress watcher.
func (k8s *k8sIngress) Stop() error {
	k8s.watch.Stop()
	return nil
}

func (k8s *k8sIngress) PreWatch(options metav1.ListOptions) (watch.Interface, error) {
	clientSet, e := k8s.ctxHandler.GetClientSet()
	if e != nil {
		return nil, fmt.Errorf("can not get clientSet: %s", e)
	}

	ingresses := clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)

	return ingresses.Watch(options)
}

// AddedEvent adds the given ingress and adds all the host names to the dns
// backend if they point to a target.
func (k8s *k8sIngress) AddedEvent(obj runtime.Object) error {
	entry, e := convertObjectToEntry(obj)
	if e != nil {
		return e
	}

	if !entry.hasTargets() {
		k8s.currentEntries[entry.String()] = entry
		return fmt.Errorf("%s %s has no target ip addresses", entry.typ, entry)
	}

	var errors *multierror.Error
	for _, host := range entry.hostNames {
		errors = multierror.Append(errors, k8s.addTargets(entry, host))
	}
	k8s.currentEntries[entry.String()] = entry
	return errors.ErrorOrNil()
}

// UpdatedEvent updates the given ingress.
// It tries to change at least as possible entries.
func (k8s *k8sIngress) UpdatedEvent(obj runtime.Object) error {
	entry, e := convertObjectToEntry(obj)
	if e != nil {
		return e
	}
	oldEntry, ok := k8s.currentEntries[entry.String()]

	if !ok {
		logrus.Warnf("Can not find old entry for %s %s. Add the new one.", entry.typ, entry)
		return k8s.AddedEvent(obj)
	}

	if !entry.hasTargets() {
		logrus.Debugf("%s %s updated. It is not anymore associated to a loadbalancer. Removing all dns entries.", entry.typ, entry)
		for _, host := range oldEntry.hostNames {
			k8s.recordManager.RemoveHost(host)
		}
		return nil
	}

	var errors *multierror.Error
	// remove old host entries
	for _, host := range entry.getRemovedHostNames(oldEntry) {
		k8s.recordManager.RemoveHost(host)
	}

	// if targets has changed remove the updated hosts and add the new ones
	if !entry.hasSameTargets(oldEntry) {
		for _, host := range entry.getUpdatedHostNames(oldEntry) {
			k8s.recordManager.RemoveHost(host)
			errors = multierror.Append(errors, k8s.addTargets(entry, host))
		}
	}

	// add the new ones
	for _, host := range entry.getAddedHostNames(oldEntry) {
		errors = multierror.Append(errors, k8s.addTargets(entry, host))
	}
	k8s.currentEntries[entry.String()] = entry
	return errors.ErrorOrNil()
}

// addTargets adds all targets of the ingress entry as target of the given host to the dns backend.
func (k8s *k8sIngress) addTargets(entry *entry, host string) *multierror.Error {
	var errors *multierror.Error
	for _, ip := range entry.targetIps {
		errors = multierror.Append(errors, k8s.recordManager.AddHost(host, ip))
	}
	for _, target := range entry.targetHosts {
		errors = multierror.Append(errors, k8s.recordManager.AddAlias(host, target))
	}
	return errors
}

// DeletedEvent is always called when the ingress was deleted or updated and can
// not be reached anymore.
func (k8s *k8sIngress) DeletedEvent(obj runtime.Object) error {
	entry, e := convertObjectToEntry(obj)
	if e != nil {
		return e
	}

	for _, host := range entry.hostNames {
		k8s.recordManager.RemoveHost(host)
	}
	delete(k8s.currentEntries, entry.String())
	logrus.Infof("DNS records for %s %s successfully removed", entry.typ, entry)
	return nil
}

// PostEvent generates a short overview about the currently handled ingresses and services.
func (k8s *k8sIngress) PostEvent() error {
	var errors *multierror.Error
	buffer := new(bytes.Buffer)
	writer := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.Debug)
	_, e := fmt.Fprintf(writer, "Name\t Namespace\t Typ\t Hostname\t Targets\n")
	errors = multierror.Append(errors, e)

	for _, entry := range k8s.currentEntries {

		_, e := fmt.Fprintf(writer,
			"%s\t %s\t %s\t %s\t %s\n",
			entry.name,
			entry.namespace,
			entry.typ,
			strings.Join(entry.hostNames, ","),
			strings.Join(entry.targetIps, ","))

		errors = multierror.Append(errors, e)
	}

	errors = multierror.Append(errors, writer.Flush())
	if errors.Len() == 0 {
		k8s.messageChannel <- &apis.MonitoringMessage{Box: pluginName, Message: buffer.String()}
	}
	return errors.ErrorOrNil()
}
