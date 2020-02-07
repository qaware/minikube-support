package mkcert

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/qaware/minikube-support/pkg/packagemanager"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
)

func Test_mkCertInstaller_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	packagemanager.SetOsPackageManager(&testPkgMgr{})
	i := &mkCertInstaller{}
	i.Install()
}

func Test_mkCertInstaller_Update(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	packagemanager.SetOsPackageManager(&testPkgMgr{})
	i := &mkCertInstaller{}
	i.Update()
}

func Test_mkCertInstaller_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	packagemanager.SetOsPackageManager(&testPkgMgr{})
	tests := []struct {
		name  string
		purge bool
	}{
		{"no purge", false},
		{"purge", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &mkCertInstaller{}
			i.Uninstall(tt.purge)
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
	case "mkcert":
		cmd, _ := args[0], args[1:]
		switch cmd {
		case "-install":
		case "-uninstall":
			os.Exit(0)
		default:
			os.Exit(1)
		}
	}
}

type testPkgMgr struct{}

func (tt *testPkgMgr) String() string {
	return "testPkgMgr"
}

func (tt *testPkgMgr) Install(pkg string) error {
	if pkg != "mkcert" && pkg != "nss" {
		return fmt.Errorf("pkg is not mkcert and not nss")
	}
	return nil
}

func (tt *testPkgMgr) Update(pkg string) error {
	if pkg != "mkcert" && pkg != "nss" {
		return fmt.Errorf("pkg is not mkcert and not nss")
	}
	return nil
}

func (tt *testPkgMgr) Uninstall(pkg string) error {
	if pkg != "mkcert" && pkg != "nss" {
		return fmt.Errorf("pkg is not mkcert and not nss")
	}
	return nil
}

func (tt *testPkgMgr) IsInstalled(pkg string) (bool, error) {
	if pkg != "mkcert" && pkg != "nss" {
		return true, nil
	}
	return false, nil
}
