package ingress

import (
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			e := ingressEntry{
				name:      tt.ingressName,
				namespace: tt.namespace,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("ingressEntry.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToIngressEntry(t *testing.T) {
	tests := []struct {
		name    string
		ingress v1beta1.Ingress
		want    ingressEntry
	}{
		{
			"full",
			v1beta1.Ingress{
				ObjectMeta: v1meta.ObjectMeta{Name: "test", Namespace: "test-ns"},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{{Host: "1"}},
					TLS:   []v1beta1.IngressTLS{{Hosts: []string{"2"}}},
				},
				Status: v1beta1.IngressStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: "ip", Hostname: "host"}}}},
			},
			ingressEntry{
				name:        "test",
				namespace:   "test-ns",
				hostNames:   []string{"1", "2"},
				targetIps:   []string{"ip"},
				targetHosts: []string{"host"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToIngressEntry(tt.ingress); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToIngressEntry() = %v, want %v", got, tt.want)
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
			e := ingressEntry{
				targetIps:   tt.targetIps,
				targetHosts: tt.targetHosts,
			}
			if got := e.hasTargets(); got != tt.want {
				t.Errorf("ingressEntry.hasTargets() = %v, want %v", got, tt.want)
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
			e := ingressEntry{
				hostNames: tt.hostNames,
			}
			if got := e.getAddedHostNames(ingressEntry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ingressEntry.getAddedHostNames() = %v, want %v", got, tt.want)
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
			e := ingressEntry{
				hostNames: tt.hostNames,
			}
			if got := e.getUpdatedHostNames(ingressEntry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ingressEntry.getUpdatedHostNames() = %v, want %v", got, tt.want)
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
			e := ingressEntry{hostNames: tt.hostNames}
			if got := e.getRemovedHostNames(ingressEntry{hostNames: tt.oldHostNames}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ingressEntry.getRemovedHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
