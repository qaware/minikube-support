// +build darwin

package os

import (
	"os"
	"testing"

	"github.com/qaware/minikube-support/pkg/testutils"
)

func Test_runBrewCommand(t *testing.T) {
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
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()

			testutils.MockWithStdOut("mkcert\nnss\n", 0, "brew", "list")
			testutils.MockWithoutResponse(1, "brew", "invalid")

			if err := runBrewCommand(tt.command); (err != nil) != tt.wantErr {
				t.Errorf("runBrewCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_brewPackageManager_Install(t *testing.T) {
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
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()

			testutils.MockWithoutResponse(0, "brew", "install", "ok")
			testutils.MockWithoutResponse(1, "brew", "install", "nok")

			b := &brewPackageManager{}
			if err := b.Install(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_brewPackageManager_Update(t *testing.T) {
	tests := []struct {
		name    string
		pkg     string
		wantErr bool
	}{
		{"ok", "ok", false},
		{"installed", "installed", false},
		{"not ok", "nok", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()

			testutils.MockWithoutResponse(0, "brew", "upgrade", "ok")
			testutils.MockWithoutResponse(1, "brew", "upgrade", "installed")
			testutils.MockWithoutResponse(2, "brew", "upgrade", "nok")

			b := &brewPackageManager{}
			if err := b.Update(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_brewPackageManager_Uninstall(t *testing.T) {
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
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()

			testutils.MockWithoutResponse(0, "brew", "uninstall", "ok")
			testutils.MockWithoutResponse(1, "brew", "uninstall", "nok")

			b := &brewPackageManager{}
			if err := b.Uninstall(tt.pkg); (err != nil) != tt.wantErr {
				t.Errorf("brewPackageManager.Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_brewPackageManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name    string
		pkg     string
		want    bool
		wantErr bool
	}{
		{"installed", "mkcert", true, false},
		{"not-installed", "invalid", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()
			testutils.MockWithStdOut("mkcert\nnss\n", 0, "brew", "list", "--formula")

			b := &brewPackageManager{}
			got, err := b.IsInstalled(tt.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsInstalled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsInstalled() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_brewPackageManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name              string
		cmdResponseStatus int
		want              bool
	}{
		{"installed", 0, true},
		{"not installed", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.StartCommandLineTest()
			defer testutils.StopCommandLineTest()
			b := &brewPackageManager{}
			testutils.MockWithoutResponse(tt.cmdResponseStatus, "which", "brew")

			if got := b.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	testutils.StandardHelperProcess(t)
}
