package mkcert

import (
	"os"
	"os/exec"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/qaware/minikube-support/pkg/packagemanager"
	"github.com/qaware/minikube-support/pkg/packagemanager/fake"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/testutils"
)

func Test_mkCertInstaller_Install(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		pkgMgrMinikube bool
		pkgMgrMkcert   bool
		wantInstall    bool
		wantUpdate     bool
	}{
		{"do-nothing->mks installed by pkg mgr", true, true, false, false},
		{"install", false, false, true, false},
		{"update", false, true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := fake.NewMockPackageManager(ctrl)
			packagemanager.SetOsPackageManager(manager)
			manager.EXPECT().IsInstalled("minikube-support").Times(1).Return(tt.pkgMgrMinikube, nil)
			manager.EXPECT().IsInstalled("mkcert").AnyTimes().Return(tt.pkgMgrMkcert, nil)
			manager.EXPECT().IsInstalled("nss").AnyTimes().Return(tt.pkgMgrMkcert, nil)

			if tt.wantInstall {
				manager.EXPECT().Install("nss").Times(1)
				manager.EXPECT().Install("mkcert").Times(1)
			}
			if tt.wantUpdate {
				manager.EXPECT().Update("nss").Times(1)
				manager.EXPECT().Update("mkcert").Times(1)
			}

			i := &mkCertInstaller{}
			i.Install()
		})
	}
}

func Test_mkCertInstaller_Update(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := fake.NewMockPackageManager(ctrl)
	packagemanager.SetOsPackageManager(manager)

	i := &mkCertInstaller{}
	i.Update()
}

func Test_mkCertInstaller_Uninstall(t *testing.T) {
	sh.ExecCommand = testutils.FakeExecCommand
	defer func() { sh.ExecCommand = exec.Command }()
	tests := []struct {
		name           string
		pkgMgrMinikube bool
		purge          bool
		wantUninstall  bool
	}{
		{"no purge", true, false, false},
		{"purge,managed-mks", true, true, false},
		{"purge", false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			manager := fake.NewMockPackageManager(ctrl)
			packagemanager.SetOsPackageManager(manager)
			manager.EXPECT().IsInstalled("minikube-support").AnyTimes().Return(tt.pkgMgrMinikube, nil)
			manager.EXPECT().IsInstalled("mkcert").AnyTimes().Return(true, nil)

			if tt.wantUninstall {
				manager.EXPECT().Uninstall("nss").Times(1)
				manager.EXPECT().Uninstall("mkcert").Times(1)
			}

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
