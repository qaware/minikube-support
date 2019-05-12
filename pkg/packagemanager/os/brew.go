// +build darwin

package os

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/packagemanager"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	"os"
)

// Implementation for the brew package manager.
type brewPackageManager struct{}

func newBrewPackageManager() packagemanager.PackageManager {
	return &brewPackageManager{}
}

func init() {
	packagemanager.SetOsPackageManager(newBrewPackageManager())
}

func (b *brewPackageManager) Install(pkg string) error {
	logrus.Infof("uses brew to install %s", pkg)
	return runBrewCommand("install", pkg)
}
func (b *brewPackageManager) Update(pkg string) error {
	logrus.Infof("uses brew to update %s", pkg)
	return runBrewCommand("update", pkg)
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
		return fmt.Errorf("run brew %s was not successful: %s", command, e)
	}
	return nil
}
