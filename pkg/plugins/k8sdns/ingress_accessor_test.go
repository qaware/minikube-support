package k8sdns

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
)

func Test_ingressAccessor_PreFetch(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		want       []runtime.Object
		want1      metav1.ListInterface
		wantErr    bool
	}{
		{
			"ok",
			false,
			[]runtime.Object{&networkingV1.Ingress{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Ingress",
					APIVersion: "networking/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: networkingV1.IngressSpec{
					DefaultBackend: &networkingV1.IngressBackend{
						Service: &networkingV1.IngressServiceBackend{Name: "dummy"},
					},
				},
			}},
			&networkingV1.IngressList{Items: []networkingV1.Ingress{{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Ingress",
					APIVersion: "networking/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: networkingV1.IngressSpec{
					DefaultBackend: &networkingV1.IngressBackend{
						Service: &networkingV1.IngressServiceBackend{Name: "dummy"},
					},
				},
			}}},
			false,
		},
		{
			"nok",
			true,
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset(&networkingV1.IngressList{Items: []networkingV1.Ingress{{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Ingress",
					APIVersion: "networking/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: networkingV1.IngressSpec{
					DefaultBackend: &networkingV1.IngressBackend{
						Service: &networkingV1.IngressServiceBackend{Name: "dummy"},
					},
				},
			}}})

			i := ingressAccessor{
				clientSet: cs,
			}

			cs.Fake.PrependReactor("*", "*", func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
				if tt.shouldFail {
					return true, nil, errors.New("dummy error")
				}
				return false, nil, nil
			})

			got, got1, err := i.PreFetch()
			if (err != nil) != tt.wantErr {
				t.Errorf("PreFetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func Test_ingressAccessor_Watch(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		wantErr    bool
	}{
		{"ok", false, false},
		{"nok", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()

			i := ingressAccessor{
				clientSet: cs,
			}
			cs.Fake.PrependWatchReactor("*", func(action testing2.Action) (handled bool, ret watch.Interface, err error) {
				if tt.shouldFail {
					return true, nil, errors.New("dummy error")
				}
				return false, nil, nil
			})

			got, err := i.Watch(metav1.ListOptions{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Watch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.shouldFail {
				assert.Implements(t, (*watch.Interface)(nil), got)
			}
		})
	}
}

func Test_ingressAccessor_ConvertToEntry(t *testing.T) {
	tests := []struct {
		name    string
		obj     runtime.Object
		want    *entry
		wantErr bool
	}{
		{
			"ingress full",
			&networkingV1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test-ns"},
				Spec: networkingV1.IngressSpec{
					Rules: []networkingV1.IngressRule{{Host: "1"}},
					TLS:   []networkingV1.IngressTLS{{Hosts: []string{"1"}}},
				},
				Status: networkingV1.IngressStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: "ip", Hostname: "host"}}}},
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
		},
		{
			"invalid obj",
			&v1.Pod{},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ingressAccessor{}
			got, err := in.ConvertToEntry(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_ingressAccessor_MatchesPreconditions(t *testing.T) {
	tests := []struct {
		name string
		obj  runtime.Object
		want bool
	}{
		{
			"ingress full",
			&networkingV1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test-ns"},
				Spec: networkingV1.IngressSpec{
					Rules: []networkingV1.IngressRule{{Host: "1"}},
					TLS:   []networkingV1.IngressTLS{{Hosts: []string{"1"}}},
				},
				Status: networkingV1.IngressStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: []v1.LoadBalancerIngress{{IP: "ip", Hostname: "host"}}}},
			},
			true,
		}, {
			"invalid obj",
			&v1.Pod{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ingressAccessor{}
			if got := in.MatchesPreconditions(tt.obj); got != tt.want {
				t.Errorf("MatchesPreconditions() = %v, want %v", got, tt.want)
			}
		})
	}
}
