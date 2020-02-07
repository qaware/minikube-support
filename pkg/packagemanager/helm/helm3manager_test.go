package helm

import (
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"os/exec"
	"testing"

	k8sFake "k8s.io/client-go/kubernetes/fake"
)

func Test_helm3Manager_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		chart          string
		release        string
		namespace      string
		wait           bool
		values         map[string]interface{}
		expectedArgs   []string
		nsExists       bool
		response       string
		responseStatus int
		lastEntryLevel logrus.Level
	}{
		{
			"success",
			"dummy/test",
			"test",
			"test",
			false,
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test"},
			true,
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"success but namespace is missing",
			"dummy/test",
			"test",
			"test",
			false,
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test"},
			false,
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"wait for success",
			"dummy/test",
			"test",
			"test",
			true,
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test", "--wait"},
			true,
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"success with values",
			"dummy/test",
			"test",
			"test",
			false,
			map[string]interface{}{"v1": []map[string]interface{}{{"h": 2, "b": "def"}}},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test", "--set", "v1\\[0].h=2", "--set", "v1\\[0].b=def"},
			true,
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"missing name and chart",
			"",
			"",
			"test",
			false,
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "", ""},
			true,
			"no release and name given",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fakeClientSet *k8sFake.Clientset
			if tt.nsExists {
				fakeClientSet = k8sFake.NewSimpleClientset(&v1.Namespace{
					TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{Name: tt.namespace},
				})
			} else {
				fakeClientSet = k8sFake.NewSimpleClientset()
			}

			m := &helm3Manager{
				context: fake.NewContextHandler(fakeClientSet, nil),
			}

			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}

			m.Install(tt.chart, tt.release, tt.namespace, tt.values, tt.wait)

			lastEntry := global.LastEntry()
			if !assert.Equal(t, tt.lastEntryLevel, lastEntry.Level) {
				s, _ := lastEntry.String()
				t.Log(s)
			}
		})
	}
}

func Test_helm3Manager_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		release        string
		purge          bool
		expectedArgs   []string
		response       string
		responseStatus int
		lastEntryLevel logrus.Level
	}{
		{
			"success no purge",
			"test",
			false,
			[]string{"uninstall", "--keep-history", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"success purge",
			"test",
			true,
			[]string{"uninstall", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"not found",
			"test",
			false,
			[]string{"uninstall", "--keep-history", "test"},
			"not found",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm3Manager{
				context: fake.NewContextHandler(nil, nil),
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}
			m.Uninstall(tt.release, tt.purge)
			lastEntry := global.LastEntry()
			if lastEntry.Level != tt.lastEntryLevel {
				t.Errorf("Expected log level of last entry %s but was %s", tt.lastEntryLevel, lastEntry.Level)
			}
		})
	}
}

func Test_helm3Manager_AddRepository(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		repoName       string
		url            string
		expectedArgs   []string
		responseStatus int
		wantErr        bool
	}{
		{"ok", "dummy", "http://localhost", []string{"repo", "add", "dummy", "http://localhost"}, 0, false},
		{"failed", "dummy", "http://localhost", []string{"repo", "add", "dummy", "http://localhost"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm3Manager{
				context: fake.NewContextHandler(nil, nil),
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: ""},
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}

			if err := m.AddRepository(tt.repoName, tt.url); (err != nil) != tt.wantErr {
				t.Errorf("AddRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_helm3Manager_UpdateRepository(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		expectedArgs   []string
		responseStatus int
		wantErr        bool
	}{
		{"ok", []string{"repo", "update"}, 0, false},
		{"failed", []string{"repo", "update"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm3Manager{
				context: fake.NewContextHandler(nil, nil),
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: ""},
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}

			if err := m.UpdateRepository(); (err != nil) != tt.wantErr {
				t.Errorf("AddRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_helm3Manager_runCommand(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name       string
		configFile string
		context    string
		want       string
		wantErr    bool
	}{
		{"No context, no config", "", "", "No context, no config", false},
		{"context, no config", "", "context", "context, no config", false},
		{"No context, config", ".kubeconfig", "", "No context, config", false},
		{"context, config", ".kubeconfig", "context", "context, config", false},
		{"invalid context", "", "invalid", "invalid context", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := fake.NewContextHandler(nil, nil)
			m := &helm3Manager{
				context: handler,
			}
			handler.ConfigFile = tt.configFile
			handler.ContextName = tt.context

			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: []string{"version"}, ResponseStatus: 0, Stdout: "No context, no config"},
				{Command: "helm", Args: []string{"version", "--kube-context", "context"}, ResponseStatus: 0, Stdout: "context, no config"},
				{Command: "helm", Args: []string{"version", "--kubeconfig", ".kubeconfig"}, ResponseStatus: 0, Stdout: "No context, config"},
				{Command: "helm", Args: []string{"version", "--kube-context", "context", "--kubeconfig", ".kubeconfig"}, ResponseStatus: 0, Stdout: "context, config"},
				{Command: "helm", Args: []string{"version", "--kube-context", "invalid"}, ResponseStatus: 1, Stdout: "invalid context"},
			}

			got, err := m.runCommand("version")
			if (err != nil) != tt.wantErr {
				t.Errorf("runCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_helm3Manager_ensureNamespaceExists(t *testing.T) {
	tests := []struct {
		name      string
		clientset *k8sFake.Clientset
		namespace string
		wantErr   bool
	}{
		{"exists", k8sFake.NewSimpleClientset(&v1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "ns"},
		}), "ns", false},
		{"notExists", k8sFake.NewSimpleClientset(), "ns", false},
		{"no clientset", nil, "ns", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &helm3Manager{
				context: fake.NewContextHandler(tt.clientset, nil),
			}
			if err := h.ensureNamespaceExists(tt.namespace); (err != nil) != tt.wantErr {
				t.Errorf("ensureNamespaceExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
