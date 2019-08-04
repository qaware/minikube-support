package helm

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=helmmanager.go

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/qaware/minikube-support/pkg/kubernetes"
	"github.com/qaware/minikube-support/pkg/sh"
	"github.com/qaware/minikube-support/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

type Manager interface {
	Init() error
	AddRepository(name string, url string) error
	UpdateRepository() error
	Install(chart string, release string, namespace string, values map[string]interface{}, wait bool)
	Uninstall(release string, purge bool)
}

type defaultManager struct {
	context     kubernetes.ContextHandler
	initialized bool
	mutex       sync.Mutex
}

func NewHelmManager(context kubernetes.ContextHandler) Manager {
	return &defaultManager{
		mutex:   sync.Mutex{},
		context: context,
	}
}

func (m *defaultManager) Init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.initialized {
		return nil
	}
	e := m.checkTiller()
	if e == nil {
		m.initialized = true
		return nil
	}

	if e = m.initTiller(); e != nil {
		return e
	}
	m.initialized = true
	return nil
}

func (m *defaultManager) Install(chart string, release string, namespace string, values map[string]interface{}, wait bool) {
	if !m.initialized {
		if e := m.Init(); e != nil {
			logrus.Errorf("Can not install helm chart: %s", e)
			return
		}
	}

	var args = []string{
		"--install", "--force",
		"--namespace", namespace,
		release, chart,
	}
	if wait {
		args = append(args, "--wait")
	}

	flatValues, e := utils.Flatten(values)
	if e != nil {
		logrus.Warnf("Can not flatten values map; Abort to install chart: %s", e)
		return
	}

	for k, v := range flatValues {
		val := shellquote.Join(fmt.Sprintf("%s=%s", k, v))
		args = append(args, "--set", val)
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
	if !m.initialized {
		if e := m.Init(); e != nil {
			logrus.Errorf("Can not uninstall helm chart: %s", e)
			return
		}
	}

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

func (m *defaultManager) AddRepository(name string, url string) error {
	if !m.initialized {
		if e := m.Init(); e != nil {
			return e
		}
	}

	_, e := m.runCommand("repo", "add", name, url)

	if e != nil {
		return e
	}
	return nil
}

func (m *defaultManager) UpdateRepository() error {
	if !m.initialized {
		if e := m.Init(); e != nil {
			return e
		}
	}
	_, e := m.runCommand("repo", "update")

	if e != nil {
		return e
	}
	return nil
}

func (m *defaultManager) runCommand(command string, args ...string) (string, error) {
	prefix := append([]string{command})
	if m.context.GetContextName() != "" {
		prefix = append(prefix, "--kube-context", m.context.GetContextName())
	}
	if m.context.GetConfigFile() != "" {
		prefix = append(prefix, "--kubeconfig", m.context.GetConfigFile())
	}

	cmd := sh.ExecCommand("helm", append(prefix, args...)...)
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
