package helm

import (
	"os/exec"
	"testing"

	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var global = test.NewGlobal()

func Test_helm2Manager_Init(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name          string
		initialized   bool
		versionStatus int
		versionMsg    string
		initStatus    int
		wantErr       bool
	}{
		{"installed", false, 0, "", 0, false},
		{"initialized", true, 0, "", 0, false},
		{"notInstalled", false, -1, "Error: could not find a ready tiller pod", 0, false},
		{"installFailure", false, -1, "Error: could not find a ready tiller pod", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm2Manager{
				context:     fake.NewContextHandler(nil, nil),
				initialized: tt.initialized,
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: tt.versionStatus, Stdout: tt.versionMsg},
				{Command: "helm", Args: []string{"init", "--wait"}, ResponseStatus: tt.initStatus},
			}
			if err := m.Init(); (err != nil) != tt.wantErr {
				t.Errorf("helm2Manager.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_helm2Manager_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		chart          string
		release        string
		namespace      string
		wait           bool
		values         map[string]interface{}
		initialized    bool
		expectedArgs   [][]string
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
			true,
			[][]string{{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test"}},
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
			true,
			[][]string{{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test", "--wait"}},
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"success uninitialized",
			"dummy/test",
			"test",
			"test",
			false,
			map[string]interface{}{},
			false,
			[][]string{{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test"}},
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
			true,
			[][]string{
				{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test", "--set", "v1\\[0].h=2", "--set", "v1\\[0].b=def"},
				{"upgrade", "--install", "--force", "--namespace", "test", "test", "dummy/test", "--set", "v1\\[0].b=def", "--set", "v1\\[0].h=2"},
			},
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
			true,
			[][]string{{"upgrade", "--install", "--force", "--namespace", "test", "", ""}},
			"no release and name given",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm2Manager{
				context:     fake.NewContextHandler(nil, nil),
				initialized: tt.initialized,
			}

			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}
			for _, args := range tt.expectedArgs {
				testutils.TestProcessResponses = append(testutils.TestProcessResponses,
					testutils.TestProcessResponse{Command: "helm", Args: args, ResponseStatus: tt.responseStatus, Stdout: tt.response})
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

func Test_helm2Manager_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		release        string
		purge          bool
		initialized    bool
		expectedArgs   []string
		response       string
		responseStatus int
		lastEntryLevel logrus.Level
	}{
		{
			"success no purge",
			"test",
			false,
			true,
			[]string{"delete", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"success no purge uninitialized",
			"test",
			false,
			false,
			[]string{"delete", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"success purge",
			"test",
			true,
			true,
			[]string{"delete", "--purge", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"not found",
			"test",
			false,
			true,
			[]string{"delete", "test"},
			"not found",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm2Manager{
				context:     fake.NewContextHandler(nil, nil),
				initialized: tt.initialized,
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: 0, Stdout: ""},
			}
			m.Uninstall(tt.release, "", tt.purge)
			lastEntry := global.LastEntry()
			if lastEntry.Level != tt.lastEntryLevel {
				t.Errorf("Expected log level of last entry %s but was %s", tt.lastEntryLevel, lastEntry.Level)
			}
		})
	}
}

func Test_helm2Manager_AddRepository(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		initialized    bool
		repoName       string
		url            string
		expectedArgs   []string
		responseStatus int
		wantErr        bool
	}{
		{"ok uninitialized", false, "dummy", "http://localhost", []string{"repo", "add", "dummy", "http://localhost"}, 0, false},
		{"ok", true, "dummy", "http://localhost", []string{"repo", "add", "dummy", "http://localhost"}, 0, false},
		{"failed", true, "dummy", "http://localhost", []string{"repo", "add", "dummy", "http://localhost"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm2Manager{
				context:     fake.NewContextHandler(nil, nil),
				initialized: tt.initialized,
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

func Test_helm2Manager_UpdateRepository(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		initialized    bool
		expectedArgs   []string
		responseStatus int
		wantErr        bool
	}{
		{"ok uninitialized", false, []string{"repo", "update"}, 0, false},
		{"ok", true, []string{"repo", "update"}, 0, false},
		{"failed", true, []string{"repo", "update"}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &helm2Manager{
				context:     fake.NewContextHandler(nil, nil),
				initialized: tt.initialized,
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

func Test_helm2Manager_runCommand(t *testing.T) {
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
			m := &helm2Manager{
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
