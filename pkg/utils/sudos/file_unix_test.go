// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package sudos

import (
	"os/exec"
	"testing"

	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
)

func TestChown(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name      string
		path      string
		uid       int
		gid       int
		recursive bool
		responses []testutils.TestProcessResponse
		wantErr   bool
	}{
		{
			"ok",
			"./path",
			1000,
			1000,
			false,
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"chown", "1000:1000", "./path"}, ResponseStatus: 0}},
			false,
		}, {
			"ok-recursive",
			"./path",
			1000,
			1000,
			true,
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"chown", "-R", "1000:1000", "./path"}, ResponseStatus: 0}},
			false,
		}, {
			"nok",
			"./path",
			1000,
			1000,
			false,
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"chown", "1000:1000", "./path"}, ResponseStatus: -1}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.SetTestProcessResponses(tt.responses)
			testutils.AddTestProcessResponse(
				testutils.TestProcessResponse{Command: "which", Args: []string{"sudo"}, ResponseStatus: 0})

			if err := Chown(tt.path, tt.uid, tt.gid, tt.recursive); (err != nil) != tt.wantErr {
				t.Errorf("Chown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMkdirAll(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name      string
		path      string
		mod       int
		responses []testutils.TestProcessResponse
		wantErr   bool
	}{
		{
			"ok",
			"./path",
			0755,
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"mkdir", "-p", "-m", "755", "./path"}, ResponseStatus: 0}},
			false,
		}, {
			"nok",
			"./path",
			0755,
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"mkdir", "-p", "-m", "755", "./path"}, ResponseStatus: -1}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.SetTestProcessResponses(tt.responses)
			testutils.AddTestProcessResponse(
				testutils.TestProcessResponse{Command: "which", Args: []string{"sudo"}, ResponseStatus: 0})

			if err := MkdirAll(tt.path, tt.mod); (err != nil) != tt.wantErr {
				t.Errorf("MkdirAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name      string
		path      string
		responses []testutils.TestProcessResponse
		wantErr   bool
	}{
		{
			"ok",
			"./path",
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"rm", "-R", "./path"}, ResponseStatus: 0}},
			false,
		},
		{
			"ok",
			"./path",
			[]testutils.TestProcessResponse{{Command: "sudo", Args: []string{"rm", "-R", "./path"}, ResponseStatus: -1}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.SetTestProcessResponses(tt.responses)
			testutils.AddTestProcessResponse(
				testutils.TestProcessResponse{Command: "which", Args: []string{"sudo"}, ResponseStatus: 0})

			if err := RemoveAll(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	testutils.StandardHelperProcess(t)
}
