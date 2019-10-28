package k8sdns

import (
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
	ingresses := s.clientSet.
		CoreV1().
		Services(v1.NamespaceAll)
	serviceList, e := ingresses.List(metav1.ListOptions{})
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
	return services.Watch(options)
}
