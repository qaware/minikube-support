package mkcert

import (
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/chr-fritz/minikube-support/pkg/packagemanager"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/sirupsen/logrus"
	"os"
)

type mkCertInstaller struct {
}

func CreateMkcertInstallerPlugin() apis.InstallablePlugin {
	return &mkCertInstaller{}
}

func (*mkCertInstaller) String() string {
	return "mkcert"
}

func (i *mkCertInstaller) Install() {
	manager := packagemanager.GetPackageManager()
	e := manager.Install("mkcert")
	if e != nil {
		logrus.Errorf("can not install mkcert: %s", e)
		return
	}
	e = manager.Install("nss")
	if e != nil {
		logrus.Errorf("can not install nss: %s", e)
	}

	i.Update()
}

func (i *mkCertInstaller) Update() {
	command := sh.ExecCommand("mkcert", "-install")
	command.Env = append(command.Env, os.Environ()...)
	output, e := command.CombinedOutput()
	if e != nil {
		logrus.Errorf("Can not install / update the current Root CA. Error: %s\nOutput: %s", e, string(output))
		return
	}
	logrus.Infof("Root CA successfully installed in browsers.\n%s", string(output))
}

func (i *mkCertInstaller) Uninstall(purge bool) {
	command := sh.ExecCommand("mkcert", "-uninstall")
	command.Env = append(command.Env, os.Environ()...)
	output, e := command.CombinedOutput()
	if e != nil {
		logrus.Errorf("Can not uninstall the current Root CA. Error: %s\nOutput: %s", e, string(output))
		return
	}
	logrus.Infof("Root CA successfully removed from browsers.\n%s", string(output))

	if purge {
		manager := packagemanager.GetPackageManager()
		e := manager.Uninstall("mkcert")
		if e != nil {
			logrus.Errorf("can not install mkcert: %s", e)
			return
		}
		e = manager.Uninstall("nss")
		if e != nil {
			logrus.Errorf("can not install nss: %s", e)
			return
		}
	}
}
