package helm

import (
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func TestNewHelmManager(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name    string
		version string
		status  int
		want    Manager
		wantErr bool
	}{
		{"helm2", "Client: v2.15.2+g8dce272", 0, &helm2Manager{}, false},
		{"helm3", "v3.0.0+ge29ce2a", 0, &helm3Manager{}, false},
		{"no helm installed", "", -1, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.TestProcessResponses = []testutils.TestProcessResponse{{
				Command:        "helm",
				Args:           []string{"version", "-c", "--short"},
				ResponseStatus: tt.status,
				Stdout:         tt.version,
			}}

			got, err := NewHelmManager(fake.NewContextHandler(nil, nil))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHelmManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.IsType(t, tt.want, got)
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}
