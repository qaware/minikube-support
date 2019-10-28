package k8sdns

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"testing"
)

func Test_serviceAccessor_PreFetch(t *testing.T) {
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
			[]runtime.Object{&v1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "extension/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: v1.ServiceSpec{
					Type: "ClusterIP",
				},
			}},
			&v1.ServiceList{Items: []v1.Service{{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "extension/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: v1.ServiceSpec{
					Type: "ClusterIP",
				},
			}}},
			false,
		}, {
			"nok",
			true,
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset(&v1.ServiceList{Items: []v1.Service{{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "extension/v1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "abc"},
				Spec: v1.ServiceSpec{
					Type: "ClusterIP",
				},
			}}})

			i := serviceAccessor{
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

func Test_serviceAccessor_Watch(t *testing.T) {
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

			i := serviceAccessor{
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
