// +build darwin

package os

import (
	"github.com/pkg/errors"
	"github.com/qaware/minikube-support/pkg/packagemanager"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Implementation for the brew package manager.
type brewPackageManager struct{}

func newBrewPackageManager() packagemanager.PackageManager {
	return &brewPackageManager{}
}

func InitOsPackageManager() {
	packagemanager.SetOsPackageManager(newBrewPackageManager())
}

func (b *brewPackageManager) Install(pkg string) error {
	logrus.Infof("uses brew to install %s", pkg)
	return runBrewCommand("install", pkg)
}

func (b *brewPackageManager) IsInstalled(pkg string) (bool, error) {
	response, e := sh.RunCmd("brew", "list")
	if e != nil {
		return false, errors.Wrapf(e, "can not check if %s is installed with brew", pkg)
	}
	if strings.Contains(response, pkg) {
		return true, nil
	}
	return false, nil
}

func (b *brewPackageManager) Update(pkg string) error {
	logrus.Infof("uses brew to update %s", pkg)
	_, e := sh.RunCmd("brew", "upgrade", pkg)

	if sh.IsExitCode(e, 1) {
		// already installed.
		return nil
	} else {
		return errors.Wrap(e, "can not upgrade pkg")
	}
}
func (b *brewPackageManager) Uninstall(pkg string) error {
	logrus.Infof("uses brew to uninstall %s", pkg)
	return runBrewCommand("uninstall", pkg)
}
func (b *brewPackageManager) String() string {
	return "BrewPackageManager"
}

func (brewPackageManager) IsOsPackageManager() bool {
	return true
}

func runBrewCommand(command string, args ...string) error {
	cmd := sh.ExecCommand("brew", append([]string{command}, args...)...)
	cmd.Env = append(cmd.Env, os.Environ()...)

	e := cmd.Run()
	if e != nil {
		return errors.Wrapf(e, "run brew %s was not successful", command)
	}
	return nil
}
