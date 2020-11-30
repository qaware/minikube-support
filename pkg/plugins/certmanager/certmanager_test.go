package certmanager

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	assert2 "github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	testing2 "k8s.io/client-go/testing"

	"github.com/qaware/minikube-support/pkg/github"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/packagemanager/helm"
	helmFake "github.com/qaware/minikube-support/pkg/packagemanager/helm/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
)

func TestNewCertManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name       string
		manager    helm.Manager
		handler    kubernetes.ContextHandler
		wantPlugin bool
	}{
		{"ok", helmFake.NewMockManager(ctrl), fake.NewContextHandler(k8sFake.NewSimpleClientset(), nil), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCertManager(tt.manager, tt.handler, github.NewClient())
			if _, ok := got.(*certManager); ok != tt.wantPlugin {
				t.Errorf("NewCertManager() got %v, wantPlugin = %v", got, tt.wantPlugin)
			}
		})
	}
}

func Test_certManager_Install(t *testing.T) {
	hook := test.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)
	tests := []struct {
		name               string
		addRepoError       error
		latestVersionError error
		lastLogEntry       string
	}{
		{"addRepoError", errors.New("failed to add repo"), nil, "Unable to add jetstack repository"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			helmManager := helmFake.NewMockManager(ctrl)
			m := &certManager{
				manager: helmManager,
			}
			helmManager.EXPECT().
				AddRepository("jetstack", "https://charts.jetstack.io").
				Return(tt.addRepoError)
			m.Install()

			testutils.CheckLogEntry(t, hook, tt.lastLogEntry)
		})
	}
}

func Test_certManager_Update(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	hook := test.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)
	helmInstallWaitPeriod = 0
	tests := []struct {
		name                   string
		latestVersion          string
		latestVersionError     error
		kApplyStatus           int
		repoUpdateError        error
		expectedLogEntryPrefix string
	}{
		{"ok", "1.0", nil, 0, nil, "CertSecret 'ca-issuer' successfully added"},
		{"failed update repos", "1.0", nil, 0, errors.New("no repo update"), "Unable to update helm repositories"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			helmManager := helmFake.NewMockManager(ctrl)
			handler := fake.NewContextHandler(k8sFake.NewSimpleClientset(), dynamicFake.NewSimpleDynamicClient(scheme.Scheme))
			testutils.SetTestProcessResponse(testutils.TestProcessResponse{
				Command: "mkcert", Args: []string{"-CAROOT"}, ResponseStatus: 0, Stdout: "fixtures/"},
			)

			m := &certManager{
				manager:        helmManager,
				contextHandler: handler,
				namespace:      "mks",
				values:         map[string]interface{}{},
			}

			helmManager.EXPECT().
				UpdateRepository().
				Return(tt.repoUpdateError).
				MinTimes(0).
				MaxTimes(1)
			helmManager.EXPECT().
				Install("jetstack/cert-manager", releaseName, "mks", gomock.Any(), true).
				MinTimes(0).
				MaxTimes(1)
			helmManager.EXPECT().
				GetVersion().
				Return("3")

			m.Update()

			testutils.CheckLogEntry(t, hook, tt.expectedLogEntryPrefix)
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
		handler                *fake.ContextHandler
		expectHelmUninstall    bool
		expectDeleteSecret     bool
		expectedLogEntryPrefix string
	}{
		{"ok",
			fake.NewContextHandler(k8sFake.NewSimpleClientset(&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mks",
					Name:      issuerName,
				}}), dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
				Object: map[string]interface{}{"apiVersion": "cert-manager.io/v1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}}})),
			true,
			true,
			"CertManager plugin successfully uninstalled.",
		},
		{"no secret",
			fake.NewContextHandler(k8sFake.NewSimpleClientset(), dynamicFake.NewSimpleDynamicClient(scheme.Scheme, &unstructured.Unstructured{
				Object: map[string]interface{}{"apiVersion": "cert-manager.io/v1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}}})),
			true,
			true,
			"Unable to uninstall the certManager plugin: 1 error occurred:\n\t* secrets \"ca-issuer\" not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := helmFake.NewMockManager(ctrl)
			m := NewCertManager(manager, tt.handler, github.NewClient())
			if tt.expectHelmUninstall {
				manager.EXPECT().Uninstall(releaseName, "mks", true)
			}
			m.Uninstall(true)

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
		existingSecret *corev1.Secret
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
			&corev1.Secret{
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
				testutils.SetTestProcessResponse(testutils.TestProcessResponse{Command: "mkcert", Args: []string{"-CAROOT"}, ResponseStatus: 0, Stdout: tt.mkcertRoot})
			} else {
				testutils.SetTestProcessResponse(testutils.TestProcessResponse{Command: "mkcert", Args: []string{"-CAROOT"}, ResponseStatus: 1})
			}

			var fakeClientSet *k8sFake.Clientset
			if tt.existingSecret != nil {
				fakeClientSet = k8sFake.NewSimpleClientset(tt.existingSecret)
			} else {
				fakeClientSet = k8sFake.NewSimpleClientset()
			}
			o := NewCertManager(nil, fake.NewContextHandler(fakeClientSet, nil), github.NewClient())
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
			Object: map[string]interface{}{"apiVersion": "cert-manager.io/v1", "kind": "ClusterIssuer", "metadata": map[string]interface{}{"name": issuerName}},
		}), "update", false},
		{"no client", nil, "create", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := fake.NewContextHandler(k8sFake.NewSimpleClientset(), tt.dynamicClient)
			o := NewCertManager(nil, handler, github.NewClient())
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
