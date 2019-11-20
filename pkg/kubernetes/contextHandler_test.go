package kubernetes

import (
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"os"
	"os/exec"
	"testing"
)

func Test_contextHandler_GetClientSet(t *testing.T) {
	tests := []struct {
		name            string
		configFile      string
		homeDir         string
		context         *string
		wantContextName string
		wantConfig      bool
		wantErr         bool
	}{
		{
			"specified",
			"valid-config_test.yaml",
			"",
			nil,
			"test",
			true,
			false,
		}, {
			"specified, different context",
			"valid-config_test.yaml",
			"",
			s("test1"),
			"test1",
			true,
			false,
		}, {
			"use home",
			"",
			"./test-home/",
			nil,
			"test",
			true,
			false,
		}, {
			"use home, different context",
			"",
			"./test-home/",
			s("test1"),
			"test1",
			true,
			false,
		}, {
			"not in home",
			"",
			"./invalid-home/",
			nil,
			"",
			false,
			true,
		}, {
			"specified but not found",
			"not-existing.yaml",
			"",
			nil,
			"",
			false,
			true,
		}, {
			"specified but invalid",
			"invalid-config_test.yaml",
			"",
			nil,
			"",
			false,
			true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Unset Kubernetes environmental variables (HOST/PORT) to have the same testing behaviour in and outside of the cluster
			//  and "HOME" environmental variable
			//	-> restore after testing completed
			oldHome := os.Getenv("HOME")
			oldK8sServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
			oldK8sServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")

			defer func() {
				_ = os.Setenv("HOME", oldHome)
				_ = os.Setenv("KUBERNETES_SERVICE_HOST", oldK8sServiceHost)
				_ = os.Setenv("KUBERNETES_SERVICE_PORT", oldK8sServicePort)
			}()

			if tt.homeDir != "" {
				_ = os.Setenv("HOME", tt.homeDir)
			}
			_ = os.Unsetenv("KUBERNETES_SERVICE_HOST")
			_ = os.Unsetenv("KUBERNETES_SERVICE_PORT")

			h := NewContextHandler(&tt.configFile, tt.context)
			got, err := h.GetClientSet()
			if (err != nil) != tt.wantErr {
				t.Errorf("contextHandler.GetClientSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (got != nil) != tt.wantConfig {
				t.Errorf("contextHandler.GetClientSet() = %v, want config=%v", got, tt.wantConfig)
			}
			assert.Equal(t, tt.wantContextName, h.GetContextName())
		})
	}
}

func Test_contextHandler_GetDynamicClient(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		wantConfig bool
		wantErr    bool
	}{
		{"specified", "valid-config_test.yaml", true, false},
		{"specified but invalid", "invalid-config_test.yaml", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Unset Kubernetes environmental variables (HOST/PORT) to have the same testing behaviour in and outside of the cluster
			//  and "HOME" environmental variable
			//	-> restore after testing completed
			oldHome := os.Getenv("HOME")
			oldK8sServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
			oldK8sServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")

			defer func() {
				_ = os.Setenv("HOME", oldHome)
				_ = os.Setenv("KUBERNETES_SERVICE_HOST", oldK8sServiceHost)
				_ = os.Setenv("KUBERNETES_SERVICE_PORT", oldK8sServicePort)
			}()

			_ = os.Unsetenv("KUBERNETES_SERVICE_HOST")
			_ = os.Unsetenv("KUBERNETES_SERVICE_PORT")

			context := ""
			h := NewContextHandler(&tt.configFile, &context)
			got, err := h.GetDynamicClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("contextHandler.GetDynamicClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (got != nil) != tt.wantConfig {
				t.Errorf("contextHandler.GetDynamicClient() = %v, want config=%v", got, tt.wantConfig)
			}
		})
	}
}

func Test_contextHandler_Kubectl(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		command        string
		configFile     string
		contextName    string
		expectedArgs   []string
		responseStatus int
		response       string
		want           string
		wantErr        bool
	}{
		{"ok", "apply", "", "", []string{"apply"}, 0, "ok", "ok", false},
		{"config file", "apply", "kubeconfig", "", []string{"apply", "--kubeconfig", "kubeconfig"}, 0, "ok", "ok", false},
		{"context", "apply", "", "context", []string{"apply", "--context", "context"}, 0, "ok", "ok", false},
		{"error", "apply", "", "", []string{"apply"}, -1, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &contextHandler{
				configFile:            &tt.configFile,
				predefinedContextName: &tt.contextName,
			}
			testutils.TestProcessResponses = []testutils.TestProcessResponse{
				{Command: "kubectl", Args: tt.expectedArgs, ResponseStatus: tt.responseStatus, Stdout: tt.response},
			}

			got, err := h.Kubectl(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("contextHandler.Kubectl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("contextHandler.Kubectl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contextHandler_IsMinikube(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		restConfig     *rest.Config
		ip             string
		responseStatus int
		want           bool
		wantErr        bool
	}{
		{"yes and initialized", &rest.Config{Host: "192.168.64.2"}, "192.168.64.2\n", 0, true, false},
		{"no and initialized", &rest.Config{}, "192.168.64.3", 0, false, false},
		{"error and initialized", &rest.Config{}, "", 1, false, true},
		{"uninitialized and no minikube vm", nil, "", 66, false, false},
		{"no and uninitialized", nil, "192.168.64.3", 0, false, false},
		{"yes and uninitialized", nil, "localhost\r\n", 0, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldHome := os.Getenv("HOME")
			defer func() {
				_ = os.Setenv("HOME", oldHome)
			}()
			_ = os.Setenv("HOME", "test-home")

			testutils.TestProcessResponses = []testutils.TestProcessResponse{{
				Command:        "minikube",
				Args:           []string{"ip"},
				ResponseStatus: tt.responseStatus,
				Stdout:         tt.ip,
				//Stderr:         "error",
			}}

			h := NewContextHandler(s(""), nil).(*contextHandler)
			h.restConfig = tt.restConfig
			got, err := h.IsMinikube()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsMinikube() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}

func s(str string) *string {
	return &str
}
