package helm

import (
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/chr-fritz/minikube-support/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"os/exec"
	"testing"
)

var global = test.NewGlobal()

func Test_defaultManager_Init(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name          string
		versionStatus int
		versionMsg    string
		initStatus    int
		wantErr       bool
	}{
		{"installed", 0, "", 0, false},
		{"notInstalled", -1, "Error: could not find a ready tiller pod", 0, false},
		{"installFailure", -1, "Error: could not find a ready tiller pod", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &defaultManager{}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: []string{"version", "-s"}, ResponseStatus: tt.versionStatus, Stdout: tt.versionMsg},
				{Command: "helm", Args: []string{"init"}, ResponseStatus: tt.initStatus},
			}
			if err := m.Init(); (err != nil) != tt.wantErr {
				t.Errorf("defaultManager.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultManager_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		chart          string
		release        string
		namespace      string
		values         map[string]interface{}
		expectedArgs   []string
		response       string
		responseStatus int
		lastEntryLevel logrus.Level
	}{
		{
			"success",
			"dummy/test",
			"test",
			"test",
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "--name", "test", "dummy/test"},
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"success with values",
			"dummy/test",
			"test",
			"test",
			map[string]interface{}{"v1": []map[string]interface{}{{"h": 2, "b": "def"}}},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "--name", "test", "dummy/test", "--set", "v1[0].h='2'", "--set", "v1[0].b='def'"},
			"ok installed",
			0,
			logrus.InfoLevel,
		}, {
			"missing name and chart",
			"",
			"",
			"test",
			map[string]interface{}{},
			[]string{"upgrade", "--install", "--force", "--namespace", "test", "--name", "", ""},
			"no release and name given",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &defaultManager{}

			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
			}

			m.Install(tt.chart, tt.release, tt.namespace, tt.values)

			lastEntry := global.LastEntry()
			if lastEntry.Level != tt.lastEntryLevel {
				t.Errorf("Expected log level of last entry %s but was %s", tt.lastEntryLevel, lastEntry.Level)
			}
		})
	}
}

func Test_defaultManager_Uninstall(t *testing.T) {
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
			[]string{"delete", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"success purge",
			"test",
			true,
			[]string{"delete", "--purge", "test"},
			"ok removed",
			0,
			logrus.InfoLevel,
		}, {
			"not found",
			"test",
			false,
			[]string{"delete", "test"},
			"not found",
			1,
			logrus.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &defaultManager{}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "helm", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
			}
			m.Uninstall(tt.release, tt.purge)
			lastEntry := global.LastEntry()
			if lastEntry.Level != tt.lastEntryLevel {
				t.Errorf("Expected log level of last entry %s but was %s", tt.lastEntryLevel, lastEntry.Level)
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}
