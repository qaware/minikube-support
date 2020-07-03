package k8sdns

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// serviceAccessor provides list and watch access to services.
type serviceAccessor struct {
	clientSet kubernetes.Interface
}

// PreFetch returns a list of all services and the corresponding list interface.
func (s serviceAccessor) PreFetch() ([]runtime.Object, metav1.ListInterface, error) {
	services := s.clientSet.
		CoreV1().
		Services(v1.NamespaceAll)
	serviceList, e := services.List(context.Background(), metav1.ListOptions{})
	if e != nil {
		return nil, nil, fmt.Errorf("can not list services: %s", e)
	}
	var items []runtime.Object
	for _, service := range serviceList.Items {
		items = append(items, &service)
	}

	return items, serviceList, nil
}

// Watch starts the watch process for services.
func (s serviceAccessor) Watch(options metav1.ListOptions) (watch.Interface, error) {
	services := s.clientSet.
		CoreV1().
		Services(v1.NamespaceAll)
	return services.Watch(context.Background(), options)
}

// ConvertToEntry converts a k8s service into the entry by flatten everything.
func (serviceAccessor) ConvertToEntry(obj runtime.Object) (*entry, error) {
	service, ok := obj.(*v1.Service)
	if !ok {
		return nil, fmt.Errorf("can not convert non service object into service")
	}
	return &entry{
		name:        service.Name,
		namespace:   service.Namespace,
		typ:         "Service",
		hostNames:   []string{fmt.Sprintf("%s.%s.svc.minikube.", service.Name, service.Namespace)},
		targetIps:   getLoadBalancerIps(service.Status.LoadBalancer),
		targetHosts: getLoadBalancerHostNames(service.Status.LoadBalancer),
	}, nil
}

// MatchesPreconditions checks if the given object matches preconditions for adding the entry.
func (serviceAccessor) MatchesPreconditions(obj runtime.Object) bool {
	service, ok := obj.(*v1.Service)
	if !ok {
		return false
	}
	return service.Spec.Type == v1.ServiceTypeLoadBalancer || service.Spec.Type == v1.ServiceTypeExternalName
}
