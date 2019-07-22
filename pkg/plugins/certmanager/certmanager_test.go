package certmanager

import (
	"github.com/magiconair/properties/assert"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"os/exec"
	"testing"
)

func TestNewCertManager(t *testing.T) {
	tests := []struct {
		name       string
		manager    helm.Manager
		handler    kubernetes.ContextHandler
		wantPlugin bool
		wantErr    bool
	}{
		{"ok", helm.NewHelmManager(), fake.NewContextHandler(k8sFake.NewSimpleClientset(), nil), true, false},
		{"no clientset", helm.NewHelmManager(), fake.NewContextHandler(nil, nil), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCertManager(tt.manager, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCertManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if _, ok := got.(*certManager); ok != tt.wantPlugin {
				t.Errorf("NewCertManager() got %v, wantPlugin = %v", got, tt.wantPlugin)
			}
		})
	}
}

func Test_certManager_applyCertSecret(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		existingSecret *v1.Secret
		mkcertRoot     string
		wantAction     string
		wantErr        bool
	}{
		{"ok, create",
			nil,
			"fixtures/",
			"create",
			false,
		},
		{"ok, update",
			&v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mks",
					Name:      issuerName,
				},
			},
			"fixtures/",
			"update",
			false,
		},
		{"failed mkcert",
			nil,
			"",
			"",
			true,
		},
		{"no cert",
			nil,
			"invailid-fixtures/",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mkcertRoot != "" {
				testutils.TestProcessResponses = []testutils.TestProcessResponse{{Command: "mkcert", Args: []string{"-CAROOT"}, ResponseStatus: 0, Stdout: tt.mkcertRoot}}
			} else {
				testutils.TestProcessResponses = []testutils.TestProcessResponse{{Command: "mkcert", Args: []string{"-CAROOT"}, ResponseStatus: 1}}
			}

			var fakeClientSet *k8sFake.Clientset
			if tt.existingSecret != nil {
				fakeClientSet = k8sFake.NewSimpleClientset(tt.existingSecret)
			} else {
				fakeClientSet = k8sFake.NewSimpleClientset()
			}
			o, _ := NewCertManager(nil, fake.NewContextHandler(fakeClientSet, nil))
			m := o.(*certManager)
			if err := m.applyCertSecret(); (err != nil) != tt.wantErr {
				t.Errorf("certManager.applyCertSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
			actions := fakeClientSet.Actions()
			if len(actions) > 1 {
				assert.Equal(t, tt.wantAction, actions[1].GetVerb())
			}
		})
	}
}

func Test_certManager_applyClusterIssuer(t *testing.T) {
	tests := []struct {
		name          string
		dynamicClient *dynamicFake.FakeDynamicClient
		action        string
		wantErr       bool
	}{
		{"ok, create", dynamicFake.NewSimpleDynamicClient(scheme.Scheme), "create", false},
		{"ok, update", dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
			Object: map[string]interface{}{"apiVersion": "certmanager.k8s.io/v1alpha1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}},
		}), "update", false},
		{"no client", nil, "create", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := fake.NewContextHandler(k8sFake.NewSimpleClientset(), tt.dynamicClient)
			o, _ := NewCertManager(nil, handler)
			m := o.(*certManager)

			if err := m.applyClusterIssuer(); (err != nil) != tt.wantErr {
				t.Errorf("certManager.applyClusterIssuer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.dynamicClient != nil && len(tt.dynamicClient.Actions()) > 1 {
				assert.Equal(t, tt.action, tt.dynamicClient.Actions()[1].GetVerb())
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}
