package k8sdns

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// ingressAccessor provides list and watch access to ingresses.
type ingressAccessor struct {
	clientSet kubernetes.Interface
}

// PreFetch returns a list of all ingresses and the corresponding list interface.
func (i ingressAccessor) PreFetch() ([]runtime.Object, v1.ListInterface, error) {
	ingresses := i.clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)
	ingressList, e := ingresses.List(metav1.ListOptions{})
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
func (i ingressAccessor) Watch(options v1.ListOptions) (watch.Interface, error) {
	ingresses := i.clientSet.
		ExtensionsV1beta1().
		Ingresses(v1.NamespaceAll)

	return ingresses.Watch(options)
}
