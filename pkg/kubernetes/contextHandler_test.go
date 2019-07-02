package kubernetes

import (
	"os"
	"testing"
)

func Test_contextHandler_GetClientSet(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		homeDir    string
		wantConfig bool
		wantErr    bool
	}{
		{"specified", "valid-config_test.yaml", "", true, false},
		{"use home", "", "./test-home/", true, false},
		{"not in home", "", "./invalid-home/", false, true},
		{"specified but not found", "not-existing.yaml", "", false, true},
		{"specified but invalid", "invalid-config_test.yaml", "", false, true},
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

			if tt.homeDir != "" {
				_ = os.Setenv("HOME", tt.homeDir)
			}
			_ = os.Unsetenv("KUBERNETES_SERVICE_HOST")
			_ = os.Unsetenv("KUBERNETES_SERVICE_PORT")

			context := ""
			h := NewContextHandler(&tt.configFile, &context)
			got, err := h.GetClientSet()
			if (err != nil) != tt.wantErr {
				t.Errorf("contextHandler.GetClientSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (got != nil) != tt.wantConfig {
				t.Errorf("contextHandler.GetClientSet() = %v, want config=%v", got, tt.wantConfig)
			}
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
