package k8sdns

import (
	"context"
	"fmt"

	networkingV1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// ingressAccessor provides list and watch access to ingresses.
type ingressAccessor struct {
	clientSet kubernetes.Interface
}

// PreFetch returns a list of all ingresses and the corresponding list interface.
func (i ingressAccessor) PreFetch() ([]runtime.Object, metaV1.ListInterface, error) {
	ingresses := i.clientSet.
		NetworkingV1().
		Ingresses(metaV1.NamespaceAll)
	ingressList, e := ingresses.List(context.Background(), metaV1.ListOptions{})
	if e != nil {
		return nil, nil, fmt.Errorf("can not list ingresses: %s", e)
	}
	var items []runtime.Object
	for _, ingress := range ingressList.Items {
		items = append(items, &ingress)
	}

	return items, ingressList, nil
}

// Watch starts the watch process for ingresses.
func (i ingressAccessor) Watch(options metaV1.ListOptions) (watch.Interface, error) {
	ingresses := i.clientSet.
		NetworkingV1().
		Ingresses(metaV1.NamespaceAll)

	return ingresses.Watch(context.Background(), options)
}

// ConvertToEntry converts a k8s ingress into the entry by flatten everything.
func (ingressAccessor) ConvertToEntry(obj runtime.Object) (*entry, error) {
	ingress, ok := obj.(*networkingV1.Ingress)
	if !ok {
		return nil, fmt.Errorf("can not convert non ingress object into ingress")
	}
	return &entry{
		name:        ingress.Name,
		namespace:   ingress.Namespace,
		typ:         "Ingress",
		hostNames:   getHostNames(ingress),
		targetIps:   getLoadBalancerIps(ingress.Status.LoadBalancer),
		targetHosts: getLoadBalancerHostNames(ingress.Status.LoadBalancer),
	}, nil
}

// MatchesPreconditions checks if the given object matches preconditions for adding the entry.
func (ingressAccessor) MatchesPreconditions(obj runtime.Object) bool {
	_, ok := obj.(*networkingV1.Ingress)
	return ok
}
