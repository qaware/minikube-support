package helm

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

type Manager interface {
	Install(chart string, release string, namespace string, values map[string]string)
	Uninstall(release string, purge bool)
}

type defaultManager struct {
}

func NewHelmManager() Manager {
	return &defaultManager{}
}

func (m *defaultManager) Install(chart string, release string, namespace string, values map[string]string) {
	var args = []string{
		"--install", "--force",
		"--namespace", namespace,
		"--name", release,
		chart,
	}
	for k, v := range values {
		args = append(args, "--set", fmt.Sprintf("'%s=%s'", k, v))
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
	cmd := exec.Command("helm", append([]string{command}, args...)...)
	cmd.Env = os.Environ()

	bytes, e := cmd.CombinedOutput()
	output := string(bytes)
	if e != nil {
		return output, fmt.Errorf("run helm %s was not successful: %s", command, e)
	}
	return output, nil
}
