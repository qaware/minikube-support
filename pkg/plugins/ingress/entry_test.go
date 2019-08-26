package ingress

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	"k8s.io/api/extensions/v1beta1"
)

func Test_ingressEntry_String(t *testing.T) {
	tests := []struct {
		ingressName string
		namespace   string
		want        string
	}{
		{"test", "default", "default/test"},
	}
	for _, tt := range tests {
		t.Run(tt.ingressName, func(t *testing.T) {
			e := entry{
				name:      tt.ingressName,
				namespace: tt.namespace,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("entry.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertObjectToEntry(t *testing.T) {
	tests := []struct {
		name    string
		obj     runtime.Object
		want    *entry
		wantErr bool
	}{
		{
			"ingress full",
			&v1beta1.Ingress{
				ObjectMeta: v1meta.ObjectMeta{Name: "test", Namespace: "test-ns"},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{{Host: "1"}},
					TLS:   []v1beta1.IngressTLS{{Hosts: []string{"1"}}},
				},
				Status: v1beta1.IngressStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: "ip", Hostname: "host"}}}},
			},
			&entry{
				name:        "test",
				namespace:   "test-ns",
				typ:         "Ingress",
				hostNames:   []string{"1"},
				targetIps:   []string{"ip"},
				targetHosts: []string{"host"},
			},
			false,
		}, {
			"service full",
			&v1.Service{
				ObjectMeta: v1meta.ObjectMeta{Name: "test", Namespace: "test-ns"},
				Status:     v1.ServiceStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: "ip", Hostname: "host"}}}},
			},
			&entry{
				name:        "test",
				namespace:   "test-ns",
				typ:         "Service",
				hostNames:   []string{"test.test-ns.svc.minikube."},
				targetIps:   []string{"ip"},
				targetHosts: []string{"host"},
			},
			false,
		}, {
			"invalid obj",
			&v1.Pod{},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertObjectToEntry(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertObjectToEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertObjectToEntry() got = %v, want %v", got.debug(), tt.want.debug())
			}
		})
	}
}

func Test_ingressEntry_hasTargets(t *testing.T) {
	tests := []struct {
		name        string
		targetIps   []string
		targetHosts []string
		want        bool
	}{
		{"none", []string{}, []string{}, false},
		{"ip only", []string{"127.0.01"}, []string{}, true},
		{"host only", []string{}, []string{"localhost"}, true},
		{"both", []string{"127.0.0.1"}, []string{"localhost"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				targetIps:   tt.targetIps,
				targetHosts: tt.targetHosts,
			}
			if got := e.hasTargets(); got != tt.want {
				t.Errorf("entry.hasTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getAddedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, nil},
		{"one new", []string{"1", "2"}, []string{"1"}, []string{"2"}},
		{"old empty", []string{"1", "2"}, []string{}, []string{"1", "2"}},
		{"new empty", []string{}, []string{"1", "2"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				hostNames: tt.hostNames,
			}
			if got := e.getAddedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getAddedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getUpdatedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, []string{"1", "2"}},
		{"one new", []string{"1", "2"}, []string{"1"}, []string{"1"}},
		{"old empty", []string{"1", "2"}, []string{}, nil},
		{"new empty", []string{}, []string{"1", "2"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{
				hostNames: tt.hostNames,
			}
			if got := e.getUpdatedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getUpdatedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressEntry_getRemovedHostNames(t *testing.T) {
	tests := []struct {
		name         string
		hostNames    []string
		oldHostNames []string
		want         []string
	}{
		{"no new", []string{"1", "2"}, []string{"1", "2"}, nil},
		{"one new", []string{"1", "2"}, []string{"1"}, nil},
		{"one old", []string{"1"}, []string{"1", "2"}, []string{"2"}},
		{"old empty", []string{"1", "2"}, []string{}, nil},
		{"new empty", []string{}, []string{"1", "2"}, []string{"1", "2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := entry{hostNames: tt.hostNames}
			if got := e.getRemovedHostNames(&entry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entry.getRemovedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Improves the debug output when tests fail
func (e *entry) debug() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"%s %s {hostNames=%s, targetIps=%s, targetHosts=%s}",
		e.typ,
		e.String(), e.hostNames, e.targetIps, e.targetHosts)
}
