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
			oldHome := os.Getenv("HOME")
			defer func() { _ = os.Setenv("HOME", oldHome) }()
			if tt.homeDir != "" {
				_ = os.Setenv("HOME", tt.homeDir)
			}

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
