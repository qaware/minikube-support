// +build darwin

package os

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/chr-fritz/minikube-support/pkg/testutils"
	"os"
	"os/exec"
	"testing"
)

func Test_runBrewCommand(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"list", "list", false},
		{"invalid command", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runBrewCommand(tt.command); (err != nil) != tt.wantErr {
				t.Errorf("runBrewCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_brewPackageManager_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name    string
		pkg     string
		wantErr bool
	}{
		{"ok", "ok", false},
		{"not ok", "nok", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &brewPackageManager{}
			if err := b.Install(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_brewPackageManager_Update(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name    string
		pkg     string
		wantErr bool
	}{
		{"ok", "ok", false},
		{"not ok", "nok", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &brewPackageManager{}
			if err := b.Update(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_brewPackageManager_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()

	tests := []struct {
		name    string
		pkg     string
		wantErr bool
	}{
		{"ok", "ok", false},
		{"not ok", "nok", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &brewPackageManager{}
			if err := b.Uninstall(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	cmd, args := testutils.ExtractMockedCommandAndArgs()
	switch cmd {
	case "brew":
		cmd, args := args[0], args[1:]
		switch cmd {
		case "list":
			fmt.Print("list")
		case "invalid":
			os.Exit(1)
		case "install":
			fallthrough
		case "uninstall":
			fallthrough
		case "update":
			if args[0] != "ok" {
				os.Exit(1)
			}
		}
	}
}
