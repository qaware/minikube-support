package certmanager

import (
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	assert2 "github.com/stretchr/testify/assert"
	testing2 "k8s.io/client-go/testing"

	"github.com/magiconair/properties/assert"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	fake2 "github.com/qaware/minikube-support/pkg/packagemanager/helm/fake"
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

func Test_certManager_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	hook := test.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)
	tests := []struct {
		name                   string
		purge                  bool
		handler                *fake.ContextHandler
		expectHelmUninstall    bool
		expectDeleteSecret     bool
		expectDeleteIssuer     bool
		expectedLogEntryPrefix string
	}{
		{"ok",
			true,
			fake.NewContextHandler(k8sFake.NewSimpleClientset(&v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mks",
					Name:      issuerName,
				}}), dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
				Object: map[string]interface{}{"apiVersion": "certmanager.k8s.io/v1alpha1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}}})),
			true,
			true,
			true,
			"CertManager plugin successfully uninstalled.",
		},
		{"ok no purge",
			false,
			fake.NewContextHandler(k8sFake.NewSimpleClientset(&v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mks",
					Name:      issuerName,
				}}), dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
				Object: map[string]interface{}{"apiVersion": "certmanager.k8s.io/v1alpha1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}}})),
			true,
			true,
			true,
			"CertManager plugin successfully uninstalled.",
		},
		{"no secret",
			false,
			fake.NewContextHandler(k8sFake.NewSimpleClientset(), dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
				Object: map[string]interface{}{"apiVersion": "certmanager.k8s.io/v1alpha1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}}})),
			true,
			true,
			true,
			"Unable to uninstall the certManager plugin: 1 error occurred:\n\t* secrets \"ca-issuer\" not found",
		},
		{"no issuer",
			false,
			fake.NewContextHandler(k8sFake.NewSimpleClientset(&v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mks",
					Name:      issuerName,
				}}), dynamicFake.NewSimpleDynamicClient(scheme.Scheme)),
			true,
			true,
			true,
			"Unable to uninstall the certManager plugin: 1 error occurred:\n\t* clusterissuers.certmanager.k8s.io \"ca-issuer\" not found",
		},
		{"no dynamic client",
			false,
			fake.NewContextHandler(k8sFake.NewSimpleClientset(), nil),
			false,
			false,
			false,
			"unable to get dynamic client: no dynamic client",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := fake2.NewMockManager(ctrl)
			m, _ := NewCertManager(manager, tt.handler)
			if tt.expectHelmUninstall {
				manager.EXPECT().Uninstall(releaseName, tt.purge)
			}
			m.Uninstall(tt.purge)

			if tt.expectDeleteIssuer {
				verifyActionResource(t, tt.handler.DynamicClient.Actions(), 0, "delete", "clusterissuers")
			}
			if tt.expectDeleteSecret {
				verifyActionResource(t, tt.handler.ClientSet.Actions(), 0, "delete", "secrets")
			}
			testutils.CheckLogEntry(t, hook, tt.expectedLogEntryPrefix)
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
			if tt.dynamicClient != nil {
				verifyActionResource(t, tt.dynamicClient.Actions(), 1, tt.action, "clusterissuers")
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}

func verifyActionResource(t *testing.T, actions []testing2.Action, item int, verb string, resource string) {
	assert2.True(t, len(actions) >= item)
	assert2.True(t, item >= 0)
	assert2.Truef(t, actions[item].Matches(verb, resource), "action %v did not match %v %v", actions[item], verb, resource)
}
