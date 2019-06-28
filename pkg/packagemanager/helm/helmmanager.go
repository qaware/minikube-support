package helm

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/sh"
	"github.com/chr-fritz/minikube-support/pkg/utils"
	"github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	"os"
)

type Manager interface {
	Init() error
	Install(chart string, release string, namespace string, values map[string]interface{})
	Uninstall(release string, purge bool)
}

type defaultManager struct {
}

func NewHelmManager() Manager {
	return &defaultManager{}
}

func (m *defaultManager) Init() error {
	e := m.checkTiller()
	if e == nil {
		return nil
	}
	return m.initTiller()
}

func (m *defaultManager) Install(chart string, release string, namespace string, values map[string]interface{}) {
	var args = []string{
		"--install", "--force",
		"--namespace", namespace,
		"--name", release,
		chart,
	}
	flatValues, e := utils.Flatten(values)
	if e != nil {
		logrus.Warnf("Can not flatten values map; Abort to install chart: %s", e)
		return
	}

	for k, v := range flatValues {
		val := shellquote.Join(fmt.Sprint(v))
		args = append(args, "--set", fmt.Sprintf("%s='%s'", k, val))
	}

	response, e := m.runCommand("upgrade", args...)
	if e != nil {
		logrus.Errorf("Can not install (%s) helm chart %s into namespace %s:\n%s", e, chart, namespace, response)
		return
	}

	logrus.Infof("Install of helm chart %s as %s/%s was successful.", chart, namespace, release)
	logrus.Debug(response)
}

func (m *defaultManager) Uninstall(release string, purge bool) {
	var e error
	var response string

	if purge {
		response, e = m.runCommand("delete", "--purge", release)
	} else {
		response, e = m.runCommand("delete", release)
	}

	if e != nil {
		logrus.Errorf("Can not delete helm release %s: %s\n%s", release, e, response)
		return
	}
	logrus.Infof("Helm release %s successfully deleted.", release)
	logrus.Debug(response)
}

func (m *defaultManager) runCommand(command string, args ...string) (string, error) {
	cmd := sh.ExecCommand("helm", append([]string{command}, args...)...)
	cmd.Env = append(cmd.Env, os.Environ()...)

	bytes, e := cmd.CombinedOutput()
	output := string(bytes)
	if e != nil {
		return output, fmt.Errorf("run helm %s was not successful: %s", command, e)
	}
	return output, nil
}

func (m *defaultManager) checkTiller() error {
	output, e := m.runCommand("version", "-s")

	if output == "Error: could not find a ready tiller pod" || e != nil {
		return fmt.Errorf("error: helm is not initialized")
	}
	return nil
}

func (m *defaultManager) initTiller() error {
	output, e := m.runCommand("init", "--wait")

	if e != nil {
		return fmt.Errorf("can not initialize tiller: %s", output)
	}
	return nil
}
