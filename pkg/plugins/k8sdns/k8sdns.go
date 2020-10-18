package k8sdns

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/plugins/coredns"
	"github.com/qaware/minikube-support/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
)

type k8sDns struct {
	ctxHandler     kubernetes.ContextHandler
	messageChannel chan *apis.MonitoringMessage
	recordManager  coredns.Manager
	watch          *kubernetes.Watcher
	accessType     AccessType
	accessor       accessor

	currentEntries map[string]*entry
}

type AccessType string

// accessor is an abstraction for accessing different types of k8s objects. For example services vs. ingresses.
type accessor interface {
	// PreFetch returns a list of all services, ingresses or other k8s objects and the corresponding list interface.
	PreFetch() ([]runtime.Object, metav1.ListInterface, error)

	// Watch starts the watch process for services, ingresses or other k8s objects.
	Watch(options metav1.ListOptions) (watch.Interface, error)

	// ConvertToEntry converts the given object into the entry by flatten everything.
	ConvertToEntry(runtime.Object) (*entry, error)

	// MatchesPreconditions checks if the given object matches preconditions for adding the entry.
	MatchesPreconditions(object runtime.Object) bool
}

const AccessTypeIngress = AccessType("ingress")
const AccessTypeService = AccessType("service")

const pluginName = "k8sdns-"

// NewK8sDns will initialize a new ingress plugin.
// It allows to configure the functions to add and remove the hosts in the dns backend.
func NewK8sDns(contextHandler kubernetes.ContextHandler, recordManager coredns.Manager, accessType AccessType) apis.StartStopPlugin {
	if recordManager == nil {
		recordManager = coredns.NewNoOpManager()
	}

	return &k8sDns{
		ctxHandler:     contextHandler,
		recordManager:  recordManager,
		accessType:     accessType,
		currentEntries: make(map[string]*entry),
	}
}

func (*k8sDns) IsSingleRunnable() bool {
	return true
}

func (k8s *k8sDns) String() string {
	return pluginName + string(k8s.accessType)
}

// Start starts the ingress plugin. It will automatically add all current ingresses.
func (k8s *k8sDns) Start(messageChannel chan *apis.MonitoringMessage) (string, error) {
	k8s.messageChannel = messageChannel

	clientSet, e := k8s.ctxHandler.GetClientSet()
	if e != nil {
		return "", fmt.Errorf("can not get clientSet: %s", e)
	}

	switch k8s.accessType {
	case AccessTypeIngress:
		k8s.accessor = ingressAccessor{clientSet: clientSet}
	case AccessTypeService:
		k8s.accessor = serviceAccessor{clientSet: clientSet}
	default:
		return "", fmt.Errorf("invalid access type given: %s", k8s.accessType)
	}

	objects, list, e := k8s.accessor.PreFetch()
	if e != nil {
		return "", e
	}

	for _, element := range objects {
		e := k8s.AddedEvent(element)
		if e != nil {
			logrus.Warnf("Can not add entries for %s: %s", k8s.accessType, e)
		}
	}
	_ = k8s.PostEvent()

	k8s.watch, e = kubernetes.NewWatcher(k8s, nil, list.GetResourceVersion())
	if e != nil {
		return "", fmt.Errorf("can not start watcher: %s", e)
	}

	k8s.watch.Start()
	return k8s.String(), nil
}

// Stop stopps the plugin.
// It will shutdown the ingress watcher.
func (k8s *k8sDns) Stop() error {
	k8s.watch.Stop()
	return nil
}

func (k8s *k8sDns) PreWatch(options metav1.ListOptions) (watch.Interface, error) {
	return k8s.accessor.Watch(options)
}

// AddedEvent adds the given ingress and adds all the host names to the dns
// backend if they point to a target.
func (k8s *k8sDns) AddedEvent(obj runtime.Object) error {
	if !k8s.accessor.MatchesPreconditions(obj) {
		logrus.Debugf("%s don't matches the preconditions. Will not add the entry.", obj.GetObjectKind().GroupVersionKind().String())
		return nil
	}

	entry, e := k8s.accessor.ConvertToEntry(obj)
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
	//noinspection GoNilness
	return errors.ErrorOrNil()
}

// UpdatedEvent updates the given ingress.
// It tries to change at least as possible entries.
func (k8s *k8sDns) UpdatedEvent(obj runtime.Object) error {
	entry, e := k8s.accessor.ConvertToEntry(obj)
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
	//noinspection GoNilness
	return errors.ErrorOrNil()
}

// addTargets adds all targets of the ingress entry as target of the given host to the dns backend.
func (k8s *k8sDns) addTargets(entry *entry, host string) *multierror.Error {
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
func (k8s *k8sDns) DeletedEvent(obj runtime.Object) error {
	entry, e := k8s.accessor.ConvertToEntry(obj)
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
func (k8s *k8sDns) PostEvent() error {
	var entryStrings []string
	for _, entry := range k8s.currentEntries {
		entryStrings = append(entryStrings, fmt.Sprintf("%s\t %s\t %s\t %s\t %s\n",
			entry.name,
			entry.namespace,
			entry.typ,
			strings.Join(entry.hostNames, ","),
			strings.Join(entry.targetIps, ",")))
	}

	table, e := utils.FormatAsTable(entryStrings, "Name\t Namespace\t Typ\t Hostname\t Targets\n")
	if e != nil {
		return e
	}

	k8s.messageChannel <- &apis.MonitoringMessage{Box: k8s.String(), Message: table}
	return nil
}
